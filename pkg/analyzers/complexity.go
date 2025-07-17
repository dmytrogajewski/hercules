package analyzers

import (
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// CyclomaticComplexityAnalyzer implement CodeAnalyzer interface
type CyclomaticComplexityAnalyzer struct{}

func (c *CyclomaticComplexityAnalyzer) Name() string {
	return "cyclomatic_complexity"
}

func (c *CyclomaticComplexityAnalyzer) Thresholds() Thresholds {
	return Thresholds{
		"complexity": {
			"green":  1,
			"yellow": 5,
			"red":    10,
		},
	}
}

// Analyze cyclomatic complexity using formula
// Cyclomatic complexity is a software metric used to measure the complexity of a program's control flow.
// The formula is: V(G) = E - N + 2P where:
// E = Number of edges in the control flow graph (CFG)
// N = Number of nodes in the CFG
// P = Number of connected components (typically 1 for a single program or function)
func (c *CyclomaticComplexityAnalyzer) Analyze(root *node.Node) (map[string]any, error) {
	if root == nil {
		return map[string]any{
			"total_complexity": 0,
			"functions":        map[string]int{},
			"function_count":   0,
		}, nil
	}

	complexityByFunction := make(map[string]int)
	totalComplexity := 0

	// Find all function nodes using canonical roles/types
	functions := root.Find(func(n *node.Node) bool {
		return hasRole(n, "Function") || hasRole(n, "Declaration") || n.Type == "Function" || n.Type == "Method"
	})

	for _, fn := range functions {
		functionName := extractFunctionName(fn)
		complexity := calculateFunctionComplexity(fn)
		complexityByFunction[functionName] = complexity
		totalComplexity += complexity
	}

	return map[string]any{
		"total_complexity": totalComplexity,
		"functions":        complexityByFunction,
		"function_count":   len(functions),
	}, nil
}

func extractFunctionName(fn *node.Node) string {
	// Try to find function name using multiple strategies
	// 1. Look for Name role in children
	for _, child := range fn.Children {
		if hasRole(child, "Name") {
			return strings.TrimSpace(child.Token)
		}
	}

	// 2. Check props for name
	if name, ok := fn.Props["name"]; ok && name != "" {
		return strings.TrimSpace(name)
	}

	// 3. Check props for function name
	if name, ok := fn.Props["function_name"]; ok && name != "" {
		return strings.TrimSpace(name)
	}

	// 4. Look for identifier with Name role in subtree
	query := "rfilter(.roles has \"Name\")"
	matches, err := fn.FindDSL(query)
	if err == nil && len(matches) > 0 {
		return strings.TrimSpace(matches[0].Token)
	}

	// 5. Fallback to anonymous
	return "anonymous"
}

func calculateFunctionComplexity(fn *node.Node) int {
	// Base complexity is 1 for any function
	complexity := 1

	// Use comprehensive UAST query DSL to find all canonical decision points
	// This query matches all common decision point types across languages
	query := `rfilter(.type == "If" || .type == "Loop" || .type == "Switch" || .type == "Case" || .type == "Try" || .type == "Catch" || .type == "Throw" || .type == "Conditional" || .type == "While" || .type == "For" || .type == "DoWhile" || .type == "ForEach" || .type == "Guard" || .type == "Assert" || .type == "Break" || .type == "Continue" || .type == "Return" || .type == "Goto" || .type == "Label" || .type == "BinaryOp" || .type == "UnaryOp")`

	matches, err := fn.FindDSL(query)
	if err != nil {
		// DSL query failed - this indicates a problem with the query or implementation
		// We should fix the query rather than fall back to manual traversal
		return complexity
	}

	// Filter matches to only include actual decision points
	decisionPoints := 0
	for _, match := range matches {
		if isDecisionPoint(match) {
			decisionPoints++
		}
	}
	complexity += decisionPoints

	return complexity
}

func isDecisionPoint(n *node.Node) bool {
	// Check by type
	decisionTypes := []string{
		"If", "Loop", "Switch", "Case", "Try", "Catch", "Throw",
		"Conditional", "While", "For", "DoWhile", "ForEach",
		"Guard", "Assert", "Break", "Continue",
		"Return", "Goto", "Label",
	}

	for _, dt := range decisionTypes {
		if n.Type == dt {
			return true
		}
	}

	// Check BinaryOp and UnaryOp with specific operators
	if n.Type == "BinaryOp" || n.Type == "UnaryOp" {
		if operator, ok := n.Props["operator"]; ok {
			// Only logical operators count as decision points
			logicalOperators := []string{"&&", "||", "!"}
			for _, op := range logicalOperators {
				if operator == op {
					return true
				}
			}
		}
	}

	// Check by roles
	decisionRoles := []string{
		"Condition", "Decision", "ControlFlow", "Exception",
		"Break", "Continue", "Return", "Goto",
	}

	for _, role := range n.Roles {
		for _, dr := range decisionRoles {
			if string(role) == dr {
				return true
			}
		}
	}

	return false
}

func hasRole(n *node.Node, role string) bool {
	for _, r := range n.Roles {
		if string(r) == role {
			return true
		}
	}
	return false
}
