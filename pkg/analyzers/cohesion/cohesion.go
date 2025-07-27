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
		return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
			"total_functions":   0,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"message":           "No functions found",
		}), nil
	}

	lcom := c.calculateLCOM(functions)
	cohesionScore := c.calculateCohesionScore(lcom, len(functions))
	functionCohesion := c.calculateFunctionCohesion(functions)
	message := c.getCohesionMessage(cohesionScore)

	// Build detailed functions table for display with assessments
	detailedFunctionsTable := make([]map[string]interface{}, 0, len(functions))
	for _, fn := range functions {
		// Determine assessments
		cohesionAssessment := c.getCohesionAssessment(fn.Cohesion)
		variableAssessment := c.getVariableAssessment(len(fn.Variables))
		sizeAssessment := c.getSizeAssessment(fn.LineCount)

		detailedFunctionsTable = append(detailedFunctionsTable, map[string]interface{}{
			"name":                fn.Name,
			"line_count":          fn.LineCount,
			"variable_count":      len(fn.Variables),
			"cohesion":            fn.Cohesion,
			"cohesion_assessment": cohesionAssessment,
			"variable_assessment": variableAssessment,
			"size_assessment":     sizeAssessment,
		})
	}

	// Build function details for result (simplified version)
	functionDetails := make([]map[string]interface{}, 0, len(functions))
	for _, fn := range functions {
		functionDetails = append(functionDetails, map[string]interface{}{
			"name":           fn.Name,
			"line_count":     fn.LineCount,
			"variable_count": len(fn.Variables),
			"cohesion":       fn.Cohesion,
		})
	}

	// Build the result with proper structure for common formatter
	result := analyze.Report{
		"analyzer_name":     "cohesion",
		"total_functions":   len(functions),
		"lcom":              lcom,
		"cohesion_score":    cohesionScore,
		"function_cohesion": functionCohesion,
		"functions":         detailedFunctionsTable,
		"message":           message,
	}

	return result, nil
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
	var functions []Function

	// Use generic traverser to find function nodes
	functionNodes := c.traverser.FindNodesByRoles(root, []string{"Function"})

	// Also find by type for broader coverage
	typeNodes := c.traverser.FindNodesByType(root, []string{"Function", "Method"})

	// Combine and deduplicate
	allNodes := make(map[*node.Node]bool)
	for _, node := range functionNodes {
		allNodes[node] = true
	}
	for _, node := range typeNodes {
		allNodes[node] = true
	}

	// Extract functions from nodes
	for node := range allNodes {
		functions = append(functions, c.extractFunction(node))
	}

	return functions, nil
}

// extractFunction extracts function data from a node
func (c *CohesionAnalyzer) extractFunction(n *node.Node) Function {
	var variables []string
	c.findVariables(n, &variables)

	name, _ := c.extractor.ExtractName(n, "function_name")
	if name == "" {
		name, _ = common.ExtractFunctionName(n)
	}

	lineCount := c.traverser.CountLines(n)

	function := Function{
		Name:      name,
		LineCount: lineCount,
		Variables: variables,
		Cohesion:  0.0, // Will be calculated below
	}

	function.Cohesion = c.calculateFunctionLevelCohesion(function)

	return function
}

// findVariables finds all variables in a function
func (c *CohesionAnalyzer) findVariables(n *node.Node, variables *[]string) {
	if n == nil {
		return
	}

	// Look for variable declarations and parameters
	if n.HasAnyType(node.UASTVariable, node.UASTParameter) && n.HasAnyRole(node.RoleDeclaration) {
		if name, ok := c.extractor.ExtractName(n, "variable_name"); ok && name != "" {
			*variables = append(*variables, name)
		} else if name, ok := common.ExtractVariableName(n); ok && name != "" {
			*variables = append(*variables, name)
		}
	}

	// Look for identifiers that represent variables
	if n.HasAnyType(node.UASTIdentifier) && n.HasAnyRole(node.RoleVariable, node.RoleName) {
		if name, ok := c.extractor.ExtractName(n, "variable_name"); ok && name != "" {
			*variables = append(*variables, name)
		} else if name, ok := common.ExtractVariableName(n); ok && name != "" {
			*variables = append(*variables, name)
		}
	}

	for _, child := range n.Children {
		c.findVariables(child, variables)
	}
}
