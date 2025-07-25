package cohesion

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// CohesionAnalyzer implements the CodeAnalyzer interface for cohesion analysis
type CohesionAnalyzer struct{}

// Name returns the analyzer name
func (c *CohesionAnalyzer) Name() string {
	return "cohesion"
}

// Thresholds returns the color-coded thresholds for cohesion metrics
func (c *CohesionAnalyzer) Thresholds() analyze.Thresholds {
	return analyze.Thresholds{
		"lcom": {
			"red":    4.0,
			"yellow": 2.0,
			"green":  1.0,
		},
		"cohesion_score": {
			"red":    0.3,
			"yellow": 0.6,
			"green":  0.8,
		},
		"function_cohesion": {
			"red":    0.2,
			"yellow": 0.5,
			"green":  0.7,
		},
	}
}

// Analyze performs cohesion analysis on the UAST
func (c *CohesionAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	if root == nil {
		return nil, fmt.Errorf("root node is nil")
	}

	// Find all functions/methods in the code
	functions := c.findFunctions(root)
	if len(functions) == 0 {
		return analyze.Report{
			"total_functions":   0,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"message":           "No functions found",
		}, nil
	}

	// Calculate LCOM (Lack of Cohesion of Methods)
	lcom := c.calculateLCOM(functions)

	// Calculate cohesion score (inverse of LCOM, normalized to 0-1)
	cohesionScore := c.calculateCohesionScore(lcom, len(functions))

	// Calculate function-level cohesion
	functionCohesion := c.calculateFunctionCohesion(functions)

	// Prepare detailed results
	functionDetails := make([]map[string]interface{}, 0)
	for _, fn := range functions {
		functionDetails = append(functionDetails, map[string]interface{}{
			"name":           fn.Name,
			"line_count":     fn.LineCount,
			"variable_count": len(fn.Variables),
			"cohesion":       fn.Cohesion,
		})
	}

	return analyze.Report{
		"total_functions":   len(functions),
		"lcom":              lcom,
		"cohesion_score":    cohesionScore,
		"function_cohesion": functionCohesion,
		"functions":         functionDetails,
		"message":           c.getCohesionMessage(cohesionScore),
	}, nil
}

// CreateAggregator creates a new result aggregator for cohesion analysis
func (c *CohesionAnalyzer) CreateAggregator() analyze.ResultAggregator {
	return &CohesionAggregator{
		combinedFunctions: make(map[string]interface{}),
		totalLCOM:         0.0,
		totalCohesion:     0.0,
		functionCount:     0,
	}
}

// FormatReport formats the cohesion analysis report for human reading
func (c *CohesionAnalyzer) FormatReport(report analyze.Report, writer io.Writer) error {
	if report == nil {
		return fmt.Errorf("report is nil")
	}

	fmt.Fprintf(writer, "🔗 Cohesion Analysis Report\n")
	fmt.Fprintf(writer, "==========================\n\n")

	// Overall metrics
	totalFunctions, _ := report["total_functions"].(int)
	lcom, _ := report["lcom"].(float64)
	cohesionScore, _ := report["cohesion_score"].(float64)
	functionCohesion, _ := report["function_cohesion"].(float64)
	message, _ := report["message"].(string)

	fmt.Fprintf(writer, "📊 Overall Metrics:\n")
	fmt.Fprintf(writer, "   Total Functions: %d\n", totalFunctions)
	fmt.Fprintf(writer, "   LCOM (Lack of Cohesion): %.2f\n", lcom)
	fmt.Fprintf(writer, "   Cohesion Score: %.2f\n", cohesionScore)
	fmt.Fprintf(writer, "   Function Cohesion: %.2f\n", functionCohesion)
	fmt.Fprintf(writer, "   Assessment: %s\n\n", message)

	// Function details
	if functions, ok := report["functions"].([]map[string]interface{}); ok && len(functions) > 0 {
		fmt.Fprintf(writer, "🔍 Function Details:\n")
		for i, fn := range functions {
			name, _ := fn["name"].(string)
			lineCount, _ := fn["line_count"].(int)
			variableCount, _ := fn["variable_count"].(int)
			cohesion, _ := fn["cohesion"].(float64)

			severity := c.getSeverityEmoji(cohesion, "function_cohesion")
			fmt.Fprintf(writer, "   %d. %s %s\n", i+1, severity, name)
			fmt.Fprintf(writer, "      Lines: %d, Variables: %d, Cohesion: %.2f\n", lineCount, variableCount, cohesion)
		}
		fmt.Fprintf(writer, "\n")
	}

	// Recommendations
	fmt.Fprintf(writer, "💡 Recommendations:\n")
	if cohesionScore < 0.3 {
		fmt.Fprintf(writer, "   • Consider splitting large functions into smaller, more focused ones\n")
		fmt.Fprintf(writer, "   • Reduce the number of variables per function\n")
		fmt.Fprintf(writer, "   • Group related functionality together\n")
	} else if cohesionScore < 0.6 {
		fmt.Fprintf(writer, "   • Some functions could benefit from refactoring\n")
		fmt.Fprintf(writer, "   • Consider extracting utility functions\n")
	} else {
		fmt.Fprintf(writer, "   • Good cohesion! Functions are well-focused\n")
		fmt.Fprintf(writer, "   • Maintain current structure\n")
	}

	return nil
}

// FormatReportJSON formats the cohesion analysis report as JSON
func (c *CohesionAnalyzer) FormatReportJSON(report analyze.Report, writer io.Writer) error {
	if report == nil {
		return fmt.Errorf("report is nil")
	}

	// Convert to JSON-compatible format
	jsonReport := map[string]interface{}{
		"analyzer":          "cohesion",
		"total_functions":   report["total_functions"],
		"lcom":              report["lcom"],
		"cohesion_score":    report["cohesion_score"],
		"function_cohesion": report["function_cohesion"],
		"message":           report["message"],
		"functions":         report["functions"],
	}

	// Use proper JSON encoding
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonReport)
}

// Function represents a function with its cohesion metrics
type Function struct {
	Name      string
	LineCount int
	Variables []string
	Cohesion  float64
}

// findFunctions finds all functions in the UAST using DSL queries
func (c *CohesionAnalyzer) findFunctions(root *node.Node) []Function {
	var functions []Function

	// Use DSL to find all function and method nodes
	functionNodes, err := root.FindDSL("filter(.type == \"Function\" || .type == \"Method\" || .roles has \"Function\")")
	if err != nil {
		// Fallback to manual traversal if DSL fails
		c.traverseFunctions(root, &functions)
		return functions
	}

	// Extract function details from found nodes
	for _, node := range functionNodes {
		functions = append(functions, c.extractFunction(node))
	}

	return functions
}

// findVariables finds all variables in a function using manual traversal
func (c *CohesionAnalyzer) findVariables(n *node.Node, variables *[]string) {
	if n == nil {
		return
	}

	// Check if this is a variable declaration or parameter
	if n.Type == "Variable" || n.Type == "Parameter" {
		if name := c.extractVariableName(n); name != "" {
			*variables = append(*variables, name)
		}
	}

	// Traverse children to find local variables in function bodies
	for _, child := range n.Children {
		c.findVariables(child, variables)
	}
}

// traverseFunctions is the fallback manual traversal method
func (c *CohesionAnalyzer) traverseFunctions(n *node.Node, functions *[]Function) {
	if n == nil {
		return
	}

	if c.isFunctionNode(n) {
		*functions = append(*functions, c.extractFunction(n))
	}

	for _, child := range n.Children {
		c.traverseFunctions(child, functions)
	}
}

// isFunctionNode checks if a node represents a function
func (c *CohesionAnalyzer) isFunctionNode(n *node.Node) bool {
	if n == nil {
		return false
	}

	// Check for function types
	if n.Type == "Function" || n.Type == "Method" {
		return true
	}

	// Check for function roles - must have Function role AND be a declaration
	hasFunctionRole := false
	hasDeclarationRole := false
	for _, role := range n.Roles {
		if role == "Function" {
			hasFunctionRole = true
		}
		if role == "Declaration" {
			hasDeclarationRole = true
		}
	}

	// Must have both Function and Declaration roles to be a real function
	return hasFunctionRole && hasDeclarationRole
}

// extractFunction extracts function information from a function node
func (c *CohesionAnalyzer) extractFunction(n *node.Node) Function {
	fn := Function{
		Name:      c.extractFunctionName(n),
		LineCount: c.calculateLineCount(n),
		Variables: c.extractVariables(n),
	}

	// Calculate function-level cohesion
	fn.Cohesion = c.calculateFunctionLevelCohesion(fn)
	return fn
}

// extractFunctionName extracts the function name from a function node
func (c *CohesionAnalyzer) extractFunctionName(n *node.Node) string {
	if n == nil {
		return "unknown"
	}

	// Try to get name from props
	if props := n.Props; props != nil {
		if name, ok := props["name"]; ok {
			return name
		}
	}

	// Try to get name from token
	if n.Token != "" {
		return n.Token
	}

	// Try to find name in children
	for _, child := range n.Children {
		if child.Type == "Identifier" {
			if slices.Contains(child.Roles, "Name") {
				return child.Token
			}
		}
	}

	return "anonymous"
}

// calculateLineCount calculates the number of lines in a function
func (c *CohesionAnalyzer) calculateLineCount(n *node.Node) int {
	if n == nil {
		return 0
	}

	// If position information is available, calculate from that
	if n.Pos != nil {
		return int(n.Pos.EndLine - n.Pos.StartLine + 1)
	}

	// Otherwise, count by traversing and counting nodes with tokens
	count := 0
	c.countLines(n, &count)
	return count
}

// countLines recursively counts lines in a function
func (c *CohesionAnalyzer) countLines(n *node.Node, count *int) {
	if n == nil {
		return
	}

	// Count this node if it has a token
	if n.Token != "" {
		*count++
	}

	// Traverse children
	for _, child := range n.Children {
		c.countLines(child, count)
	}
}

// extractVariables extracts variable names from a function
func (c *CohesionAnalyzer) extractVariables(n *node.Node) []string {
	var variables []string
	c.findVariables(n, &variables)
	return variables
}

// extractVariableName extracts variable name from a variable node
func (c *CohesionAnalyzer) extractVariableName(n *node.Node) string {
	if n == nil {
		return ""
	}

	// For parameters, prefer the name from props (this gives us the clean parameter name)
	if props := n.Props; props != nil {
		if name, ok := props["name"]; ok {
			return name
		}
	}

	// For other variables, try token first
	if n.Token != "" {
		// Skip parameter declarations that contain parentheses (like "(n *Node)")
		if n.Type == "Parameter" && (strings.Contains(n.Token, "(") || strings.Contains(n.Token, ")")) {
			return ""
		}

		// For Variable nodes, extract just the variable name from assignments
		if n.Type == "Variable" && (strings.Contains(n.Token, ":=") || strings.Contains(n.Token, "=")) {
			parts := strings.Split(n.Token, " ")
			if len(parts) > 0 {
				varName := strings.TrimSpace(parts[0])
				if varName != "" && varName != "_" {
					return varName
				}
			}
		}

		return n.Token
	}

	// Only look in children if no props or token found
	for _, child := range n.Children {
		if child.Type == "Identifier" {
			for _, role := range child.Roles {
				if role == "Name" {
					return child.Token
				}
			}
		}
	}

	return ""
}

// calculateLCOM calculates the Lack of Cohesion of Methods
func (c *CohesionAnalyzer) calculateLCOM(functions []Function) float64 {
	if len(functions) <= 1 {
		return 0.0
	}

	// Count shared variables between function pairs
	sharedPairs := 0
	totalPairs := 0

	for i := 0; i < len(functions); i++ {
		for j := i + 1; j < len(functions); j++ {
			totalPairs++
			if c.haveSharedVariables(functions[i], functions[j]) {
				sharedPairs++
			}
		}
	}

	if totalPairs == 0 {
		return 0.0
	}

	// LCOM = (P - Q) where P = pairs that don't share variables, Q = pairs that do
	// P = totalPairs - sharedPairs, Q = sharedPairs
	return float64(totalPairs-sharedPairs) - float64(sharedPairs)
}

// haveSharedVariables checks if two functions share any variables
func (c *CohesionAnalyzer) haveSharedVariables(fn1, fn2 Function) bool {
	for _, var1 := range fn1.Variables {
		for _, var2 := range fn2.Variables {
			if var1 == var2 {
				return true
			}
		}
	}
	return false
}

// calculateCohesionScore calculates a normalized cohesion score (0-1)
func (c *CohesionAnalyzer) calculateCohesionScore(lcom float64, functionCount int) float64 {
	if functionCount <= 1 {
		return 1.0
	}

	// Normalize LCOM to 0-1 range
	maxLCOM := float64(functionCount * (functionCount - 1) / 2)
	if maxLCOM == 0 {
		return 1.0
	}

	// Cohesion score is inverse of normalized LCOM
	normalizedLCOM := lcom / maxLCOM
	cohesionScore := 1.0 - normalizedLCOM

	// Ensure score is between 0 and 1
	return math.Max(0.0, math.Min(1.0, cohesionScore))
}

// calculateFunctionCohesion calculates average function-level cohesion
func (c *CohesionAnalyzer) calculateFunctionCohesion(functions []Function) float64 {
	if len(functions) == 0 {
		return 1.0
	}

	total := 0.0
	for _, fn := range functions {
		total += fn.Cohesion
	}

	return total / float64(len(functions))
}

// calculateFunctionLevelCohesion calculates cohesion for a single function
func (c *CohesionAnalyzer) calculateFunctionLevelCohesion(fn Function) float64 {
	if fn.LineCount == 0 {
		return 1.0
	}

	// Simple cohesion metric: fewer variables per line = higher cohesion
	variableDensity := float64(len(fn.Variables)) / float64(fn.LineCount)

	// Normalize to 0-1 range (lower density = higher cohesion)
	cohesion := 1.0 - math.Min(1.0, variableDensity)

	return cohesion
}

// getCohesionMessage returns a human-readable message based on cohesion score
func (c *CohesionAnalyzer) getCohesionMessage(score float64) string {
	switch {
	case score >= 0.8:
		return "Excellent cohesion - functions are well-focused and related"
	case score >= 0.6:
		return "Good cohesion - functions are generally well-organized"
	case score >= 0.3:
		return "Fair cohesion - some functions could benefit from refactoring"
	default:
		return "Poor cohesion - functions lack focus and should be refactored"
	}
}

// getSeverityEmoji returns an emoji based on the metric value and thresholds
func (c *CohesionAnalyzer) getSeverityEmoji(value float64, metricName string) string {
	thresholds := c.Thresholds()
	if metricThresholds, ok := thresholds[metricName]; ok {
		if green, ok := metricThresholds["green"].(float64); ok && value >= green {
			return "🟢"
		}
		if yellow, ok := metricThresholds["yellow"].(float64); ok && value >= yellow {
			return "🟡"
		}
		return "🔴"
	}
	return "⚪"
}

// CohesionAggregator aggregates results from multiple cohesion analyses
type CohesionAggregator struct {
	combinedFunctions map[string]interface{}
	totalLCOM         float64
	totalCohesion     float64
	functionCount     int
}

// Aggregate combines multiple cohesion analysis results
func (ca *CohesionAggregator) Aggregate(results map[string]analyze.Report) {
	for _, report := range results {
		if report == nil {
			continue
		}

		// Aggregate LCOM
		if lcom, ok := report["lcom"].(float64); ok {
			ca.totalLCOM += lcom
		}

		// Aggregate cohesion scores
		if cohesion, ok := report["cohesion_score"].(float64); ok {
			ca.totalCohesion += cohesion
		}

		// Count functions
		if count, ok := report["total_functions"].(int); ok {
			ca.functionCount += count
		}

		// Combine function details
		if functions, ok := report["functions"].([]map[string]interface{}); ok {
			for _, fn := range functions {
				if name, ok := fn["name"].(string); ok {
					ca.combinedFunctions[name] = fn
				}
			}
		}
	}
}

// GetResult returns the aggregated cohesion analysis result
func (ca *CohesionAggregator) GetResult() analyze.Report {
	if ca.functionCount == 0 {
		return analyze.Report{
			"total_functions":   0,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"message":           "No functions found",
		}
	}

	avgLCOM := ca.totalLCOM / float64(len(ca.combinedFunctions))
	avgCohesion := ca.totalCohesion / float64(len(ca.combinedFunctions))

	// Convert combined functions back to slice
	functions := make([]map[string]interface{}, 0, len(ca.combinedFunctions))
	for _, fn := range ca.combinedFunctions {
		if fnMap, ok := fn.(map[string]interface{}); ok {
			functions = append(functions, fnMap)
		}
	}

	// Sort functions by name for consistent output
	sort.Slice(functions, func(i, j int) bool {
		nameI, _ := functions[i]["name"].(string)
		nameJ, _ := functions[j]["name"].(string)
		return strings.Compare(nameI, nameJ) < 0
	})

	return analyze.Report{
		"total_functions":   ca.functionCount,
		"lcom":              avgLCOM,
		"cohesion_score":    avgCohesion,
		"function_cohesion": avgCohesion,
		"functions":         functions,
		"message":           ca.getAggregatedMessage(avgCohesion),
	}
}

// getAggregatedMessage returns a message for aggregated results
func (ca *CohesionAggregator) getAggregatedMessage(score float64) string {
	switch {
	case score >= 0.8:
		return "Excellent overall cohesion across all analyzed code"
	case score >= 0.6:
		return "Good overall cohesion with room for improvement"
	case score >= 0.3:
		return "Fair overall cohesion - consider refactoring some functions"
	default:
		return "Poor overall cohesion - significant refactoring recommended"
	}
}
