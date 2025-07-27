package cohesion

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

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

	functions, err := c.findFunctions(root)
	if err != nil {
		return nil, err
	}

	if len(functions) == 0 {
		return c.buildEmptyResult(), nil
	}

	metrics := c.calculateMetrics(functions)
	result := c.buildResult(functions, metrics)

	return result, nil
}

// buildEmptyResult creates an empty result when no functions are found
func (c *CohesionAnalyzer) buildEmptyResult() analyze.Report {
	return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
		"total_functions":   0,
		"lcom":              0.0,
		"cohesion_score":    1.0,
		"function_cohesion": 1.0,
		"message":           "No functions found",
	})
}

// calculateMetrics calculates all cohesion metrics for the functions
func (c *CohesionAnalyzer) calculateMetrics(functions []Function) map[string]float64 {
	lcom := c.calculateLCOM(functions)
	cohesionScore := c.calculateCohesionScore(lcom, len(functions))
	functionCohesion := c.calculateFunctionCohesion(functions)

	return map[string]float64{
		"lcom":              lcom,
		"cohesion_score":    cohesionScore,
		"function_cohesion": functionCohesion,
	}
}

// buildResult constructs the final analysis result
func (c *CohesionAnalyzer) buildResult(functions []Function, metrics map[string]float64) analyze.Report {
	detailedFunctionsTable := c.buildDetailedFunctionsTable(functions)
	message := c.getCohesionMessage(metrics["cohesion_score"])

	return analyze.Report{
		"analyzer_name":     "cohesion",
		"total_functions":   len(functions),
		"lcom":              metrics["lcom"],
		"cohesion_score":    metrics["cohesion_score"],
		"function_cohesion": metrics["function_cohesion"],
		"functions":         detailedFunctionsTable,
		"message":           message,
	}
}

// buildDetailedFunctionsTable creates the detailed functions table with assessments
func (c *CohesionAnalyzer) buildDetailedFunctionsTable(functions []Function) []map[string]interface{} {
	table := make([]map[string]interface{}, 0, len(functions))

	for _, fn := range functions {
		entry := map[string]interface{}{
			"name":                fn.Name,
			"line_count":          fn.LineCount,
			"variable_count":      len(fn.Variables),
			"cohesion":            fn.Cohesion,
			"cohesion_assessment": c.getCohesionAssessment(fn.Cohesion),
			"variable_assessment": c.getVariableAssessment(len(fn.Variables)),
			"size_assessment":     c.getSizeAssessment(fn.LineCount),
		}
		table = append(table, entry)
	}

	return table
}

// FormatReport formats the analysis report for display
func (c *CohesionAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
	formatter := common.NewFormatter(common.FormatConfig{
		ShowProgressBars: true,
		ShowTables:       true,
		ShowDetails:      true,
		SkipHeader:       true,
	})

	formatted := formatter.FormatReport(report)
	_, err := fmt.Fprint(w, formatted)
	return err
}

// FormatReportJSON formats the analysis report as JSON
func (c *CohesionAnalyzer) FormatReportJSON(report analyze.Report, w io.Writer) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(jsonData))
	return err
}

// getCohesionMessage returns a message based on the cohesion score
func (c *CohesionAnalyzer) getCohesionMessage(score float64) string {
	if score >= 0.8 {
		return "Excellent cohesion - functions are well-focused and cohesive"
	}
	if score >= 0.6 {
		return "Good cohesion - functions have reasonable focus"
	}
	if score >= 0.3 {
		return "Fair cohesion - some functions could be more focused"
	}
	return "Poor cohesion - functions lack focus and should be refactored"
}

// getSeverityEmoji returns severity and emoji based on score and thresholds
func (c *CohesionAnalyzer) getSeverityEmoji(score, greenThreshold, yellowThreshold float64) (string, string) {
	if score >= greenThreshold {
		return "Good", "🟢"
	}
	if score >= yellowThreshold {
		return "Fair", "🟡"
	}
	return "Poor", "🔴"
}

// getCohesionAssessment returns an assessment with emoji for cohesion
func (c *CohesionAnalyzer) getCohesionAssessment(cohesion float64) string {
	if cohesion >= 0.8 {
		return "🟢 Excellent"
	}
	if cohesion >= 0.6 {
		return "🟡 Good"
	}
	if cohesion >= 0.3 {
		return "🟡 Fair"
	}
	return "🔴 Poor"
}

// getVariableAssessment returns an assessment with emoji for variable count
func (c *CohesionAnalyzer) getVariableAssessment(count int) string {
	if count <= 3 {
		return "🟢 Few"
	}
	if count <= 7 {
		return "🟡 Moderate"
	}
	return "🔴 Many"
}

// getSizeAssessment returns an assessment with emoji for function size
func (c *CohesionAnalyzer) getSizeAssessment(lineCount int) string {
	if lineCount <= 10 {
		return "🟢 Small"
	}
	if lineCount <= 30 {
		return "🟡 Medium"
	}
	return "🔴 Large"
}

// findFunctions finds all functions using the generic traverser
func (c *CohesionAnalyzer) findFunctions(root *node.Node) ([]Function, error) {
	functionNodes := c.traverser.FindNodesByRoles(root, []string{"Function"})
	typeNodes := c.traverser.FindNodesByType(root, []string{"Function", "Method"})

	allNodes := c.deduplicateNodes(functionNodes, typeNodes)
	return c.extractFunctionsFromNodes(allNodes), nil
}

// deduplicateNodes combines and deduplicates function nodes
func (c *CohesionAnalyzer) deduplicateNodes(functionNodes, typeNodes []*node.Node) []*node.Node {
	nodeMap := make(map[*node.Node]bool)

	for _, node := range functionNodes {
		nodeMap[node] = true
	}
	for _, node := range typeNodes {
		nodeMap[node] = true
	}

	result := make([]*node.Node, 0, len(nodeMap))
	for node := range nodeMap {
		result = append(result, node)
	}

	return result
}

// extractFunctionsFromNodes extracts Function structs from UAST nodes
func (c *CohesionAnalyzer) extractFunctionsFromNodes(nodes []*node.Node) []Function {
	functions := make([]Function, 0, len(nodes))

	for _, node := range nodes {
		functions = append(functions, c.extractFunction(node))
	}

	return functions
}

// extractFunction extracts function data from a node
func (c *CohesionAnalyzer) extractFunction(n *node.Node) Function {
	variables := c.extractVariables(n)
	name := c.extractFunctionName(n)
	lineCount := c.traverser.CountLines(n)

	function := Function{
		Name:      name,
		LineCount: lineCount,
		Variables: variables,
		Cohesion:  0.0,
	}

	function.Cohesion = c.calculateFunctionLevelCohesion(function)
	return function
}

// extractFunctionName extracts the function name from a node
func (c *CohesionAnalyzer) extractFunctionName(n *node.Node) string {
	name, _ := c.extractor.ExtractName(n, "function_name")
	if name == "" {
		name, _ = common.ExtractFunctionName(n)
	}
	return name
}

// extractVariables extracts all variables from a function node
func (c *CohesionAnalyzer) extractVariables(n *node.Node) []string {
	var variables []string
	c.findVariables(n, &variables)
	return variables
}

// findVariables finds all variables in a function
func (c *CohesionAnalyzer) findVariables(n *node.Node, variables *[]string) {
	if n == nil {
		return
	}

	c.processVariableNode(n, variables)
	c.processChildren(n, variables)
}

// processVariableNode processes a single node for variable extraction
func (c *CohesionAnalyzer) processVariableNode(n *node.Node, variables *[]string) {
	if c.isVariableDeclaration(n) {
		c.addVariableIfValid(n, variables)
	}

	if c.isVariableIdentifier(n) {
		c.addVariableIfValid(n, variables)
	}
}

// isVariableDeclaration checks if a node represents a variable declaration
func (c *CohesionAnalyzer) isVariableDeclaration(n *node.Node) bool {
	return n.HasAnyType(node.UASTVariable, node.UASTParameter) &&
		n.HasAnyRole(node.RoleDeclaration)
}

// isVariableIdentifier checks if a node represents a variable identifier
func (c *CohesionAnalyzer) isVariableIdentifier(n *node.Node) bool {
	return n.HasAnyType(node.UASTIdentifier) &&
		n.HasAnyRole(node.RoleVariable, node.RoleName)
}

// addVariableIfValid adds a variable name to the list if it's valid
func (c *CohesionAnalyzer) addVariableIfValid(n *node.Node, variables *[]string) {
	if name, ok := c.extractor.ExtractName(n, "variable_name"); ok && name != "" {
		*variables = append(*variables, name)
	} else if name, ok := common.ExtractVariableName(n); ok && name != "" {
		*variables = append(*variables, name)
	}
}

// processChildren recursively processes child nodes
func (c *CohesionAnalyzer) processChildren(n *node.Node, variables *[]string) {
	for _, child := range n.Children {
		c.findVariables(child, variables)
	}
}
