package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/analyzers"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "herr",
		Short: "Hercules Code Analysis - Analyze UAST output with various metrics",
		Long: `Herr (Hercules Code Analysis) provides comprehensive code analysis tools
that work with UAST output from the 'uast parse' command.

Key features:
  • Cyclomatic complexity analysis
  • Code structure metrics
  • Performance analysis
  • Quality assessment

Examples:
  uast parse main.go | herr analyze                    # Analyze single file
  uast parse *.go | herr analyze                       # Analyze all Go files
  uast parse main.go | herr analyze --format json     # JSON output
  uast parse main.go | herr analyze | herr report     # Generate comprehensive report
  uast parse --all | herr analyze | herr report       # Full codebase report`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")

	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(reportCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func analyzeCmd() *cobra.Command {
	var output, format string
	var analyzerList []string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze UAST input with various metrics",
		Long: `Analyze UAST input from stdin or files with various code analysis metrics.

Examples:
  uast parse main.go | herr analyze                    # Analyze from stdin
  herr analyze input.json                              # Analyze from file
  herr analyze --analyzers complexity input.json       # Run specific analyzer
  herr analyze --format json input.json                # JSON output format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(args, output, format, analyzerList)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format (text, json)")
	cmd.Flags().StringSliceVarP(&analyzerList, "analyzers", "a", []string{}, "specific analyzers to run (default: all)")

	return cmd
}

func reportCmd() *cobra.Command {
	var output, format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a comprehensive report of UAST parsing and analysis results",
		Long: `Generate a comprehensive report of UAST parsing and analysis results.

Examples:
  uast parse main.go | herr analyze | herr report     # Generate comprehensive report
  uast parse --all | herr analyze | herr report       # Full codebase report`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport(args, output, format)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "output format (text, json)")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("herr version 1.0.0")
		},
	}
}

func runAnalyze(args []string, output, format string, analyzerList []string) error {
	var input io.Reader = os.Stdin

	// If files are provided, read from the first file
	if len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()
		input = file
	}

	// Read multiple JSON objects from input (one per file from uast parse)
	decoder := json.NewDecoder(input)
	allResults := make(map[string]analyzers.AnalyzerReport)

	// Initialize combined results
	combinedComplexity := 0
	combinedFunctions := make(map[string]int)
	totalFunctionCount := 0

	for {
		var uastNode *node.Node
		err := decoder.Decode(&uastNode)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to parse UAST input: %w", err)
		}

		// Create analyzer factory with available analyzers
		availableAnalyzers := []analyzers.CodeAnalyzer{
			&analyzers.CyclomaticComplexityAnalyzer{},
		}

		factory := analyzers.NewFactory(availableAnalyzers)

		// Determine which analyzers to run
		analyzersToRun := analyzerList
		if len(analyzersToRun) == 0 {
			// Run all available analyzers
			for _, analyzer := range availableAnalyzers {
				analyzersToRun = append(analyzersToRun, analyzer.Name())
			}
		}

		// Run analyzers for this file
		results, err := factory.RunAnalyzers(uastNode, analyzersToRun)
		if err != nil {
			return fmt.Errorf("failed to run analyzers: %w", err)
		}

		// Aggregate results across files
		for analyzerName, report := range results {
			if analyzerName == "cyclomatic_complexity" {
				if complexity, ok := report["total_complexity"].(int); ok {
					combinedComplexity += complexity
				}
				if functions, ok := report["functions"].(map[string]int); ok {
					for funcName, funcComplexity := range functions {
						combinedFunctions[funcName] = funcComplexity
					}
				}
				if functionCount, ok := report["function_count"].(int); ok {
					totalFunctionCount += functionCount
				}
			}
		}
	}

	// Create combined results
	allResults["cyclomatic_complexity"] = analyzers.AnalyzerReport{
		"total_complexity": combinedComplexity,
		"functions":        combinedFunctions,
		"function_count":   totalFunctionCount,
	}

	// Output results
	return outputResults(allResults, output, format)
}

type CodebaseReport struct {
	CodebaseInfo    CodebaseInfo                        `json:"codebase_info"`
	AnalyzersReport map[string]analyzers.AnalyzerReport `json:"analyzers_report"`
}

type CodebaseInfo struct {
	TotalFiles     int            `json:"total_files"`
	FileTypes      map[string]int `json:"file_types"`
	TotalFunctions int            `json:"total_functions"`
	TotalLines     int            `json:"total_lines"`
	Languages      map[string]int `json:"languages"`
	Structure      StructureInfo  `json:"structure"`
}

type StructureInfo struct {
	Functions  int `json:"functions"`
	Classes    int `json:"classes"`
	Interfaces int `json:"interfaces"`
	Methods    int `json:"methods"`
	Variables  int `json:"variables"`
	Imports    int `json:"imports"`
	Comments   int `json:"comments"`
}

func runReport(args []string, output, format string) error {
	var input io.Reader = os.Stdin

	// If files are provided, read from the first file
	if len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()
		input = file
	}

	// Read multiple JSON objects from input (one per file from uast parse)
	decoder := json.NewDecoder(input)

	// Aggregate codebase info across all files
	combinedCodebaseInfo := CodebaseInfo{
		FileTypes: make(map[string]int),
		Languages: make(map[string]int),
		Structure: StructureInfo{},
	}

	// Aggregate analyzer results
	combinedAnalyzersReport := make(map[string]analyzers.AnalyzerReport)
	combinedComplexity := 0
	combinedFunctions := make(map[string]int)
	totalFunctionCount := 0
	fileCount := 0

	for {
		var uastNode *node.Node
		err := decoder.Decode(&uastNode)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to parse UAST input: %w", err)
		}

		fileCount++

		// Generate codebase info for this file
		codebaseInfo := generateCodebaseInfo(uastNode)

		// Aggregate codebase info
		combinedCodebaseInfo.TotalFiles++
		combinedCodebaseInfo.TotalFunctions += codebaseInfo.TotalFunctions
		combinedCodebaseInfo.Structure.Functions += codebaseInfo.Structure.Functions
		combinedCodebaseInfo.Structure.Classes += codebaseInfo.Structure.Classes
		combinedCodebaseInfo.Structure.Interfaces += codebaseInfo.Structure.Interfaces
		combinedCodebaseInfo.Structure.Variables += codebaseInfo.Structure.Variables
		combinedCodebaseInfo.Structure.Imports += codebaseInfo.Structure.Imports
		combinedCodebaseInfo.Structure.Comments += codebaseInfo.Structure.Comments

		// Create analyzer factory with available analyzers
		availableAnalyzers := []analyzers.CodeAnalyzer{
			&analyzers.CyclomaticComplexityAnalyzer{},
		}

		factory := analyzers.NewFactory(availableAnalyzers)

		// Run all analyzers
		analyzersToRun := []string{}
		for _, analyzer := range availableAnalyzers {
			analyzersToRun = append(analyzersToRun, analyzer.Name())
		}

		// Run analyzers for this file
		analyzersReport, err := factory.RunAnalyzers(uastNode, analyzersToRun)
		if err != nil {
			return fmt.Errorf("failed to run analyzers: %w", err)
		}

		// Aggregate analyzer results
		for analyzerName, report := range analyzersReport {
			if analyzerName == "cyclomatic_complexity" {
				if complexity, ok := report["total_complexity"].(int); ok {
					combinedComplexity += complexity
				}
				if functions, ok := report["functions"].(map[string]int); ok {
					for funcName, funcComplexity := range functions {
						// Add file prefix to avoid name conflicts
						uniqueName := fmt.Sprintf("%s (file %d)", funcName, fileCount)
						combinedFunctions[uniqueName] = funcComplexity
					}
				}
				if functionCount, ok := report["function_count"].(int); ok {
					totalFunctionCount += functionCount
				}
			}
		}
	}

	// Create combined analyzer results
	combinedAnalyzersReport["cyclomatic_complexity"] = analyzers.AnalyzerReport{
		"total_complexity": combinedComplexity,
		"functions":        combinedFunctions,
		"function_count":   totalFunctionCount,
	}

	// Create comprehensive report
	report := CodebaseReport{
		CodebaseInfo:    combinedCodebaseInfo,
		AnalyzersReport: combinedAnalyzersReport,
	}

	// Output report
	return outputReport(report, output, format)
}

func generateCodebaseInfo(root *node.Node) CodebaseInfo {
	if root == nil {
		return CodebaseInfo{}
	}

	info := CodebaseInfo{
		FileTypes: make(map[string]int),
		Languages: make(map[string]int),
		Structure: StructureInfo{},
	}

	// Count different node types
	root.VisitPreOrder(func(n *node.Node) {
		switch n.Type {
		case "Function", "Method":
			info.Structure.Functions++
		case "Class":
			info.Structure.Classes++
		case "Interface":
			info.Structure.Interfaces++
		case "Variable":
			info.Structure.Variables++
		case "Import":
			info.Structure.Imports++
		case "Comment", "DocString":
			info.Structure.Comments++
		}
	})

	info.TotalFunctions = info.Structure.Functions
	info.TotalFiles = 1 // For now, assuming single file input

	return info
}

func outputReport(report CodebaseReport, output, format string) error {
	var writer io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	switch format {
	case "json":
		return outputJSONReport(report, writer)
	case "text":
		return outputTextReport(report, writer)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func outputJSONReport(report CodebaseReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func outputTextReport(report CodebaseReport, writer io.Writer) error {
	c := color.New(color.FgCyan)

	// Header
	c.Fprintf(writer, "\n"+strings.Repeat("=", 60)+"\n")
	c.Fprintf(writer, "                    CODEBASE ANALYSIS REPORT\n")
	c.Fprintf(writer, strings.Repeat("=", 60)+"\n\n")

	// Codebase Info
	outputCodebaseInfo(report.CodebaseInfo, writer)

	// Analyzers Report
	outputAnalyzersReport(report.AnalyzersReport, writer)

	return nil
}

func outputCodebaseInfo(info CodebaseInfo, writer io.Writer) {
	c := color.New(color.FgBlue)
	c.Fprintf(writer, "📊 CODEBASE OVERVIEW\n")
	c.Fprintf(writer, strings.Repeat("-", 30)+"\n")

	c.Fprintf(writer, "📁 Total Files: %d\n", info.TotalFiles)
	c.Fprintf(writer, "🔧 Total Functions: %d\n", info.TotalFunctions)

	if len(info.FileTypes) > 0 {
		c.Fprintf(writer, "📄 File Types:\n")
		for fileType, count := range info.FileTypes {
			c.Fprintf(writer, "   • %s: %d\n", fileType, count)
		}
	}

	c.Fprintf(writer, "\n🏗️  STRUCTURE BREAKDOWN\n")
	c.Fprintf(writer, strings.Repeat("-", 30)+"\n")
	c.Fprintf(writer, "🔧 Functions: %d\n", info.Structure.Functions)
	c.Fprintf(writer, "🏛️  Classes: %d\n", info.Structure.Classes)
	c.Fprintf(writer, "📋 Interfaces: %d\n", info.Structure.Interfaces)
	c.Fprintf(writer, "📦 Variables: %d\n", info.Structure.Variables)
	c.Fprintf(writer, "📥 Imports: %d\n", info.Structure.Imports)
	c.Fprintf(writer, "💬 Comments: %d\n", info.Structure.Comments)
	c.Fprintf(writer, "\n")
}

func outputAnalyzersReport(analyzersReport map[string]analyzers.AnalyzerReport, writer io.Writer) {
	c := color.New(color.FgGreen)
	c.Fprintf(writer, "🔍 ANALYZERS REPORT\n")
	c.Fprintf(writer, strings.Repeat("-", 30)+"\n")

	for analyzerName, report := range analyzersReport {
		switch analyzerName {
		case "cyclomatic_complexity":
			outputComplexityReport(report, writer)
		default:
			outputGenericReport(report, writer)
		}
	}
}

func outputResults(results map[string]analyzers.AnalyzerReport, output, format string) error {
	var writer io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	switch format {
	case "json":
		return outputJSON(results, writer)
	case "text":
		return outputText(results, writer)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func outputJSON(results map[string]analyzers.AnalyzerReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func outputText(results map[string]analyzers.AnalyzerReport, writer io.Writer) error {
	c := color.New(color.FgCyan)

	for analyzerName, report := range results {
		c.Fprintf(writer, "\n=== %s ===\n", strings.ToUpper(analyzerName))

		switch analyzerName {
		case "cyclomatic_complexity":
			outputComplexityReport(report, writer)
		default:
			// Generic output for unknown analyzers
			outputGenericReport(report, writer)
		}
	}

	return nil
}

func outputComplexityReport(report analyzers.AnalyzerReport, writer io.Writer) {
	c := color.New(color.FgGreen)

	// Get thresholds for this analyzer
	analyzer := &analyzers.CyclomaticComplexityAnalyzer{}
	thresholds := analyzer.Thresholds()

	// Extract complexity thresholds
	complexityThresholds, ok := thresholds["complexity"]
	if !ok {
		c.Fprintf(writer, "Error: No complexity thresholds found\n")
		return
	}

	green, _ := complexityThresholds["green"].(int)
	yellow, _ := complexityThresholds["yellow"].(int)
	red, _ := complexityThresholds["red"].(int)

	// Output threshold information
	c.Fprintf(writer, "Complexity Thresholds:\n")
	c.Fprintf(writer, "  Green (Good): ≤ %d\n", green)
	c.Fprintf(writer, "  Yellow (Warning): ≤ %d\n", yellow)
	c.Fprintf(writer, "  Red (High): > %d\n\n", red)

	if totalComplexity, ok := report["total_complexity"].(int); ok {
		c.Fprintf(writer, "Total Complexity: %d\n", totalComplexity)
	}

	if functionCount, ok := report["function_count"].(int); ok {
		c.Fprintf(writer, "Functions Analyzed: %d\n", functionCount)
	}

	if functions, ok := report["functions"].(map[string]int); ok {
		if len(functions) > 0 {
			// Sort functions by complexity descending
			type funcEntry struct {
				Name       string
				Complexity int
			}
			var funcList []funcEntry
			for name, complexity := range functions {
				funcList = append(funcList, funcEntry{Name: name, Complexity: complexity})
			}
			sort.Slice(funcList, func(i, j int) bool {
				return funcList[i].Complexity > funcList[j].Complexity
			})

			c.Fprintf(writer, "\nFunction Breakdown (sorted by complexity):\n")
			for _, entry := range funcList {
				severity, emoji := getSeverityEmoji(entry.Complexity, green, yellow, red)
				severity.Fprintf(writer, "  %s %s: %d\n", emoji, entry.Name, entry.Complexity)
			}
		}
	}
}

func getSeverityEmoji(value, green, yellow, red int) (*color.Color, string) {
	if value <= green {
		return color.New(color.FgGreen), "🟢"
	} else if value <= yellow {
		return color.New(color.FgYellow), "🟡"
	} else {
		return color.New(color.FgRed), "🔴"
	}
}

func outputGenericReport(report analyzers.AnalyzerReport, writer io.Writer) {
	c := color.New(color.FgBlue)

	for key, value := range report {
		c.Fprintf(writer, "%s: %v\n", key, value)
	}
}
