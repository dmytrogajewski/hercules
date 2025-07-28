package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dmytrogajewski/hercules/api/proto/pb"
	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/config"
	"github.com/dmytrogajewski/hercules/internal/pkg/leaves"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/internal/pkg/version"
	grpcserver "github.com/dmytrogajewski/hercules/internal/server/grpc"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

// AnalysisRequest represents a request to analyze a repository
type AnalysisRequest struct {
	Repository string            `json:"repository"`
	Analyses   []string          `json:"analyses"`
	Options    map[string]string `json:"options,omitempty"`
}

// AnalysisResponse represents the response from an analysis
type AnalysisResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Results   map[string]interface{} `json:"results,omitempty"`
	Metadata  *pb.Metadata           `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AnalysisJob represents a running analysis job
type AnalysisJob struct {
	ID        string
	Status    string
	StartTime time.Time
	EndTime   *time.Time
	Error     string
	Request   AnalysisRequest
}

// Server represents the Hercules HTTP server
type Server struct {
	config    *config.Config
	router    *mux.Router
	jobs      map[string]*AnalysisJob
	jobsMutex sync.RWMutex
	cache     core.CacheBackend
}

// NewServer creates a new Hercules server
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		config: cfg,
		router: mux.NewRouter(),
		jobs:   make(map[string]*AnalysisJob),
	}

	// Initialize cache backend
	if cfg.Cache.Enabled {
		cacheConfig := core.CacheConfig{
			Backend:    cfg.Cache.Backend,
			LocalPath:  cfg.Cache.Directory,
			S3Bucket:   cfg.Cache.S3Bucket,
			S3Region:   cfg.Cache.S3Region,
			S3Endpoint: cfg.Cache.S3Endpoint,
			S3Prefix:   cfg.Cache.S3Prefix,
			DefaultTTL: cfg.Cache.TTL,
		}

		cache, err := core.NewCacheBackend(cacheConfig)
		if err != nil {
			logger := core.GetLogger()
			logger.Warnf("Warning: Failed to initialize cache backend: %v", err)
		} else {
			s.cache = cache
			logger := core.GetLogger()
			logger.Infof("Cache backend initialized: %s", cfg.Cache.Backend)
		}
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/analyze", s.analyzeHandler).Methods("POST")
	api.HandleFunc("/analyses", s.listAnalysesHandler).Methods("GET")
	api.HandleFunc("/status/{id}", s.statusHandler).Methods("GET")

	// Static file serving for documentation
	s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))

	// Root redirect to docs
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   version.Binary,
		"hash":      version.BinaryGitHash,
		"config": map[string]interface{}{
			"server_port": s.config.Server.Port,
			"cache_dir":   s.config.Cache.Directory,
		},
	})
}

// listAnalysesHandler returns available analysis types
func (s *Server) listAnalysesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	analyses := []map[string]string{
		{"name": "burndown", "description": "Line burndown statistics for project, files and developers"},
		{"name": "couples", "description": "Coupling statistics for files and developers"},
		{"name": "devs", "description": "Developer activity statistics"},
		{"name": "commits-stat", "description": "Statistics for each commit"},
		{"name": "file-history", "description": "File history analysis"},
		{"name": "imports-per-dev", "description": "Import usage per developer"},
		{"name": "shotness", "description": "Structural hotness analysis"},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"analyses":  analyses,
		"timestamp": time.Now(),
	})
}

// analyzeHandler handles analysis requests
func (s *Server) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Repository == "" {
		http.Error(w, "Repository URL is required", http.StatusBadRequest)
		return
	}

	if len(req.Analyses) == 0 {
		http.Error(w, "At least one analysis type is required", http.StatusBadRequest)
		return
	}

	// Check concurrent analysis limit
	s.jobsMutex.RLock()
	if len(s.jobs) >= s.config.Analysis.MaxConcurrentAnalyses {
		s.jobsMutex.RUnlock()
		http.Error(w, "Too many concurrent analyses", http.StatusTooManyRequests)
		return
	}
	s.jobsMutex.RUnlock()

	// Create job ID
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	// Create job
	job := &AnalysisJob{
		ID:        jobID,
		Status:    "running",
		StartTime: time.Now(),
		Request:   req,
	}

	// Add job to tracking
	s.jobsMutex.Lock()
	s.jobs[jobID] = job
	s.jobsMutex.Unlock()

	// Start analysis in a goroutine
	go s.runAnalysis(job)

	response := AnalysisResponse{
		Status:    "accepted",
		Message:   "Analysis started",
		Timestamp: time.Now(),
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// statusHandler handles status requests
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	jobID := vars["id"]
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	s.jobsMutex.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMutex.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	response := AnalysisResponse{
		Status:    job.Status,
		Timestamp: time.Now(),
	}

	if job.Error != "" {
		response.Message = job.Error
	} else if job.Status == "completed" {
		response.Message = "Analysis completed"
	} else {
		response.Message = "Analysis in progress"
	}

	json.NewEncoder(w).Encode(response)
}

// runAnalysis performs the actual analysis
func (s *Server) runAnalysis(job *AnalysisJob) {
	defer func() {
		now := time.Now()
		job.EndTime = &now
		if job.Status == "running" {
			job.Status = "completed"
		}
	}()

	// Check cache first if available
	if s.cache != nil {
		cacheKey := core.GenerateCacheKey(job.Request.Repository, "main", "", strings.Join(job.Request.Analyses, ","))
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil {
			// Parse cached results
			var cachedResults map[string]interface{}
			if err := json.Unmarshal(cached, &cachedResults); err == nil {
				job.Status = "completed"
				logger := core.GetLogger()
				logger.Infof("Analysis completed from cache for %s", job.Request.Repository)
				return
			}
		}
	}

	// Create a temporary directory for this analysis
	tempDir, err := os.MkdirTemp(s.config.Cache.Directory, "hercules-*")
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to create temp directory: %v", err)
		logger := core.GetLogger()
		logger.Errorf("Failed to create temp directory: %v", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Load repository
	repository := loadRepository(job.Request.Repository, tempDir, true, "")

	// Create pipeline
	pipeline := core.NewPipeline(repository)

	// Configure pipeline based on request
	facts := make(map[string]interface{})

	// Set defaults from config
	facts[plumbing.ConfigTicksSinceStartTickSize] = s.config.Analysis.DefaultTickSize
	facts["Burndown.Granularity"] = s.config.Analysis.DefaultGranularity
	facts["Burndown.Sampling"] = s.config.Analysis.DefaultSampling

	// Parse options from request
	for key, value := range job.Request.Options {
		switch key {
		case "tick-size":
			if size, err := strconv.Atoi(value); err == nil {
				facts[plumbing.ConfigTicksSinceStartTickSize] = size
			}
		case "granularity":
			if gran, err := strconv.Atoi(value); err == nil {
				facts["Burndown.Granularity"] = gran
			}
		case "sampling":
			if samp, err := strconv.Atoi(value); err == nil {
				facts["Burndown.Sampling"] = samp
			}
		}
	}

	// Deploy requested analyses
	var deployed []core.LeafPipelineItem

	for _, analysis := range job.Request.Analyses {
		switch analysis {
		case "burndown":
			item := pipeline.DeployItem(&leaves.BurndownAnalysis{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "couples":
			item := pipeline.DeployItem(&leaves.CouplesAnalysis{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "devs":
			item := pipeline.DeployItem(&leaves.DevsAnalysis{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "commits-stat":
			item := pipeline.DeployItem(&leaves.CommitsAnalysis{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "file-history":
			item := pipeline.DeployItem(&leaves.FileHistoryAnalysis{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "imports-per-dev":
			item := pipeline.DeployItem(&leaves.ImportsPerDeveloper{}).(core.LeafPipelineItem)
			deployed = append(deployed, item)
		case "shotness":
			// Shotness analysis is disabled for now
			continue
		}
	}

	// Initialize and run pipeline
	if err := pipeline.Initialize(facts); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to initialize pipeline: %v", err)
		logger := core.GetLogger()
		logger.Errorf("Failed to initialize pipeline: %v", err)
		return
	}

	commits, err := pipeline.Commits(false)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to get commits: %v", err)
		logger := core.GetLogger()
		logger.Errorf("Failed to get commits: %v", err)
		return
	}

	results, err := pipeline.Run(commits)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to run pipeline: %v", err)
		logger := core.GetLogger()
		logger.Errorf("Failed to run pipeline: %v", err)
		return
	}

	// Convert results to JSON-serializable format
	jsonResults := make(map[string]interface{})
	for _, item := range deployed {
		result := results[item]
		if result != nil {
			jsonResults[item.Flag()] = map[string]interface{}{
				"name":        item.Name(),
				"description": item.Description(),
				"result_type": fmt.Sprintf("%T", result),
			}
		}
	}

	// Cache results if cache is available
	if s.cache != nil {
		cacheKey := core.GenerateCacheKey(job.Request.Repository, "main", "", strings.Join(job.Request.Analyses, ","))
		if cachedData, err := json.Marshal(jsonResults); err == nil {
			ttl := s.config.Cache.TTL
			if ttl == 0 {
				ttl = 24 * time.Hour // Default 24 hours
			}
			if err := s.cache.Set(context.Background(), cacheKey, cachedData, ttl); err != nil {
				logger := core.GetLogger()
				logger.Warnf("Warning: Failed to cache results: %v", err)
			} else {
				logger := core.GetLogger()
				logger.Infof("Results cached for %s", job.Request.Repository)
			}
		}
	}

	job.Status = "completed"
	logger := core.GetLogger()
	logger.Infof("Analysis completed for %s", job.Request.Repository)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	logger := core.GetLogger()
	logger.Infof("Starting Hercules server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	return server.ListenAndServe()
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Hercules as an HTTP and/or gRPC server",
	Long: `Start Hercules as an HTTP and/or gRPC server to provide analysis capabilities via REST API and gRPC.

The server provides endpoints for:
- HTTP: POST /api/v1/analyze, GET /api/v1/analyses, GET /health, GET /docs/
- gRPC: HerculesService with Health, ListAnalyses, SubmitAnalysis, GetAnalysisStatus, StreamAnalysisProgress

Use --config to specify a configuration file`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")

		// Load configuration
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			logger := core.GetLogger()
			logger.Errorf("Failed to load configuration: %v", err)
			os.Exit(1)
		}

		// Create cache directory if it doesn't exist
		if cfg.Cache.Directory != "" {
			if err := os.MkdirAll(cfg.Cache.Directory, 0755); err != nil {
				logger := core.GetLogger()
				logger.Errorf("Failed to create cache directory: %v", err)
				os.Exit(1)
			}
		}

		if cfg.Server.Enabled {
			logger := core.GetLogger()
			logger.Infof("Starting Hercules server...")
			startHTTPServer(cfg)
		}

		if cfg.GRPC.Enabled {
			logger := core.GetLogger()
			logger.Infof("Starting Hercules gRPC server...")
			startGRPCServer(cfg)
		}
	},
}

func startGRPCServer(cfg *config.Config) {
	logger := core.GetLogger()
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	grpcServer := grpcserver.NewServer(cfg, grpcAddr, logger)
	if err := grpcServer.Start(); err != nil {
		logger.Errorf("gRPC server failed: %v", err)
	}
}

func startHTTPServer(cfg *config.Config) {
	server := NewServer(cfg)
	if err := server.Start(); err != nil {
		logger := core.GetLogger()
		logger.Errorf("HTTP server failed: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("config", "c", "", "Path to configuration file")
}
