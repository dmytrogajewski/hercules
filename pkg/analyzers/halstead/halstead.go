package halstead

import (
	"encoding/json"
	"io"
	"math"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/fatih/color"
)

// HalsteadAnalyzer provides Halstead complexity measures analysis
type HalsteadAnalyzer struct{}

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
	return "halstead_analysis"
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

// FormatReport formats Halstead analysis results as human-readable text
func (h *HalsteadAnalyzer) FormatReport(report analyze.Report, writer io.Writer) error {
	colorWriter := color.New(color.FgBlue)

	// Get thresholds for this analyzer
	thresholds := h.Thresholds()

	// Output threshold information
	colorWriter.Fprintf(writer, "Halstead Complexity Measures:\n")
	colorWriter.Fprintf(writer, "------------------------------\n\n")

	// Volume thresholds
	if volumeThresholds, ok := thresholds["volume"]; ok {
		green, _ := volumeThresholds["green"].(int)
		yellow, _ := volumeThresholds["yellow"].(int)
		red, _ := volumeThresholds["red"].(int)
		colorWriter.Fprintf(writer, "Volume Thresholds:\n")
		colorWriter.Fprintf(writer, "  Green (Good): ≤ %d\n", green)
		colorWriter.Fprintf(writer, "  Yellow (Warning): %d to %d\n", green+1, yellow)
		colorWriter.Fprintf(writer, "  Red (High): > %d\n\n", red)
	}

	// Difficulty thresholds
	if difficultyThresholds, ok := thresholds["difficulty"]; ok {
		green, _ := difficultyThresholds["green"].(int)
		yellow, _ := difficultyThresholds["yellow"].(int)
		red, _ := difficultyThresholds["red"].(int)
		colorWriter.Fprintf(writer, "Difficulty Thresholds:\n")
		colorWriter.Fprintf(writer, "  Green (Good): ≤ %d\n", green)
		colorWriter.Fprintf(writer, "  Yellow (Warning): %d to %d\n", green+1, yellow)
		colorWriter.Fprintf(writer, "  Red (High): > %d\n\n", red)
	}

	// Effort thresholds
	if effortThresholds, ok := thresholds["effort"]; ok {
		green, _ := effortThresholds["green"].(int)
		yellow, _ := effortThresholds["yellow"].(int)
		red, _ := effortThresholds["red"].(int)
		colorWriter.Fprintf(writer, "Effort Thresholds:\n")
		colorWriter.Fprintf(writer, "  Green (Good): ≤ %d\n", green)
		colorWriter.Fprintf(writer, "  Yellow (Warning): %d to %d\n", green+1, yellow)
		colorWriter.Fprintf(writer, "  Red (High): > %d\n\n", red)
	}

	// Output metrics
	if fileMetrics, ok := report["file_metrics"].(map[string]any); ok {
		colorWriter.Fprintf(writer, "File-Level Metrics:\n")
		if volume, ok := fileMetrics["volume"].(float64); ok {
			colorWriter.Fprintf(writer, "  Volume: %.2f\n", volume)
		}
		if difficulty, ok := fileMetrics["difficulty"].(float64); ok {
			colorWriter.Fprintf(writer, "  Difficulty: %.2f\n", difficulty)
		}
		if effort, ok := fileMetrics["effort"].(float64); ok {
			colorWriter.Fprintf(writer, "  Effort: %.2f\n", effort)
		}
		if timeToProgram, ok := fileMetrics["time_to_program"].(float64); ok {
			colorWriter.Fprintf(writer, "  Time to Program: %.2f seconds\n", timeToProgram)
		}
		if deliveredBugs, ok := fileMetrics["delivered_bugs"].(float64); ok {
			colorWriter.Fprintf(writer, "  Delivered Bugs: %.2f\n", deliveredBugs)
		}
	}

	// Output function-level metrics
	if functions, ok := report["functions"].(map[string]any); ok {
		if len(functions) > 0 {
			colorWriter.Fprintf(writer, "\nFunction-Level Metrics:\n")
			for funcName, funcData := range functions {
				if funcMetrics, ok := funcData.(map[string]any); ok {
					colorWriter.Fprintf(writer, "  %s:\n", funcName)
					if volume, ok := funcMetrics["volume"].(float64); ok {
						colorWriter.Fprintf(writer, "    Volume: %.2f\n", volume)
					}
					if difficulty, ok := funcMetrics["difficulty"].(float64); ok {
						colorWriter.Fprintf(writer, "    Difficulty: %.2f\n", difficulty)
					}
					if effort, ok := funcMetrics["effort"].(float64); ok {
						colorWriter.Fprintf(writer, "    Effort: %.2f\n", effort)
					}
					if timeToProgram, ok := funcMetrics["time_to_program"].(float64); ok {
						colorWriter.Fprintf(writer, "    Time to Program: %.2f seconds\n", timeToProgram)
					}
					if deliveredBugs, ok := funcMetrics["delivered_bugs"].(float64); ok {
						colorWriter.Fprintf(writer, "    Delivered Bugs: %.2f\n", deliveredBugs)
					}
				}
			}
		}
	}

	return nil
}

// FormatReportJSON formats Halstead analysis results as JSON
func (h *HalsteadAnalyzer) FormatReportJSON(report analyze.Report, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// DefaultConfig returns default Halstead analysis configuration
func (h *HalsteadAnalyzer) DefaultConfig() HalsteadConfig {
	return HalsteadConfig{
		IncludeFunctionBreakdown: true,
		IncludeTimeEstimate:      true,
		IncludeBugEstimate:       true,
	}
}

func (h *HalsteadAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
	return h.AnalyzeWithConfig(root, h.DefaultConfig())
}

func (h *HalsteadAnalyzer) AnalyzeWithConfig(root *node.Node, config HalsteadConfig) (analyze.Report, error) {
	if root == nil {
		return h.emptyResult(), nil
	}

	// Find all functions in the code
	functions := h.findFunctions(root)

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

	// Build result
	result := analyze.Report{
		"file_count": len(functions),
		"functions":  functionMetrics,
		"file_metrics": map[string]any{
			"distinct_operators": fileMetrics.DistinctOperators,
			"distinct_operands":  fileMetrics.DistinctOperands,
			"total_operators":    fileMetrics.TotalOperators,
			"total_operands":     fileMetrics.TotalOperands,
			"vocabulary":         fileMetrics.Vocabulary,
			"length":             fileMetrics.Length,
			"estimated_length":   fileMetrics.EstimatedLength,
			"volume":             fileMetrics.Volume,
			"difficulty":         fileMetrics.Difficulty,
			"effort":             fileMetrics.Effort,
			"time_to_program":    fileMetrics.TimeToProgram,
			"delivered_bugs":     fileMetrics.DeliveredBugs,
		},
	}

	return result, nil
}

// findFunctions finds all function nodes in the AST
func (h *HalsteadAnalyzer) findFunctions(root *node.Node) []*node.Node {
	var functions []*node.Node

	if root == nil {
		return functions
	}

	// Check if current node is a function
	if h.isFunctionNode(root) {
		functions = append(functions, root)
	}

	// Recursively search children
	for _, child := range root.Children {
		functions = append(functions, h.findFunctions(child)...)
	}

	return functions
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
		operators[operator]++
	} else if h.isOperand(node) {
		operand := h.getOperandName(node)
		operands[operand]++
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

// isFunctionNode determines if a node represents a function definition
func (h *HalsteadAnalyzer) isFunctionNode(n *node.Node) bool {
	if n == nil {
		return false
	}

	switch n.Type {
	case node.UASTFunction, node.UASTFunctionDecl, node.UASTMethod,
		node.UASTLambda:
		return true
	}

	for _, role := range n.Roles {
		switch role {
		case node.RoleFunction, node.RoleDeclaration:
			return true
		}
	}

	return false
}

// getOperatorName extracts the operator name from a node
func (h *HalsteadAnalyzer) getOperatorName(node *node.Node) string {
	if node == nil {
		return ""
	}

	// Try to get operator from properties
	if op, ok := node.Props["operator"]; ok {
		return op
	}

	// Fall back to type
	return node.Type
}

// getOperandName extracts the operand name from a node
func (h *HalsteadAnalyzer) getOperandName(node *node.Node) string {
	if node == nil {
		return ""
	}

	// Try to get name from properties
	if name, ok := node.Props["name"]; ok {
		return name
	}

	if value, ok := node.Props["value"]; ok {
		return value
	}

	// Fall back to type
	return node.Type
}

// extractFunctionName extracts the function name from a function node
func (h *HalsteadAnalyzer) extractFunctionName(n *node.Node) string {
	if n == nil {
		return ""
	}

	// Try to get function name from properties
	if name, ok := n.Props["name"]; ok {
		return name
	}

	// Look for identifier child with name role
	for _, child := range n.Children {
		if child.Type == node.UASTIdentifier {
			for _, role := range child.Roles {
				if role == node.RoleName {
					if name, ok := child.Props["name"]; ok {
						return name
					}
				}
			}
		}
	}

	return ""
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

// emptyResult returns an empty result structure
func (h *HalsteadAnalyzer) emptyResult() analyze.Report {
	return analyze.Report{
		"file_count": 0,
		"functions":  make(map[string]*FunctionHalsteadMetrics),
		"file_metrics": map[string]any{
			"distinct_operators": 0,
			"distinct_operands":  0,
			"total_operators":    0,
			"total_operands":     0,
			"vocabulary":         0,
			"length":             0,
			"estimated_length":   0.0,
			"volume":             0.0,
			"difficulty":         0.0,
			"effort":             0.0,
			"time_to_program":    0.0,
			"delivered_bugs":     0.0,
		},
	}
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
	if report, ok := results["halstead_analysis"]; ok {
		if fileMetrics, ok := report["file_metrics"].(map[string]any); ok {
			if volume, ok := fileMetrics["volume"].(float64); ok {
				h.combinedVolume += volume
			}
			if difficulty, ok := fileMetrics["difficulty"].(float64); ok {
				h.combinedDifficulty += difficulty
			}
			if effort, ok := fileMetrics["effort"].(float64); ok {
				h.combinedEffort += effort
			}
			if timeToProgram, ok := fileMetrics["time_to_program"].(float64); ok {
				h.combinedTimeToProgram += timeToProgram
			}
			if deliveredBugs, ok := fileMetrics["delivered_bugs"].(float64); ok {
				h.combinedDeliveredBugs += deliveredBugs
			}
		}
		if halsteadFunctions, ok := report["functions"].(map[string]any); ok {
			for funcName, funcData := range halsteadFunctions {
				h.combinedFunctions[funcName] = funcData
			}
		}
	}
}

func (h *HalsteadAggregator) GetResult() analyze.Report {
	return analyze.Report{
		"file_metrics": map[string]any{
			"volume":          h.combinedVolume,
			"difficulty":      h.combinedDifficulty,
			"effort":          h.combinedEffort,
			"time_to_program": h.combinedTimeToProgram,
			"delivered_bugs":  h.combinedDeliveredBugs,
		},
		"functions": h.combinedFunctions,
	}
}
