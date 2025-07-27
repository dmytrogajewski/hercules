package complexity

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// ComplexityAnalyzer provides comprehensive complexity analysis
type ComplexityAnalyzer struct {
	traverser *common.UASTTraverser
	extractor *common.DataExtractor
}

// NewComplexityAnalyzer creates a new ComplexityAnalyzer
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		traverser: common.NewUASTTraverser(common.TraversalConfig{
			MaxDepth:    10,
			IncludeRoot: true,
		}),
		extractor: common.NewDataExtractor(common.ExtractionConfig{
			DefaultExtractors: true,
		}),
	}
}

// ComplexityMetrics holds different types of complexity measurements
type ComplexityMetrics struct {
	CyclomaticComplexity   int                        `json:"cyclomatic_complexity"`
	CognitiveComplexity    int                        `json:"cognitive_complexity"`
	NestingDepth           int                        `json:"nesting_depth"`
	DecisionPoints         int                        `json:"decision_points"`
	FunctionMetrics        map[string]FunctionMetrics `json:"function_metrics"`
	TotalFunctions         int                        `json:"total_functions"`
	AverageComplexity      float64                    `json:"average_complexity"`
	MaxComplexity          int                        `json:"max_complexity"`
	ComplexityDistribution map[string]int             `json:"complexity_distribution"`
}

// FunctionMetrics holds complexity metrics for individual functions
type FunctionMetrics struct {
	Name                 string `json:"name"`
	CyclomaticComplexity int    `json:"cyclomatic_complexity"`
	CognitiveComplexity  int    `json:"cognitive_complexity"`
	NestingDepth         int    `json:"nesting_depth"`
	DecisionPoints       int    `json:"decision_points"`
	LinesOfCode          int    `json:"lines_of_code"`
	Parameters           int    `json:"parameters"`
	ReturnStatements     int    `json:"return_statements"`
}

// ComplexityConfig holds configuration for complexity analysis
type ComplexityConfig struct {
	IncludeCognitiveComplexity bool
	IncludeNestingDepth        bool
	IncludeDecisionPoints      bool
	IncludeLOCMetrics          bool
	MaxNestingDepth            int
	ComplexityThresholds       map[string]int
}

func (c *ComplexityAnalyzer) Name() string {
	return "complexity"
}

// Thresholds returns the color-coded thresholds for complexity metrics
func (c *ComplexityAnalyzer) Thresholds() analyze.Thresholds {
	return analyze.Thresholds{
		"cyclomatic_complexity": {
			"green":  1,
			"yellow": 5,
			"red":    10,
		},
		"cognitive_complexity": {
			"green":  1,
			"yellow": 7,
			"red":    15,
		},
		"nesting_depth": {
			"green":  1,
			"yellow": 3,
			"red":    5,
		},
	}
}

// CreateAggregator returns a new aggregator for complexity analysis
func (c *ComplexityAnalyzer) CreateAggregator() analyze.ResultAggregator {
	return NewComplexityAggregator()
}

// FormatReport formats complexity analysis results as human-readable text
func (c *ComplexityAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
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

// FormatReportJSON formats complexity analysis results as JSON
func (c *ComplexityAnalyzer) FormatReportJSON(report analyze.Report, w io.Writer) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(jsonData))
	return err
}

// DefaultConfig returns default complexity analysis configuration
func (c *ComplexityAnalyzer) DefaultConfig() ComplexityConfig {
	return ComplexityConfig{
		IncludeCognitiveComplexity: true,
		IncludeNestingDepth:        true,
		IncludeDecisionPoints:      true,
		IncludeLOCMetrics:          true,
		MaxNestingDepth:            10,
		ComplexityThresholds: map[string]int{
			"cyclomatic_green":  1,
			"cyclomatic_yellow": 5,
			"cyclomatic_red":    10,
			"cognitive_green":   1,
			"cognitive_yellow":  7,
			"cognitive_red":     15,
			"nesting_green":     1,
			"nesting_yellow":    3,
			"nesting_red":       5,
		},
	}
}

func (c *ComplexityAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	if root == nil {
		return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
			"total_functions":    0,
			"average_complexity": 0.0,
			"max_complexity":     0,
			"total_complexity":   0,
			"message":            "No functions found",
		}), nil
	}

	config := c.DefaultConfig()
	functions := c.findFunctions(root)

	if len(functions) == 0 {
		return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
			"total_functions":    0,
			"average_complexity": 0.0,
			"max_complexity":     0,
			"total_complexity":   0,
			"message":            "No functions found",
		}), nil
	}

	// Calculate metrics for each function
	functionMetrics, totals := c.calculateAllFunctionMetrics(functions, config)

	// Build detailed functions table for display with assessments
	detailedFunctionsTable := make([]map[string]interface{}, 0, len(functionMetrics))
	for _, metrics := range functionMetrics {
		// Determine complexity assessments
		complexityAssessment := c.getComplexityAssessment(metrics.CyclomaticComplexity, config.ComplexityThresholds)
		cognitiveAssessment := c.getCognitiveAssessment(metrics.CognitiveComplexity)
		nestingAssessment := c.getNestingAssessment(metrics.NestingDepth)

		detailedFunctionsTable = append(detailedFunctionsTable, map[string]interface{}{
			"name":                  metrics.Name,
			"cyclomatic_complexity": metrics.CyclomaticComplexity,
			"cognitive_complexity":  metrics.CognitiveComplexity,
			"nesting_depth":         metrics.NestingDepth,
			"lines_of_code":         metrics.LinesOfCode,
			"complexity_assessment": complexityAssessment,
			"cognitive_assessment":  cognitiveAssessment,
			"nesting_assessment":    nestingAssessment,
		})
	}

	// Build function details for result (simplified version)
	functionDetails := make([]map[string]interface{}, 0, len(functionMetrics))
	for _, metrics := range functionMetrics {
		functionDetails = append(functionDetails, map[string]interface{}{
			"name":                  metrics.Name,
			"cyclomatic_complexity": metrics.CyclomaticComplexity,
			"cognitive_complexity":  metrics.CognitiveComplexity,
			"nesting_depth":         metrics.NestingDepth,
			"decision_points":       metrics.DecisionPoints,
			"lines_of_code":         metrics.LinesOfCode,
			"parameters":            metrics.Parameters,
			"return_statements":     metrics.ReturnStatements,
		})
	}

	// Calculate averages
	var avgComplexity float64
	if len(functions) > 0 {
		avgComplexity = float64(totals["cyclomatic"]) / float64(len(functions))
	}

	// Build message
	message := c.getComplexityMessage(avgComplexity)

	// Build the result with proper structure for common formatter
	result := analyze.Report{
		"analyzer_name":        "complexity",
		"total_functions":      len(functions),
		"average_complexity":   avgComplexity,
		"max_complexity":       totals["max"],
		"total_complexity":     totals["cyclomatic"],
		"cognitive_complexity": totals["cognitive"],
		"nesting_depth":        totals["nesting"],
		"decision_points":      totals["decisions"],
		"functions":            detailedFunctionsTable,
		"message":              message,
	}

	return result, nil
}

// findFunctions finds all functions in the UAST using common traverser
func (c *ComplexityAnalyzer) findFunctions(root *node.Node) []*node.Node {
	// Use common traverser to find function nodes
	functionNodes := c.traverser.FindNodesByType(root, []string{node.UASTFunction, node.UASTMethod})

	// Also find by roles for broader coverage
	roleNodes := c.traverser.FindNodesByRoles(root, []string{node.RoleFunction})

	// Combine and deduplicate
	allNodes := make(map[*node.Node]bool)
	for _, node := range functionNodes {
		allNodes[node] = true
	}
	for _, node := range roleNodes {
		allNodes[node] = true
	}

	// Convert back to slice
	var functions []*node.Node
	for node := range allNodes {
		if c.isFunctionNode(node) {
			functions = append(functions, node)
		}
	}

	return functions
}

// isFunctionNode checks if a node represents a function
func (c *ComplexityAnalyzer) isFunctionNode(n *node.Node) bool {
	if n == nil {
		return false
	}

	return n.HasAnyType(node.UASTFunction, node.UASTMethod) ||
		n.HasAllRoles(node.RoleFunction, node.RoleDeclaration)
}

func (c *ComplexityAnalyzer) calculateAllFunctionMetrics(functions []*node.Node, config ComplexityConfig) (map[string]FunctionMetrics, map[string]int) {
	functionMetrics := make(map[string]FunctionMetrics)
	totals := map[string]int{
		"cyclomatic": 0,
		"cognitive":  0,
		"nesting":    0,
		"decisions":  0,
		"max":        0,
	}

	complexityDistribution := map[string]int{
		"green":  0,
		"yellow": 0,
		"red":    0,
	}

	for _, fn := range functions {
		metrics := c.calculateFunctionMetrics(fn)
		functionMetrics[metrics.Name] = metrics

		totals["cyclomatic"] += metrics.CyclomaticComplexity
		totals["cognitive"] += metrics.CognitiveComplexity
		totals["nesting"] += metrics.NestingDepth
		totals["decisions"] += metrics.DecisionPoints

		if metrics.CyclomaticComplexity > totals["max"] {
			totals["max"] = metrics.CyclomaticComplexity
		}

		// Update complexity distribution
		complexityLevel := c.getComplexityLevel(metrics.CyclomaticComplexity, config.ComplexityThresholds)
		complexityDistribution[complexityLevel]++
	}

	totals["distribution_green"] = complexityDistribution["green"]
	totals["distribution_yellow"] = complexityDistribution["yellow"]
	totals["distribution_red"] = complexityDistribution["red"]

	return functionMetrics, totals
}

func (c *ComplexityAnalyzer) calculateFunctionMetrics(fn *node.Node) FunctionMetrics {
	name := c.extractFunctionName(fn)

	metrics := FunctionMetrics{
		Name:                 name,
		CyclomaticComplexity: c.calculateCyclomaticComplexity(fn),
		CognitiveComplexity:  c.calculateCognitiveComplexity(fn),
		NestingDepth:         c.calculateNestingDepth(fn),
		DecisionPoints:       c.countDecisionPoints(fn),
		LinesOfCode:          c.estimateLinesOfCode(fn),
		Parameters:           c.countParameters(fn),
		ReturnStatements:     c.countReturnStatements(fn),
	}

	return metrics
}

func (c *ComplexityAnalyzer) calculateCyclomaticComplexity(fn *node.Node) int {
	complexity := 1 // Base complexity

	fn.VisitPreOrder(func(n *node.Node) {
		if c.isDecisionPoint(n) {
			complexity++
		}
	})

	return complexity
}

func (c *ComplexityAnalyzer) calculateCognitiveComplexity(fn *node.Node) int {
	complexity := 0

	fn.VisitPreOrder(func(n *node.Node) {
		if c.isCognitiveComplexityPoint(n) {
			complexity++
		}
	})

	return complexity
}

func (c *ComplexityAnalyzer) calculateNestingDepth(fn *node.Node) int {
	maxDepth := 0
	currentDepth := 0

	fn.VisitPreOrder(func(n *node.Node) {
		if c.isNestingStart(n) {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if c.isNestingEnd(n) {
			currentDepth--
		}
	})

	return maxDepth
}

func (c *ComplexityAnalyzer) countDecisionPoints(fn *node.Node) int {
	count := 0

	fn.VisitPreOrder(func(n *node.Node) {
		if c.isDecisionPoint(n) {
			count++
		}
	})

	return count
}

func (c *ComplexityAnalyzer) estimateLinesOfCode(fn *node.Node) int {
	// Simple estimation based on node count and token length
	loc := 0

	fn.VisitPreOrder(func(n *node.Node) {
		if n.Token != "" {
			lines := strings.Count(n.Token, "\n") + 1
			loc += lines
		}
	})

	return loc
}

func (c *ComplexityAnalyzer) countParameters(fn *node.Node) int {
	// Look for parameter nodes using common traverser
	paramNodes := c.traverser.FindNodesByRoles(fn, []string{node.RoleArgument, node.RoleParameter})
	return len(paramNodes)
}

func (c *ComplexityAnalyzer) countReturnStatements(fn *node.Node) int {
	// Look for return statements using common traverser
	returnNodes := c.traverser.FindNodesByType(fn, []string{node.UASTReturn})
	returnStmts := c.traverser.FindNodesByRoles(fn, []string{node.RoleReturn})
	return len(returnNodes) + len(returnStmts)
}

func (c *ComplexityAnalyzer) isDecisionPoint(n *node.Node) bool {
	// Check node types that are always decision points
	if c.isAlwaysDecisionPoint(string(n.Type)) {
		return true
	}

	// Check conditional node types
	if c.isConditionalDecisionPoint(n) {
		return true
	}

	// Check role-based decision points
	return c.hasDecisionRole(n)
}

func (c *ComplexityAnalyzer) isAlwaysDecisionPoint(nodeType string) bool {
	switch nodeType {
	case node.UASTSwitch, node.UASTCase, node.UASTTry, node.UASTCatch,
		node.UASTThrow, node.UASTBreak, node.UASTContinue:
		return true
	default:
		return false
	}
}

func (c *ComplexityAnalyzer) isConditionalDecisionPoint(n *node.Node) bool {
	switch n.Type {
	case node.UASTIf:
		return n.HasAnyRole(node.RoleCondition)
	case node.UASTLoop:
		return c.hasLoopRole(n)
	case node.UASTBinaryOp, node.UASTUnaryOp:
		return c.hasLogicalOperator(n)
	}
	return false
}

func (c *ComplexityAnalyzer) hasLoopRole(n *node.Node) bool {
	return n.HasAnyRole(node.RoleCondition) || n.HasAnyRole(node.RoleLoop)
}

func (c *ComplexityAnalyzer) hasLogicalOperator(n *node.Node) bool {
	if operator, ok := n.Props["operator"]; ok {
		return c.isLogicalOperator(operator)
	}
	return false
}

func (c *ComplexityAnalyzer) hasDecisionRole(n *node.Node) bool {
	for _, role := range n.Roles {
		if string(role) == node.RoleCondition ||
			string(role) == node.RoleBreak ||
			string(role) == node.RoleContinue {
			return true
		}
	}
	return false
}

func (c *ComplexityAnalyzer) isCognitiveComplexityPoint(n *node.Node) bool {
	// Cognitive complexity includes additional constructs
	if c.isDecisionPoint(n) {
		return true
	}

	// Additional cognitive complexity points
	if n.Type == node.UASTCall || n.Type == node.UASTFunction {
		// Function calls can increase cognitive load
		return true
	}

	// Nested expressions
	if n.Type == node.UASTBinaryOp && len(n.Children) > 2 {
		return true
	}

	return false
}

func (c *ComplexityAnalyzer) isNestingStart(n *node.Node) bool {
	return n.Type == node.UASTIf || n.Type == node.UASTLoop ||
		n.Type == node.UASTSwitch || n.Type == node.UASTTry ||
		n.Type == node.UASTBlock || n.Type == node.UASTFunction
}

func (c *ComplexityAnalyzer) isNestingEnd(n *node.Node) bool {
	return n.Type == node.UASTBlock || n.Type == node.UASTFunction
}

func (c *ComplexityAnalyzer) isLogicalOperator(operator string) bool {
	logicalOps := map[string]bool{
		"&&": true, "||": true, "!": true,
		"and": true, "or": true, "not": true,
		"AND": true, "OR": true, "NOT": true,
	}
	return logicalOps[operator]
}

func (c *ComplexityAnalyzer) extractFunctionName(fn *node.Node) string {
	// Try common function name extraction first
	if name, ok := common.ExtractFunctionName(fn); ok && name != "" {
		return name
	}

	// Try to extract name using common extractor
	if name, ok := c.extractor.ExtractName(fn, "function_name"); ok && name != "" {
		return name
	}

	// Check properties for name
	name := c.extractNameFromProps(fn)
	if name != "" {
		return name
	}

	// For methods, try to get class name + method name
	if fn.Type == node.UASTMethod {
		name = c.extractMethodFullName(fn)
		if name != "" {
			return name
		}
	}

	// Try to find identifier with name role using common traverser
	nameNodes := c.traverser.FindNodesByRoles(fn, []string{node.RoleName})
	if len(nameNodes) > 0 {
		if name, ok := c.extractor.ExtractNameFromToken(nameNodes[0]); ok && name != "" {
			return name
		}
	}

	return "anonymous"
}

func (c *ComplexityAnalyzer) extractNameFromProps(fn *node.Node) string {
	props := []string{"name", "function_name", "method_name"}
	for _, prop := range props {
		if name, ok := fn.Props[prop]; ok && name != "" {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

func (c *ComplexityAnalyzer) extractMethodFullName(fn *node.Node) string {
	className := c.extractClassName(fn)
	methodName := c.extractMethodName(fn)
	if className != "" && methodName != "" {
		return className + "." + methodName
	}
	if methodName != "" {
		return methodName
	}
	return ""
}

func (c *ComplexityAnalyzer) extractClassName(fn *node.Node) string {
	// Check if there's a class name in properties first
	if className, ok := fn.Props["class_name"]; ok && className != "" {
		return strings.TrimSpace(className)
	}

	// Try to find class name using common traverser
	classNodes := c.traverser.FindNodesByType(fn, []string{node.UASTClass})
	if len(classNodes) > 0 {
		if name, ok := common.ExtractFunctionName(classNodes[0]); ok && name != "" {
			return name
		}
	}

	// Look for class name in ancestors
	return c.findClassNameInAncestors(fn)
}

func (c *ComplexityAnalyzer) findClassNameInAncestors(fn *node.Node) string {
	ancestors := fn.Ancestors(fn)
	for _, ancestor := range ancestors {
		if ancestor.Type == node.UASTClass {
			return c.extractFunctionName(ancestor)
		}
	}
	return ""
}

func (c *ComplexityAnalyzer) extractMethodName(fn *node.Node) string {
	// Try common function name extraction first
	if name, ok := common.ExtractFunctionName(fn); ok && name != "" {
		return name
	}

	// Check properties first
	name := c.extractNameFromProps(fn)
	if name != "" {
		return name
	}

	// Try to find name using common traverser
	nameNodes := c.traverser.FindNodesByRoles(fn, []string{node.RoleName})
	if len(nameNodes) > 0 {
		if name, ok := c.extractor.ExtractNameFromToken(nameNodes[0]); ok && name != "" {
			return name
		}
	}

	// Look for method name in children
	return c.findMethodNameInChildren(fn)
}

func (c *ComplexityAnalyzer) findMethodNameInChildren(fn *node.Node) string {
	for _, child := range fn.Children {
		if child.HasAnyRole(node.RoleName) {
			return strings.TrimSpace(child.Token)
		}
	}
	return ""
}

func (c *ComplexityAnalyzer) getComplexityLevel(complexity int, thresholds map[string]int) string {
	if complexity <= thresholds["cyclomatic_green"] {
		return "green"
	} else if complexity <= thresholds["cyclomatic_yellow"] {
		return "yellow"
	} else {
		return "red"
	}
}

// getComplexityMessage returns a message based on the average complexity score
func (c *ComplexityAnalyzer) getComplexityMessage(avgComplexity float64) string {
	if avgComplexity <= 1.0 {
		return "Excellent complexity - functions are simple and maintainable"
	}
	if avgComplexity <= 3.0 {
		return "Good complexity - functions have reasonable complexity"
	}
	if avgComplexity <= 7.0 {
		return "Fair complexity - some functions could be simplified"
	}
	return "High complexity - functions are complex and should be refactored"
}

// getComplexityAssessment returns an assessment with emoji for cyclomatic complexity
func (c *ComplexityAnalyzer) getComplexityAssessment(complexity int, thresholds map[string]int) string {
	level := c.getComplexityLevel(complexity, thresholds)
	switch level {
	case "green":
		return "🟢 Simple"
	case "yellow":
		return "🟡 Moderate"
	case "red":
		return "🔴 Complex"
	default:
		return "⚪ Unknown"
	}
}

// getCognitiveAssessment returns an assessment with emoji for cognitive complexity
func (c *ComplexityAnalyzer) getCognitiveAssessment(complexity int) string {
	if complexity <= 5 {
		return "🟢 Low"
	}
	if complexity <= 10 {
		return "🟡 Medium"
	}
	return "🔴 High"
}

// getNestingAssessment returns an assessment with emoji for nesting depth
func (c *ComplexityAnalyzer) getNestingAssessment(depth int) string {
	if depth <= 3 {
		return "🟢 Shallow"
	}
	if depth <= 5 {
		return "🟡 Moderate"
	}
	return "🔴 Deep"
}
