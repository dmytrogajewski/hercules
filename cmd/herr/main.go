package main

import (
	"fmt"
	"os"

	"github.com/dmytrogajewski/hercules/cmd/herr/commands"
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
  • Halstead complexity measures
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

	// Add commands
	rootCmd.AddCommand(commands.NewAnalyzeCommand())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
