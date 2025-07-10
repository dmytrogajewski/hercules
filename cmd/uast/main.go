package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dmytrogajewski/hercules/pkg/uast"

	"github.com/spf13/cobra"
)

var parseOutputWriter io.Writer = os.Stdout

func main() {
	rootCmd := &cobra.Command{
		Use:   "uast",
		Short: "Unified AST CLI: parse, query, format, and diff code using UAST",
	}

	rootCmd.AddCommand(parseCmd())
	rootCmd.AddCommand(queryCmd())
	rootCmd.AddCommand(fmtCmd())
	rootCmd.AddCommand(diffCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseCmd() *cobra.Command {
	var (
		lang   string
		out    string
		format string
	)
	cmd := &cobra.Command{
		Use:   "parse [file]",
		Short: "Parse source code to UAST (JSON/protobuf output)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				input    io.Reader = os.Stdin
				filename string
			)
			if len(args) > 0 && args[0] != "-" {
				f, err := os.Open(args[0])
				if err != nil {
					printErrAndExit("failed to open input file: %v", err)
				}
				defer f.Close()
				input = f
				filename = args[0]
			} else {
				filename = "stdin.go" // fallback for language detection
			}
			code, err := io.ReadAll(input)
			if err != nil {
				printErrAndExit("failed to read input: %v", err)
			}
			parser, err := uast.NewParser()
			if err != nil {
				printErrAndExit("failed to initialize parser: %v", err)
			}
			if lang != "" {
				// override filename extension for language detection
				filename = "file." + lang
			}
			node, err := parser.Parse(filename, code)
			if err != nil {
				printErrAndExit("parse error: %v", err)
			}
			var output io.Writer = parseOutputWriter
			if out != "" {
				f, err := os.Create(out)
				if err != nil {
					printErrAndExit("failed to create output file: %v", err)
				}
				defer f.Close()
				output = f
			}
			switch format {
			case "", "json":
				enc := json.NewEncoder(output)
				enc.SetIndent("", "  ")
				if err := enc.Encode(node); err != nil {
					printErrAndExit("failed to encode UAST as JSON: %v", err)
				}
			// TODO: support protobuf output
			case "proto":
				printErrAndExit("protobuf output not implemented yet")
			default:
				printErrAndExit("unknown format: %s", format)
			}
		},
	}
	cmd.Flags().StringVar(&lang, "lang", "", "Language (overrides file extension)")
	cmd.Flags().StringVar(&out, "out", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&format, "format", "json", "Output format: json|proto (default: json)")
	return cmd
}

func printErrAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func queryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Run DSL queries on UAST (input: file/stdin, output: JSON)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("query: not implemented yet")
		},
	}
}

func fmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fmt",
		Short: "Pretty-print or normalize UAST JSON",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("fmt: not implemented yet")
		},
	}
}

func diffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Compare two UASTs and report structural changes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("diff: not implemented yet")
		},
	}
}
