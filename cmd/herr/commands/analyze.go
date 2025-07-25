package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/cohesion"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/complexity"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/halstead"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/spf13/cobra"
)

// AnalyzeCommand holds the flags for the analyze command
type AnalyzeCommand struct {
	output       string
	format       string
	analyzerList []string
}

// NewAnalyzeCommand creates and configures the analyze command
func NewAnalyzeCommand() *cobra.Command {
	cmd := &AnalyzeCommand{}

	cobraCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze code complexity and other metrics",
		Long:  "Analyze code complexity and other metrics from UAST input",
		RunE:  cmd.Run,
	}

	// Add flags
	cobraCmd.Flags().StringVarP(&cmd.output, "output", "o", "", "Output file (default: stdout)")
	cobraCmd.Flags().StringVarP(&cmd.format, "format", "f", "text", "Output format: text or json")
	cobraCmd.Flags().StringSliceVarP(&cmd.analyzerList, "analyzers", "a", []string{}, "Specific analyzers to run (comma-separated)")

	return cobraCmd
}

// Run executes the analyze command
func (c *AnalyzeCommand) Run(cmd *cobra.Command, args []string) error {
	// Create input reader
	inputReader := c.createInputReader()

	// Initialize analyzer service
	analyzerService := c.newService()

	// Run analysis and format results
	return analyzerService.AnalyzeAndFormat(inputReader, c.analyzerList, c.format, c.createOutputWriter())
}

// newService creates a new analyzer service
func (c *AnalyzeCommand) newService() *Service {
	return &Service{
		availableAnalyzers: []analyze.CodeAnalyzer{
			&complexity.ComplexityAnalyzer{},
			&halstead.HalsteadAnalyzer{},
			&cohesion.CohesionAnalyzer{},
		},
	}
}

// createInputReader creates an input reader (stdin or file)
func (c *AnalyzeCommand) createInputReader() *os.File {
	return os.Stdin
}

// createOutputWriter creates an output writer (stdout or file)
func (c *AnalyzeCommand) createOutputWriter() *os.File {
	if c.output == "" {
		return os.Stdout
	}

	file, err := os.Create(c.output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		return os.Stdout
	}

	return file
}

// Service provides a high-level interface for running analysis
type Service struct {
	availableAnalyzers []analyze.CodeAnalyzer
}

// AnalyzeAndFormat runs analysis and formats the results
func (s *Service) AnalyzeAndFormat(input io.Reader, analyzerList []string, format string, writer io.Writer) error {
	// Run analysis
	results, err := s.Analyze(input, analyzerList)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Format and output results
	if format == "json" {
		return s.formatJSON(results, writer)
	} else {
		return s.formatText(results, writer)
	}
}

// Analyze runs analysis on UAST input and returns aggregated results
func (s *Service) Analyze(input io.Reader, analyzerList []string) (map[string]analyze.Report, error) {
	// Read multiple JSON objects from input (one per file from uast parse)
	decoder := json.NewDecoder(input)
	allResults := make(map[string]analyze.Report)

	// Initialize aggregators for each analyzer
	aggregators := make(map[string]analyze.ResultAggregator)

	// Determine which analyzers to run
	analyzersToRun := analyzerList
	if len(analyzersToRun) == 0 {
		// Run all available analyzers
		for _, analyzer := range s.availableAnalyzers {
			analyzersToRun = append(analyzersToRun, analyzer.Name())
		}
	}

	// Initialize aggregators for each analyzer
	for _, analyzerName := range analyzersToRun {
		analyzer := s.findAnalyzer(analyzerName)
		if analyzer != nil {
			aggregators[analyzerName] = analyzer.CreateAggregator()
		}
	}

	for {
		var uastNode *node.Node
		err := decoder.Decode(&uastNode)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse UAST input: %w", err)
		}

		// Run analyzers for this file
		results, err := s.runAnalyzers(uastNode, analyzersToRun)
		if err != nil {
			return nil, fmt.Errorf("failed to run analyzers: %w", err)
		}

		// Aggregate results for each analyzer
		for analyzerName, aggregator := range aggregators {
			if report, ok := results[analyzerName]; ok {
				aggregator.Aggregate(map[string]analyze.Report{analyzerName: report})
			}
		}
	}

	// Build final results from aggregators
	for analyzerName, aggregator := range aggregators {
		allResults[analyzerName] = aggregator.GetResult()
	}

	return allResults, nil
}

// formatJSON formats all results as JSON
func (s *Service) formatJSON(results map[string]analyze.Report, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// formatText formats all results as human-readable text
func (s *Service) formatText(results map[string]analyze.Report, writer io.Writer) error {
	for analyzerName, report := range results {
		// Write analyzer header
		fmt.Fprintf(writer, "\n=== %s ===\n", strings.ToUpper(analyzerName))

		// Format the report using the analyzer's own formatter
		analyzer := s.findAnalyzer(analyzerName)
		if analyzer != nil {
			if err := analyzer.FormatReport(report, writer); err != nil {
				return fmt.Errorf("failed to format %s report: %w", analyzerName, err)
			}
		} else {
			// Fallback for unknown analyzers
			for key, value := range report {
				fmt.Fprintf(writer, "%s: %v\n", key, value)
			}
		}
	}
	return nil
}

// runAnalyzers runs the specified analyzers on a single UAST node
func (s *Service) runAnalyzers(uastNode *node.Node, analyzerList []string) (map[string]analyze.Report, error) {
	factory := analyze.NewFactory(s.availableAnalyzers)
	return factory.RunAnalyzers(uastNode, analyzerList)
}

// findAnalyzer finds an analyzer by name
func (s *Service) findAnalyzer(name string) analyze.CodeAnalyzer {
	for _, analyzer := range s.availableAnalyzers {
		if analyzer.Name() == name {
			return analyzer
		}
	}
	return nil
}
