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
		Short: "Hercules Code Analysis - Unified code analysis tool",
		Long: `Herr (Hercules Code Analysis) provides comprehensive code analysis tools.

Commands:
  analyze   Static analysis on UAST output (complexity, cohesion, comments, halstead, imports)
  history   Git repository history analysis (burndown, couples, devs, file-history, imports, sentiment, shotness, typos)

Static Analysis Examples:
  uast parse main.go | herr analyze                        # Analyze single file
  uast parse main.go | herr analyze -a complexity,halstead # Specific analyzers
  uast parse *.go | herr analyze -f json                   # JSON output

History Analysis Examples:
  herr history -a burndown .                               # Burndown analysis
  herr history -a burndown,couples,devs .                  # Multiple analyzers
  herr history -a devs --head .                            # Latest commit only`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")

	// Add commands
	rootCmd.AddCommand(commands.NewAnalyzeCommand())
	rootCmd.AddCommand(commands.NewHistoryCommand())
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
