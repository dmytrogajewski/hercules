package analyzers

import (
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

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

var decisionPointTypes = map[string]bool{
	node.UASTIf:       true,
	node.UASTLoop:     true,
	node.UASTSwitch:   true,
	node.UASTCase:     true,
	node.UASTTry:      true,
	node.UASTCatch:    true,
	node.UASTThrow:    true,
	node.UASTBreak:    true,
	node.UASTContinue: true,
}

var logicalOperators = map[string]bool{
	"&&": true,
	"||": true,
	"!":  true,
}

var decisionPointRoles = map[string]bool{
	node.RoleCondition: true,
	node.RoleBreak:     true,
	node.RoleContinue:  true,
}

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

	functions := root.Find(func(n *node.Node) bool {
		return n.HasRole(node.RoleFunction) ||
			n.HasRole(node.RoleDeclaration) ||
			n.Type == node.UASTFunction ||
			n.Type == node.UASTMethod
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
	for _, child := range fn.Children {
		if child.HasRole(node.RoleName) {
			return strings.TrimSpace(child.Token)
		}
	}

	if name, ok := fn.Props["name"]; ok && name != "" {
		return strings.TrimSpace(name)
	}

	if name, ok := fn.Props["function_name"]; ok && name != "" {
		return strings.TrimSpace(name)
	}

	matches, err := fn.FindDSL("rfilter(.roles has \"Name\")")
	if err == nil && len(matches) > 0 {
		return strings.TrimSpace(matches[0].Token)
	}

	return "anonymous"
}

func calculateFunctionComplexity(fn *node.Node) int {
	complexity := 1
	fn.VisitPreOrder(func(n *node.Node) {
		if isDecisionPoint(n) {
			complexity++
		}
	})
	return complexity
}

func isDecisionPoint(n *node.Node) bool {
	if n.Type == node.UASTIf {
		for _, role := range n.Roles {
			if string(role) == node.RoleCondition || string(role) == node.RoleBranch {
				return true
			}
		}
		return false
	}

	if n.Type == node.UASTLoop || n.Type == node.UASTSwitch ||
		n.Type == node.UASTCase || n.Type == node.UASTTry || n.Type == node.UASTCatch ||
		n.Type == node.UASTThrow || n.Type == node.UASTBreak || n.Type == node.UASTContinue {
		return true
	}

	if n.Type == node.UASTBinaryOp || n.Type == node.UASTUnaryOp {
		if operator, ok := n.Props["operator"]; ok {
			return logicalOperators[operator]
		}
	}

	for _, role := range n.Roles {
		if string(role) == "Condition" {
			return true
		}
	}

	return false
}
