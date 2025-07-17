package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/mapping"
	"github.com/spf13/cobra"
)

func mappingCmd() *cobra.Command {
	var nodeTypesPath, mappingPath, format string
	var coverage, generate bool

	cmd := &cobra.Command{
		Use:   "mapping",
		Short: "UAST mapping helpers: grammar analysis, classification, coverage",
		Long:  `Analyze node-types.json, classify nodes, and compute mapping coverage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMappingHelper(nodeTypesPath, mappingPath, format, coverage, generate)
		},
	}

	cmd.Flags().StringVar(&nodeTypesPath, "node-types", "", "Path to node-types.json (required)")
	cmd.Flags().StringVar(&mappingPath, "mapping", "", "Path to mapping DSL file (optional)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	cmd.Flags().BoolVar(&coverage, "coverage", false, "Compute mapping coverage if mapping is provided")
	cmd.Flags().BoolVar(&generate, "generate", false, "Generate .uastmap DSL from node-types.json")
	cmd.MarkFlagRequired("node-types")

	return cmd
}

func runMappingHelper(nodeTypesPath, mappingPath, format string, coverage, generate bool) error {
	jsonData, err := os.ReadFile(nodeTypesPath)
	if err != nil {
		return fmt.Errorf("failed to read node-types.json: %w", err)
	}
	nodes, err := mapping.ParseNodeTypes(jsonData)
	if err != nil {
		return fmt.Errorf("failed to parse node-types.json: %w", err)
	}
	nodes = mapping.ApplyHeuristicClassification(nodes)

	if generate {
		dsl := mapping.GenerateMappingDSL(nodes)
		fmt.Print(dsl)
		return nil
	}

	var rules []mapping.MappingRule
	if mappingPath != "" {
		data, err := os.ReadFile(mappingPath)
		if err != nil {
			return fmt.Errorf("failed to read mapping DSL: %w", err)
		}
		rules, err = (&mapping.MappingParser{}).ParseMapping(string(data))
		if err != nil {
			return fmt.Errorf("failed to load mapping DSL: %w", err)
		}
	}

	if format == "json" {
		out := map[string]interface{}{
			"node_count": len(nodes),
			"categories": summarizeCategories(nodes),
			"nodes":      nodes,
		}
		if coverage && len(rules) > 0 {
			cov, err := mapping.CoverageAnalysis(rules, nodes)
			if err != nil {
				return err
			}
			out["coverage"] = cov
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	fmt.Printf("Node types: %d\n", len(nodes))
	cats := summarizeCategories(nodes)
	for cat, count := range cats {
		fmt.Printf("  %s: %d\n", cat, count)
	}
	if coverage && len(rules) > 0 {
		cov, err := mapping.CoverageAnalysis(rules, nodes)
		if err != nil {
			return err
		}
		fmt.Printf("Coverage: %.2f%%\n", cov*100)
	}
	return nil
}

func summarizeCategories(nodes []mapping.NodeTypeInfo) map[string]int {
	cats := map[string]int{}
	for _, n := range nodes {
		cats[fmt.Sprintf("%v", n.Category)]++
	}
	return cats
}
