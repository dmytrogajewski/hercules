package main

import (
	"bytes"
	"strings"
	"testing"

	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/dmytrogajewski/hercules/pkg/uast"
	"github.com/spf13/cobra"
)

func TestUASTCLI_HelpAndSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		wantOut string
		wantErr bool
	}{
		{[]string{"--help"}, "Unified AST CLI", false},
		{[]string{"parse", "--help"}, "Parse source code", false},
		{[]string{"query", "--help"}, "Run DSL queries", false},
		{[]string{"fmt", "--help"}, "Pretty-print", false},
		{[]string{"diff", "--help"}, "Compare two UASTs", false},
		{[]string{"unknown"}, "unknown command", true},
	}
	for _, tt := range tests {
		rootCmd := buildTestRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(tt.args)
		err := rootCmd.Execute()
		out := buf.String()
		if tt.wantErr && err == nil {
			t.Errorf("args %v: expected error, got nil", tt.args)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("args %v: unexpected error: %v", tt.args, err)
		}
		if !strings.Contains(out, tt.wantOut) {
			t.Errorf("args %v: output missing %q\ngot: %s", tt.args, tt.wantOut, out)
		}
	}
}

func TestUASTCLI_ParseCommand_GoFunction(t *testing.T) {
	source := []byte(`package main
func add(a, b int) int { return a + b }`)
	tmpfile, err := ioutil.TempFile("", "test-*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(source); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	rootCmd := buildTestRootCmd()
	buf := new(bytes.Buffer)
	parseOutputWriter = buf
	defer func() { parseOutputWriter = os.Stdout }()
	rootCmd.SetOut(ioutil.Discard)
	rootCmd.SetErr(ioutil.Discard)
	rootCmd.SetArgs([]string{"parse", tmpfile.Name()})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("parse command failed: %v", err)
	}
	output := buf.Bytes()
	if len(output) == 0 {
		t.Fatalf("no output from parse command")
	}
	var node uast.Node
	if err := json.Unmarshal(output, &node); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, string(output))
	}
	// Find function node
	var fn *uast.Node
	for _, child := range node.Children {
		if child.Type == "go:function" || child.Type == "Function" || child.Type == "FunctionDecl" {
			fn = child
			break
		}
	}
	if fn == nil {
		t.Fatalf("No function node found; got children: %+v", node.Children)
	}
	if fn.Props["name"] != "add" {
		t.Errorf("Function node has wrong name prop: got %q, want 'add'", fn.Props["name"])
	}
}

func buildTestRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "uast",
		Short: "Unified AST CLI: parse, query, format, and diff code using UAST",
	}
	rootCmd.AddCommand(parseCmd())
	rootCmd.AddCommand(queryCmd())
	rootCmd.AddCommand(fmtCmd())
	rootCmd.AddCommand(diffCmd())
	return rootCmd
}
