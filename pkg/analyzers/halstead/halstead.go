package halstead

import (
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/analyzers/common"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// HalsteadAnalyzer provides Halstead complexity measures analysis
type HalsteadAnalyzer struct {
	traverser *common.UASTTraverser
	extractor *common.DataExtractor
}

// NewHalsteadAnalyzer creates a new HalsteadAnalyzer with common modules
func NewHalsteadAnalyzer() *HalsteadAnalyzer {
	return &HalsteadAnalyzer{
		traverser: common.NewUASTTraverser(common.TraversalConfig{}),
		extractor: common.NewDataExtractor(common.ExtractionConfig{
			DefaultExtractors: true,
		}),
	}
}

// HalsteadMetrics holds all Halstead complexity measures
type HalsteadMetrics struct {
	// Basic counts
	DistinctOperators int `json:"distinct_operators"` // η1
	DistinctOperands  int `json:"distinct_operands"`  // η2
	TotalOperators    int `json:"total_operators"`    // N1
	TotalOperands     int `json:"total_operands"`     // N2

	// Calculated measures
	Vocabulary      int     `json:"vocabulary"`       // η = η1 + η2
	Length          int     `json:"length"`           // N = N1 + N2
	EstimatedLength float64 `json:"estimated_length"` // N̂ = η1*log2(η1) + η2*log2(η2)
	Volume          float64 `json:"volume"`           // V = N * log2(η)
	Difficulty      float64 `json:"difficulty"`       // D = (η1/2) * (N2/η2)
	Effort          float64 `json:"effort"`           // E = D * V
	TimeToProgram   float64 `json:"time_to_program"`  // T = E/18 (seconds)
	DeliveredBugs   float64 `json:"delivered_bugs"`   // B = V/3000

	// Per-function breakdown
	Functions map[string]*FunctionHalsteadMetrics `json:"functions"`
}

// FunctionHalsteadMetrics contains Halstead metrics for a single function
type FunctionHalsteadMetrics struct {
	Name              string  `json:"name"`
	DistinctOperators int     `json:"distinct_operators"`
	DistinctOperands  int     `json:"distinct_operands"`
	TotalOperators    int     `json:"total_operators"`
	TotalOperands     int     `json:"total_operands"`
	Vocabulary        int     `json:"vocabulary"`
	Length            int     `json:"length"`
	EstimatedLength   float64 `json:"estimated_length"`
	Volume            float64 `json:"volume"`
	Difficulty        float64 `json:"difficulty"`
	Effort            float64 `json:"effort"`
	TimeToProgram     float64 `json:"time_to_program"`
	DeliveredBugs     float64 `json:"delivered_bugs"`
}

// HalsteadConfig holds configuration for Halstead analysis
type HalsteadConfig struct {
	IncludeFunctionBreakdown bool
	IncludeTimeEstimate      bool
	IncludeBugEstimate       bool
}

func (h *HalsteadAnalyzer) Name() string {
	return "halstead"
}

func (h *HalsteadAnalyzer) Thresholds() analyze.Thresholds {
	return analyze.Thresholds{
		"volume": {
			"green":  100,
			"yellow": 1000,
			"red":    5000,
		},
		"difficulty": {
			"green":  5,
			"yellow": 15,
			"red":    30,
		},
		"effort": {
			"green":  1000,
			"yellow": 10000,
			"red":    50000,
		},
	}
}

// CreateAggregator returns a new aggregator for Halstead analysis
func (h *HalsteadAnalyzer) CreateAggregator() analyze.ResultAggregator {
	return &HalsteadAggregator{
		combinedFunctions: make(map[string]any),
	}
}

// Analyze performs Halstead analysis on the UAST
func (h *HalsteadAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	if root == nil {
		return nil, fmt.Errorf("root node is nil")
	}

	functions := h.findFunctions(root)

	if len(functions) == 0 {
		return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
			"total_functions": 0,
			"volume":          0.0,
			"difficulty":      0.0,
			"effort":          0.0,
			"message":         "No functions found",
		}), nil
	}

	// Calculate metrics for each function
	functionMetrics := make(map[string]*FunctionHalsteadMetrics)
	fileOperators := make(map[string]int)
	fileOperands := make(map[string]int)

	for _, fn := range functions {
		funcName := h.extractFunctionName(fn)
		if funcName == "" {
			funcName = "anonymous"
		}

		funcMetrics := h.calculateFunctionHalsteadMetrics(fn)
		funcMetrics.Name = funcName
		functionMetrics[funcName] = funcMetrics

		// Aggregate file-level metrics
		h.aggregateOperatorsAndOperands(fn, fileOperators, fileOperands)
	}

	// Calculate file-level metrics
	fileMetrics := &HalsteadMetrics{
		DistinctOperators: len(fileOperators),
		DistinctOperands:  len(fileOperands),
		TotalOperators:    h.sumMap(fileOperators),
		TotalOperands:     h.sumMap(fileOperands),
		Functions:         functionMetrics,
	}

	h.calculateHalsteadMetrics(fileMetrics)

	// Build detailed functions table for display with assessments
	detailedFunctionsTable := make([]map[string]interface{}, 0, len(functionMetrics))
	for _, fn := range functionMetrics {
		// Determine assessments
		volumeAssessment := h.getVolumeAssessment(fn.Volume)
		difficultyAssessment := h.getDifficultyAssessment(fn.Difficulty)
		effortAssessment := h.getEffortAssessment(fn.Effort)

		detailedFunctionsTable = append(detailedFunctionsTable, map[string]interface{}{
			"name":                  fn.Name,
			"volume":                fn.Volume,
			"difficulty":            fn.Difficulty,
			"effort":                fn.Effort,
			"delivered_bugs":        fn.DeliveredBugs,
			"volume_assessment":     volumeAssessment,
			"difficulty_assessment": difficultyAssessment,
			"effort_assessment":     effortAssessment,
		})
	}

	// Build function details for result (simplified version)
	functionDetails := make([]map[string]interface{}, 0, len(functionMetrics))
	for _, fn := range functionMetrics {
		functionDetails = append(functionDetails, map[string]interface{}{
			"name":               fn.Name,
			"volume":             fn.Volume,
			"difficulty":         fn.Difficulty,
			"effort":             fn.Effort,
			"time_to_program":    fn.TimeToProgram,
			"delivered_bugs":     fn.DeliveredBugs,
			"distinct_operators": fn.DistinctOperators,
			"distinct_operands":  fn.DistinctOperands,
		})
	}

	message := h.getHalsteadMessage(fileMetrics.Volume, fileMetrics.Difficulty, fileMetrics.Effort)

	// Build the result with proper structure for common formatter
	result := analyze.Report{
		"analyzer_name":      "halstead",
		"total_functions":    len(functionDetails),
		"functions":          detailedFunctionsTable,
		"volume":             fileMetrics.Volume,
		"difficulty":         fileMetrics.Difficulty,
		"effort":             fileMetrics.Effort,
		"time_to_program":    fileMetrics.TimeToProgram,
		"delivered_bugs":     fileMetrics.DeliveredBugs,
		"distinct_operators": fileMetrics.DistinctOperators,
		"distinct_operands":  fileMetrics.DistinctOperands,
		"total_operators":    fileMetrics.TotalOperators,
		"total_operands":     fileMetrics.TotalOperands,
		"vocabulary":         fileMetrics.Vocabulary,
		"length":             fileMetrics.Length,
		"estimated_length":   fileMetrics.EstimatedLength,
		"message":            message,
	}

	return result, nil
}

// FormatReport formats the analysis report for display
func (h *HalsteadAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
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
func (h *HalsteadAnalyzer) FormatReportJSON(report analyze.Report, w io.Writer) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(jsonData))
	return err
}

// getHalsteadMessage returns a message based on the Halstead metrics
func (h *HalsteadAnalyzer) getHalsteadMessage(volume, difficulty, effort float64) string {
	if volume <= 100 && difficulty <= 5 && effort <= 1000 {
		return "Excellent complexity - code is simple and maintainable"
	}
	if volume <= 1000 && difficulty <= 15 && effort <= 10000 {
		return "Good complexity - code is reasonably complex"
	}
	if volume <= 5000 && difficulty <= 30 && effort <= 50000 {
		return "Fair complexity - consider simplifying some functions"
	}
	return "High complexity - code should be refactored for better maintainability"
}

// getVolumeAssessment returns an assessment with emoji for volume
func (h *HalsteadAnalyzer) getVolumeAssessment(volume float64) string {
	if volume <= 100 {
		return "🟢 Low"
	}
	if volume <= 1000 {
		return "🟡 Medium"
	}
	return "🔴 High"
}

// getDifficultyAssessment returns an assessment with emoji for difficulty
func (h *HalsteadAnalyzer) getDifficultyAssessment(difficulty float64) string {
	if difficulty <= 5 {
		return "🟢 Simple"
	}
	if difficulty <= 15 {
		return "🟡 Moderate"
	}
	return "🔴 Complex"
}

// getEffortAssessment returns an assessment with emoji for effort
func (h *HalsteadAnalyzer) getEffortAssessment(effort float64) string {
	if effort <= 1000 {
		return "🟢 Low"
	}
	if effort <= 10000 {
		return "🟡 Medium"
	}
	return "🔴 High"
}

// getBugAssessment returns an assessment with emoji for delivered bugs
func (h *HalsteadAnalyzer) getBugAssessment(bugs float64) string {
	if bugs <= 0.1 {
		return "🟢 Low Risk"
	}
	if bugs <= 0.5 {
		return "🟡 Medium Risk"
	}
	return "🔴 High Risk"
}

// findFunctions finds all functions using the generic traverser
func (h *HalsteadAnalyzer) findFunctions(root *node.Node) []*node.Node {
	// Use common traverser to find function nodes
	functionNodes := h.traverser.FindNodesByType(root, []string{node.UASTFunction, node.UASTMethod})

	// Also find by roles for broader coverage
	roleNodes := h.traverser.FindNodesByRoles(root, []string{node.RoleFunction})

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
		if h.isFunctionNode(node) {
			functions = append(functions, node)
		}
	}

	return functions
}

// isFunctionNode checks if a node represents a function
func (h *HalsteadAnalyzer) isFunctionNode(n *node.Node) bool {
	if n == nil {
		return false
	}

	return n.HasAnyType(node.UASTFunction, node.UASTMethod) ||
		n.HasAllRoles(node.RoleFunction, node.RoleDeclaration)
}

// extractFunctionName extracts the function name using common extractor
func (h *HalsteadAnalyzer) extractFunctionName(n *node.Node) string {
	if name, ok := h.extractor.ExtractName(n, "function_name"); ok && name != "" {
		return name
	}
	if name, ok := common.ExtractFunctionName(n); ok && name != "" {
		return name
	}
	return ""
}

// calculateFunctionHalsteadMetrics calculates Halstead metrics for a single function
func (h *HalsteadAnalyzer) calculateFunctionHalsteadMetrics(fn *node.Node) *FunctionHalsteadMetrics {
	operators := make(map[string]int)
	operands := make(map[string]int)

	h.collectOperatorsAndOperands(fn, operators, operands)

	metrics := &FunctionHalsteadMetrics{
		DistinctOperators: len(operators),
		DistinctOperands:  len(operands),
		TotalOperators:    h.sumMap(operators),
		TotalOperands:     h.sumMap(operands),
	}

	h.calculateHalsteadMetrics(metrics)

	return metrics
}

// collectOperatorsAndOperands recursively collects operators and operands from a node
func (h *HalsteadAnalyzer) collectOperatorsAndOperands(node *node.Node, operators map[string]int, operands map[string]int) {
	if node == nil {
		return
	}

	// Determine if this node is an operator or operand
	if h.isOperator(node) {
		operator := h.getOperatorName(node)
		operators[string(operator)]++
	} else if h.isOperand(node) {
		operand := h.getOperandName(node)
		operands[string(operand)]++
	}

	// Recursively process children
	for _, child := range node.Children {
		h.collectOperatorsAndOperands(child, operators, operands)
	}
}

// aggregateOperatorsAndOperands aggregates operators and operands for file-level metrics
func (h *HalsteadAnalyzer) aggregateOperatorsAndOperands(node *node.Node, operators map[string]int, operands map[string]int) {
	h.collectOperatorsAndOperands(node, operators, operands)
}

// isOperator determines if a node represents an operator
func (h *HalsteadAnalyzer) isOperator(n *node.Node) bool {
	if n == nil {
		return false
	}

	// Check node type for operators
	switch n.Type {
	case node.UASTBinaryOp, node.UASTUnaryOp, node.UASTAssignment,
		node.UASTCall, node.UASTIndex, node.UASTSlice:
		return true
	}

	// Check for operator roles
	for _, role := range n.Roles {
		switch role {
		case node.RoleOperator, node.RoleAssignment, node.RoleCall:
			return true
		}
	}

	return false
}

// isOperand determines if a node represents an operand
func (h *HalsteadAnalyzer) isOperand(n *node.Node) bool {
	if n == nil {
		return false
	}

	// Check node type for operands
	switch n.Type {
	case node.UASTIdentifier, node.UASTLiteral, node.UASTVariable,
		node.UASTParameter, node.UASTField:
		return true
	}

	// Check for operand roles
	for _, role := range n.Roles {
		switch role {
		case node.RoleName, node.RoleLiteral, node.RoleVariable,
			node.RoleParameter, node.RoleArgument, node.RoleValue:
			return true
		}
	}

	return false
}

// getOperatorName extracts the operator name from a node
func (h *HalsteadAnalyzer) getOperatorName(n *node.Node) node.Type {
	if n == nil {
		return ""
	}

	// Try to get operator from properties
	if op, ok := n.Props["operator"]; ok {
		return node.Type(op)
	}

	// Fall back to type
	return n.Type
}

// getOperandName extracts the operand name from a node
func (h *HalsteadAnalyzer) getOperandName(n *node.Node) node.Type {
	if n == nil {
		return ""
	}

	// Try to get name from properties
	if name, ok := n.Props["name"]; ok {
		return node.Type(name)
	}

	if value, ok := n.Props["value"]; ok {
		return node.Type(value)
	}

	// Fall back to type
	return n.Type
}

// calculateHalsteadMetrics calculates all Halstead complexity measures
func (h *HalsteadAnalyzer) calculateHalsteadMetrics(metrics interface{}) {
	var m *HalsteadMetrics
	var fm *FunctionHalsteadMetrics

	switch v := metrics.(type) {
	case *HalsteadMetrics:
		m = v
	case *FunctionHalsteadMetrics:
		fm = v
		m = &HalsteadMetrics{
			DistinctOperators: fm.DistinctOperators,
			DistinctOperands:  fm.DistinctOperands,
			TotalOperators:    fm.TotalOperators,
			TotalOperands:     fm.TotalOperands,
		}
	}

	// Calculate basic measures
	m.Vocabulary = m.DistinctOperators + m.DistinctOperands
	m.Length = m.TotalOperators + m.TotalOperands

	// Calculate estimated length: N̂ = η1*log2(η1) + η2*log2(η2)
	if m.DistinctOperators > 0 {
		m.EstimatedLength += float64(m.DistinctOperators) * math.Log2(float64(m.DistinctOperators))
	}
	if m.DistinctOperands > 0 {
		m.EstimatedLength += float64(m.DistinctOperands) * math.Log2(float64(m.DistinctOperands))
	}

	// Calculate volume: V = N * log2(η)
	if m.Vocabulary > 0 {
		m.Volume = float64(m.Length) * math.Log2(float64(m.Vocabulary))
	}

	// Calculate difficulty: D = (η1/2) * (N2/η2)
	if m.DistinctOperands > 0 {
		m.Difficulty = (float64(m.DistinctOperators) / 2.0) * (float64(m.TotalOperands) / float64(m.DistinctOperands))
	}

	// Calculate effort: E = D * V
	m.Effort = m.Difficulty * m.Volume

	// Calculate time to program: T = E/18 (seconds)
	m.TimeToProgram = m.Effort / 18.0

	// Calculate delivered bugs: B = V/3000 (using the more recent formula)
	m.DeliveredBugs = m.Volume / 3000.0

	// Update function metrics if applicable
	if fm != nil {
		fm.Vocabulary = m.Vocabulary
		fm.Length = m.Length
		fm.EstimatedLength = m.EstimatedLength
		fm.Volume = m.Volume
		fm.Difficulty = m.Difficulty
		fm.Effort = m.Effort
		fm.TimeToProgram = m.TimeToProgram
		fm.DeliveredBugs = m.DeliveredBugs
	}
}

// sumMap sums all values in a map
func (h *HalsteadAnalyzer) sumMap(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

// HalsteadAggregator aggregates Halstead analysis results
type HalsteadAggregator struct {
	combinedVolume        float64
	combinedDifficulty    float64
	combinedEffort        float64
	combinedTimeToProgram float64
	combinedDeliveredBugs float64
	combinedFunctions     map[string]any
}

func (h *HalsteadAggregator) Aggregate(results map[string]analyze.Report) {
	if report, ok := results["halstead"]; ok {
		// Aggregate metrics directly from report
		if volume, ok := report["volume"].(float64); ok {
			h.combinedVolume += volume
		}
		if difficulty, ok := report["difficulty"].(float64); ok {
			h.combinedDifficulty += difficulty
		}
		if effort, ok := report["effort"].(float64); ok {
			h.combinedEffort += effort
		}
		if timeToProgram, ok := report["time_to_program"].(float64); ok {
			h.combinedTimeToProgram += timeToProgram
		}
		if deliveredBugs, ok := report["delivered_bugs"].(float64); ok {
			h.combinedDeliveredBugs += deliveredBugs
		}

		// Aggregate functions
		if functions, ok := report["functions"].([]map[string]any); ok {
			for _, funcData := range functions {
				if name, ok := funcData["name"].(string); ok {
					h.combinedFunctions[name] = funcData
				}
			}
		}
	}
}

func (h *HalsteadAggregator) GetResult() analyze.Report {
	// Convert functions map to slice for common formatter
	var functionDetails []map[string]interface{}
	for _, funcData := range h.combinedFunctions {
		if funcMap, ok := funcData.(map[string]interface{}); ok {
			functionDetails = append(functionDetails, funcMap)
		}
	}

	return analyze.Report{
		"analyzer_name":   "halstead",
		"total_functions": len(functionDetails),
		"functions":       functionDetails,
		"volume":          h.combinedVolume,
		"difficulty":      h.combinedDifficulty,
		"effort":          h.combinedEffort,
		"time_to_program": h.combinedTimeToProgram,
		"delivered_bugs":  h.combinedDeliveredBugs,
		"message":         "Halstead complexity analysis completed",
	}
}
