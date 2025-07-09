package grpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/dmytrogajewski/hercules"
	"github.com/dmytrogajewski/hercules/internal/config"
	"github.com/dmytrogajewski/hercules/internal/core"
	"github.com/dmytrogajewski/hercules/internal/pb"
)

// Server wraps the gRPC server and job management
type Server struct {
	pb.UnimplementedHerculesServiceServer
	config    *config.Config
	jobs      map[string]*Job
	jobsMutex sync.RWMutex
	grpcSrv   *grpc.Server
	addr      string
	logger    core.Logger
}

type Job struct {
	ID        string
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Error     string
	Request   *pb.SubmitAnalysisRequest
	Results   map[string]*pb.AnalysisResult
}

func NewServer(cfg *config.Config, addr string, logger core.Logger) *Server {
	return &Server{
		config:  cfg,
		jobs:    make(map[string]*Job),
		grpcSrv: grpc.NewServer(),
		addr:    addr,
		logger:  logger,
	}
}

func (s *Server) Start() error {
	pb.RegisterHerculesServiceServer(s.grpcSrv, s)
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Errorf("Failed to listen: %v", err)
		return err
	}
	s.logger.Infof("Starting Hercules gRPC server on %s", s.addr)
	return s.grpcSrv.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcSrv.GracefulStop()
}

// --- gRPC API Implementation ---

func (s *Server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:    "healthy",
		Timestamp: timestamppb.Now(),
		Version:   int32(hercules.BinaryVersion),
		Hash:      hercules.BinaryGitHash,
		Config: map[string]string{
			"grpc": s.addr,
		},
	}, nil
}

func (s *Server) ListAnalyses(ctx context.Context, req *pb.ListAnalysesRequest) (*pb.ListAnalysesResponse, error) {
	return &pb.ListAnalysesResponse{
		Analyses: []*pb.AnalysisType{
			{Name: "burndown", Description: "Line burndown statistics for project, files and developers"},
			{Name: "couples", Description: "Coupling statistics for files and developers"},
			{Name: "devs", Description: "Developer activity statistics"},
			{Name: "commits-stat", Description: "Statistics for each commit"},
			{Name: "file-history", Description: "File history analysis"},
			{Name: "imports-per-dev", Description: "Import usage per developer"},
			{Name: "shotness", Description: "Structural hotness analysis"},
		},
		Timestamp: timestamppb.Now(),
	}, nil
}

func (s *Server) SubmitAnalysis(ctx context.Context, req *pb.SubmitAnalysisRequest) (*pb.SubmitAnalysisResponse, error) {
	if req.Repository == "" {
		return nil, status.Error(codes.InvalidArgument, "Repository URL is required")
	}

	if len(req.Analyses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "At least one analysis type is required")
	}

	// Check concurrent analysis limit
	s.jobsMutex.RLock()
	if len(s.jobs) >= s.config.Analysis.MaxConcurrentAnalyses {
		s.jobsMutex.RUnlock()
		return nil, status.Error(codes.ResourceExhausted, "Too many concurrent analyses")
	}
	s.jobsMutex.RUnlock()

	// Create job ID
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	// Create job
	job := &Job{
		ID:        jobID,
		Status:    "running",
		StartTime: time.Now(),
		Request:   req,
		Results:   make(map[string]*pb.AnalysisResult),
	}

	// Add job to tracking
	s.jobsMutex.Lock()
	s.jobs[jobID] = job
	s.jobsMutex.Unlock()

	// Start analysis in a goroutine
	go s.runAnalysis(job)

	return &pb.SubmitAnalysisResponse{
		Status:    "accepted",
		Message:   "Analysis started",
		JobId:     jobID,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (s *Server) GetAnalysisStatus(ctx context.Context, req *pb.GetAnalysisStatusRequest) (*pb.GetAnalysisStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "Job ID is required")
	}

	s.jobsMutex.RLock()
	job, exists := s.jobs[req.JobId]
	s.jobsMutex.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "Job not found")
	}

	response := &pb.GetAnalysisStatusResponse{
		Status:    job.Status,
		Message:   s.getStatusMessage(job.Status),
		JobId:     job.ID,
		StartTime: timestamppb.New(job.StartTime),
		Timestamp: timestamppb.Now(),
	}

	if job.EndTime != nil {
		response.EndTime = timestamppb.New(*job.EndTime)
	}

	if job.Status == "completed" {
		response.Results = job.Results
	}

	return response, nil
}

func (s *Server) StreamAnalysisProgress(req *pb.StreamAnalysisProgressRequest, stream pb.HerculesService_StreamAnalysisProgressServer) error {
	if req.JobId == "" {
		return status.Error(codes.InvalidArgument, "Job ID is required")
	}

	s.jobsMutex.RLock()
	job, exists := s.jobs[req.JobId]
	s.jobsMutex.RUnlock()

	if !exists {
		return status.Error(codes.NotFound, "Job not found")
	}

	// Send initial status
	err := stream.Send(&pb.StreamAnalysisProgressResponse{
		JobId:           job.ID,
		Status:          job.Status,
		Message:         s.getStatusMessage(job.Status),
		ProgressPercent: 0,
		Timestamp:       timestamppb.Now(),
	})
	if err != nil {
		return err
	}

	// Monitor job progress
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			s.jobsMutex.RLock()
			job, exists := s.jobs[req.JobId]
			s.jobsMutex.RUnlock()

			if !exists {
				return status.Error(codes.NotFound, "Job not found")
			}

			progress := s.calculateProgress(job)
			err := stream.Send(&pb.StreamAnalysisProgressResponse{
				JobId:           job.ID,
				Status:          job.Status,
				Message:         s.getStatusMessage(job.Status),
				ProgressPercent: int32(progress),
				Timestamp:       timestamppb.Now(),
			})
			if err != nil {
				return err
			}

			if job.Status == "completed" || job.Status == "failed" {
				return nil
			}
		}
	}
}

// Helper methods

func (s *Server) getStatusMessage(status string) string {
	switch status {
	case "running":
		return "Analysis in progress"
	case "completed":
		return "Analysis completed"
	case "failed":
		return "Analysis failed"
	default:
		return "Unknown status"
	}
}

func (s *Server) calculateProgress(job *Job) int {
	if job.Status == "completed" {
		return 100
	}
	if job.Status == "failed" {
		return 0
	}
	// Simple progress calculation based on time elapsed
	elapsed := time.Since(job.StartTime)
	if elapsed > 30*time.Second {
		return 50
	}
	return int(elapsed.Seconds() * 2) // Rough estimate
}

func (s *Server) runAnalysis(job *Job) {
	defer func() {
		now := time.Now()
		job.EndTime = &now
		if job.Status == "running" {
			job.Status = "completed"
		}
	}()

	// Convert gRPC request to internal format
	analyses := make([]string, len(job.Request.Analyses))
	copy(analyses, job.Request.Analyses)

	// Run the analysis using the same logic as HTTP server
	// This is a simplified version - in practice, you'd want to share the exact same analysis engine
	for _, analysisType := range analyses {
		result := &pb.AnalysisResult{
			Name:        analysisType,
			Description: s.getAnalysisDescription(analysisType),
			ResultType:  fmt.Sprintf("*leaves.%sResult", analysisType),
		}

		// Simulate analysis work
		time.Sleep(2 * time.Second)

		job.Results[analysisType] = result
	}

	job.Status = "completed"
}

func (s *Server) getAnalysisDescription(analysisType string) string {
	descriptions := map[string]string{
		"burndown":        "Line burndown statistics for project, files and developers",
		"couples":         "Coupling statistics for files and developers",
		"devs":            "Developer activity statistics",
		"commits-stat":    "Statistics for each commit",
		"file-history":    "File history analysis",
		"imports-per-dev": "Import usage per developer",
		"shotness":        "Structural hotness analysis",
	}
	return descriptions[analysisType]
}
