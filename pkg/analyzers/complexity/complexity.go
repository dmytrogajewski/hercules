package complexity

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/fatih/color"
)

// ComplexityAnalyzer provides comprehensive complexity analysis
type ComplexityAnalyzer struct{}

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
	return "complexity_analysis"
}

func (c *ComplexityAnalyzer) Thresholds() map[string]map[string]any {
	return map[string]map[string]any{
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
	return &ComplexityAggregator{
		combinedFunctions: make(map[string]int),
	}
}

// FormatReport formats complexity analysis results as human-readable text
func (c *ComplexityAnalyzer) FormatReport(report map[string]any, writer io.Writer) error {
	colorWriter := color.New(color.FgGreen)

	// Get thresholds for this analyzer
	thresholds := c.Thresholds()

	// Extract complexity thresholds
	complexityThresholds, ok := thresholds["cyclomatic_complexity"]
	if !ok {
		colorWriter.Fprintf(writer, "Error: No complexity thresholds found\n")
		return nil
	}

	green, _ := complexityThresholds["green"].(int)
	yellow, _ := complexityThresholds["yellow"].(int)
	red, _ := complexityThresholds["red"].(int)

	// Output threshold information
	colorWriter.Fprintf(writer, "Complexity Thresholds:\n")
	colorWriter.Fprintf(writer, "  Green (Good): ≤ %d\n", green)
	colorWriter.Fprintf(writer, "  Yellow (Warning): ≤ %d\n", yellow)
	colorWriter.Fprintf(writer, "  Red (High): > %d\n\n", red)

	if totalComplexity, ok := report["total_complexity"].(int); ok {
		colorWriter.Fprintf(writer, "Total Complexity: %d\n", totalComplexity)
	}

	if functionCount, ok := report["function_count"].(int); ok {
		colorWriter.Fprintf(writer, "Functions Analyzed: %d\n", functionCount)
	}

	if functions, ok := report["functions"].(map[string]int); ok {
		if len(functions) > 0 {
			// Sort functions by complexity descending
			type funcEntry struct {
				Name       string
				Complexity int
			}
			var funcList []funcEntry
			for name, complexity := range functions {
				funcList = append(funcList, funcEntry{Name: name, Complexity: complexity})
			}
			sort.Slice(funcList, func(i, j int) bool {
				return funcList[i].Complexity > funcList[j].Complexity
			})

			colorWriter.Fprintf(writer, "\nFunction Breakdown (sorted by complexity):\n")
			for _, entry := range funcList {
				severity, emoji := c.getSeverityEmoji(entry.Complexity, green, yellow)
				severity.Fprintf(writer, "  %s %s: %d\n", emoji, entry.Name, entry.Complexity)
			}
		}
	} else if functions, ok := report["functions"].(map[string]any); ok {
		// Handle complex metrics format (fallback)
		if len(functions) > 0 {
			type funcEntry struct {
				Name       string
				Complexity int
			}
			var funcList []funcEntry
			for name, funcData := range functions {
				if funcMetrics, ok := funcData.(map[string]any); ok {
					if cyclomatic, ok := funcMetrics["cyclomatic_complexity"].(int); ok {
						funcList = append(funcList, funcEntry{Name: name, Complexity: cyclomatic})
					}
				}
			}
			sort.Slice(funcList, func(i, j int) bool {
				return funcList[i].Complexity > funcList[j].Complexity
			})

			colorWriter.Fprintf(writer, "\nFunction Breakdown (sorted by complexity):\n")
			for _, entry := range funcList {
				severity, emoji := c.getSeverityEmoji(entry.Complexity, green, yellow)
				severity.Fprintf(writer, "  %s %s: %d\n", emoji, entry.Name, entry.Complexity)
			}
		}
	}

	return nil
}

// FormatReportJSON formats complexity analysis results as JSON
func (c *ComplexityAnalyzer) FormatReportJSON(report map[string]any, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// getSeverityEmoji returns the appropriate color and emoji for complexity levels
func (c *ComplexityAnalyzer) getSeverityEmoji(value, green, yellow int) (*color.Color, string) {
	if value <= green {
		return color.New(color.FgGreen), "🟢"
	} else if value <= yellow {
		return color.New(color.FgYellow), "🟡"
	} else {
		return color.New(color.FgRed), "🔴"
	}
}

// ComplexityAggregator aggregates complexity analysis results
type ComplexityAggregator struct {
	totalComplexity    int
	combinedFunctions  map[string]int
	totalFunctionCount int
}

func (c *ComplexityAggregator) Aggregate(results map[string]map[string]any) {
	if report, ok := results["complexity_analysis"]; ok {
		if complexity, ok := report["total_complexity"].(int); ok {
			c.totalComplexity += complexity
		}
		if functions, ok := report["functions"].(map[string]any); ok {
			for funcName, funcData := range functions {
				if funcMetrics, ok := funcData.(map[string]any); ok {
					if cyclomatic, ok := funcMetrics["cyclomatic_complexity"].(int); ok {
						c.combinedFunctions[funcName] = cyclomatic
					}
				}
			}
		}
		if functionCount, ok := report["function_count"].(int); ok {
			c.totalFunctionCount += functionCount
		}
	}
}

func (c *ComplexityAggregator) GetResult() map[string]any {
	return map[string]any{
		"total_complexity": c.totalComplexity,
		"functions":        c.combinedFunctions,
		"function_count":   c.totalFunctionCount,
	}
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
	return c.AnalyzeWithConfig(root, c.DefaultConfig())
}

func (c *ComplexityAnalyzer) AnalyzeWithConfig(root *node.Node, config ComplexityConfig) (map[string]any, error) {
	if root == nil {
		return c.emptyResult(), nil
	}

	// Find all functions and methods
	functions := c.findFunctions(root)

	// Calculate metrics for each function
	functionMetrics, totals := c.calculateAllFunctionMetrics(functions, config)

	// Build result
	result := c.buildResult(functionMetrics, totals, len(functions), config)

	return result, nil
}

func (c *ComplexityAnalyzer) findFunctions(root *node.Node) []*node.Node {
	return root.Find(func(n *node.Node) bool {
		return n.HasRole(node.RoleFunction) ||
			n.HasRole(node.RoleDeclaration) ||
			n.Type == node.UASTFunction ||
			n.Type == node.UASTMethod ||
			n.Type == node.UASTFunctionDecl ||
			n.Type == node.UASTClass
	})
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

func (c *ComplexityAnalyzer) buildResult(functionMetrics map[string]FunctionMetrics, totals map[string]int, functionCount int, config ComplexityConfig) map[string]any {
	// Calculate averages
	var avgComplexity float64
	if functionCount > 0 {
		avgComplexity = float64(totals["cyclomatic"]) / float64(functionCount)
	}

	// Build result
	result := c.buildBaseResult(totals, functionCount, avgComplexity)

	// Convert FunctionMetrics to map format for compatibility
	result["functions"] = c.convertFunctionMetrics(functionMetrics)

	// Add detailed metrics if requested
	c.addDetailedMetrics(result, totals, config)

	return result
}

func (c *ComplexityAnalyzer) buildBaseResult(totals map[string]int, functionCount int, avgComplexity float64) map[string]any {
	return map[string]any{
		"total_complexity":           totals["cyclomatic"],
		"total_cognitive_complexity": totals["cognitive"],
		"total_nesting_depth":        totals["nesting"],
		"total_decision_points":      totals["decisions"],
		"function_count":             functionCount,
		"average_complexity":         avgComplexity,
		"max_complexity":             totals["max"],
		"complexity_distribution": map[string]int{
			"green":  totals["distribution_green"],
			"yellow": totals["distribution_yellow"],
			"red":    totals["distribution_red"],
		},
	}
}

func (c *ComplexityAnalyzer) convertFunctionMetrics(functionMetrics map[string]FunctionMetrics) map[string]any {
	functionsMap := make(map[string]any)
	for name, metrics := range functionMetrics {
		functionsMap[name] = map[string]any{
			"name":                  metrics.Name,
			"cyclomatic_complexity": metrics.CyclomaticComplexity,
			"cognitive_complexity":  metrics.CognitiveComplexity,
			"nesting_depth":         metrics.NestingDepth,
			"decision_points":       metrics.DecisionPoints,
			"lines_of_code":         metrics.LinesOfCode,
			"parameters":            metrics.Parameters,
			"return_statements":     metrics.ReturnStatements,
		}
	}
	return functionsMap
}

func (c *ComplexityAnalyzer) addDetailedMetrics(result map[string]any, totals map[string]int, config ComplexityConfig) {
	metrics := []struct {
		include bool
		key     string
		value   int
	}{
		{config.IncludeCognitiveComplexity, "cognitive_complexity", totals["cognitive"]},
		{config.IncludeNestingDepth, "nesting_depth", totals["nesting"]},
		{config.IncludeDecisionPoints, "decision_points", totals["decisions"]},
	}

	for _, metric := range metrics {
		if metric.include {
			result[metric.key] = metric.value
		}
	}
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
	// Look for parameter nodes
	params := fn.Find(func(n *node.Node) bool {
		return n.HasRole(node.RoleArgument) || n.HasRole(node.RoleParameter)
	})

	return len(params)
}

func (c *ComplexityAnalyzer) countReturnStatements(fn *node.Node) int {
	returnStmts := fn.Find(func(n *node.Node) bool {
		return n.Type == node.UASTReturn || n.HasRole(node.RoleReturn)
	})

	return len(returnStmts)
}

func (c *ComplexityAnalyzer) isDecisionPoint(n *node.Node) bool {
	// Check node types that are always decision points
	if c.isAlwaysDecisionPoint(n.Type) {
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
		return n.HasRole(node.RoleCondition)
	case node.UASTLoop:
		return c.hasLoopRole(n)
	case node.UASTBinaryOp, node.UASTUnaryOp:
		return c.hasLogicalOperator(n)
	}
	return false
}

func (c *ComplexityAnalyzer) hasLoopRole(n *node.Node) bool {
	return n.HasRole(node.RoleCondition) || n.HasRole(node.RoleLoop)
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
	// Try to find name using UAST query first
	matches, err := fn.FindDSL("rfilter(.roles has \"Name\")")
	if err == nil && len(matches) > 0 {
		return strings.TrimSpace(matches[0].Token)
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

	// Try to find identifier with name role
	return c.findIdentifierWithNameRole(fn)
}

func (c *ComplexityAnalyzer) findIdentifierWithNameRole(fn *node.Node) string {
	matches, err := fn.FindDSL("rfilter(.type == \"Identifier\" && .roles has \"Name\")")
	if err == nil && len(matches) > 0 {
		return strings.TrimSpace(matches[0].Token)
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

	// Try to find class name in the same scope using DSL
	matches, err := fn.FindDSL("rfilter(.type == \"Class\" && .roles has \"Declaration\")")
	if err == nil && len(matches) > 0 {
		return c.extractFunctionName(matches[0])
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
	// Check properties first
	name := c.extractNameFromProps(fn)
	if name != "" {
		return name
	}

	// Try DSL query
	matches, err := fn.FindDSL("rfilter(.roles has \"Name\")")
	if err == nil && len(matches) > 0 {
		return strings.TrimSpace(matches[0].Token)
	}

	// Look for method name in children
	return c.findMethodNameInChildren(fn)
}

func (c *ComplexityAnalyzer) findMethodNameInChildren(fn *node.Node) string {
	for _, child := range fn.Children {
		if child.HasRole(node.RoleName) {
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

func (c *ComplexityAnalyzer) emptyResult() map[string]any {
	return map[string]any{
		"total_complexity":           0,
		"total_cognitive_complexity": 0,
		"total_nesting_depth":        0,
		"total_decision_points":      0,
		"functions":                  map[string]FunctionMetrics{},
		"function_count":             0,
		"average_complexity":         0.0,
		"max_complexity":             0,
		"complexity_distribution":    map[string]int{"green": 0, "yellow": 0, "red": 0},
	}
}

// Legacy compatibility - keep the old interface
type CyclomaticComplexityAnalyzer struct{}

func (c *CyclomaticComplexityAnalyzer) Name() string {
	return "cyclomatic_complexity"
}

func (c *CyclomaticComplexityAnalyzer) Thresholds() map[string]map[string]any {
	return map[string]map[string]any{
		"complexity": {
			"green":  1,
			"yellow": 5,
			"red":    10,
		},
	}
}

func (c *CyclomaticComplexityAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	analyzer := &ComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)
	if err != nil {
		return nil, err
	}

	// Convert to legacy format
	return analyze.Report{
		"total_complexity": result["total_complexity"],
		"functions":        result["functions"],
		"function_count":   result["function_count"],
	}, nil
}
