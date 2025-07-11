package uast

import (
	"fmt"
	"strings"
)

// ParseDSL parses a DSL query string and returns the root DSLNode AST.
// Returns an error for invalid syntax or unsupported constructs.
func ParseDSL(input string) (DSLNode, error) {
	parser := &QueryDSL{Buffer: input}
	if isParserInitFailed(parser.Init()) {
		return nil, createParserInitError(parser.Init())
	}
	if isParseFailed(parser.Parse()) {
		return nil, createParseError(parser.Parse())
	}
	ast := parser.tokens32.AST()
	return ConvertAST(ast, input), nil
}

func isParserInitFailed(err error) bool {
	return err != nil
}

func createParserInitError(err error) error {
	return fmt.Errorf("parser initialization failed: %w", err)
}

func isParseFailed(err error) bool {
	return err != nil
}

func createParseError(err error) error {
	errStr := err.Error()
	if isUnknownInputError(errStr) {
		return fmt.Errorf("parse error at 1:1: unknown input")
	}
	return fmt.Errorf("parse error: %w", err)
}

func isUnknownInputError(errStr string) bool {
	return strings.Contains(errStr, "parse error near Unknown")
}

// stringifyAST returns a compact, function-call style string representation of the legacy DSLNode AST.
func stringifyAST(n DSLNode) string {
	switch v := n.(type) {
	case *PipelineNode:
		return stringifyPipelineNode(v)
	case *MapNode:
		return stringifyMapNode(v)
	case *RMapNode:
		return stringifyRMapNode(v)
	case *FilterNode:
		return stringifyFilterNode(v)
	case *RFilterNode:
		return stringifyRFilterNode(v)
	case *ReduceNode:
		return stringifyReduceNode(v)
	case *FieldNode:
		return stringifyFieldNode(v)
	case *LiteralNode:
		return stringifyLiteralNode(v)
	case *CallNode:
		return stringifyCallNode(v)
	case string:
		return v
	case nil:
		return "<nil>"
	default:
		return "<unknown>"
	}
}

func stringifyPipelineNode(node *PipelineNode) string {
	var stages []string
	for _, s := range node.Stages {
		stages = append(stages, stringifyAST(s))
	}
	return "Pipeline(" + strings.Join(stages, " | ") + ")"
}

func stringifyMapNode(node *MapNode) string {
	return "Map(" + stringifyAST(node.Expr) + ")"
}

func stringifyRMapNode(node *RMapNode) string {
	return "RMap(" + stringifyAST(node.Expr) + ")"
}

func stringifyFilterNode(node *FilterNode) string {
	return "Filter(" + stringifyAST(node.Expr) + ")"
}

func stringifyRFilterNode(node *RFilterNode) string {
	return "RFilter(" + stringifyAST(node.Expr) + ")"
}

func stringifyReduceNode(node *ReduceNode) string {
	return "Reduce(" + stringifyAST(node.Expr) + ")"
}

func stringifyFieldNode(node *FieldNode) string {
	return "Field(" + node.Name + ")"
}

func stringifyLiteralNode(node *LiteralNode) string {
	return "Literal(" + stringifyAST(node.Value) + ")"
}

func stringifyCallNode(node *CallNode) string {
	if hasNoArgs(node) {
		return "Call(" + node.Name + ")"
	}
	return stringifyCallNodeWithArgs(node)
}

func hasNoArgs(node *CallNode) bool {
	return node.Args == nil || len(node.Args) == 0
}

func stringifyCallNodeWithArgs(node *CallNode) string {
	var args []string
	for _, a := range node.Args {
		args = append(args, stringifyAST(a))
	}
	return "Call(" + node.Name + ", " + strings.Join(args, ", ") + ")"
}

// ConvertAST converts a *node32 parse tree to the legacy DSLNode AST.
func ConvertAST(n *node32, buffer string) DSLNode {
	if isNilNode(n) {
		return nil
	}
	rule := rul3s[n.pegRule]
	return convertNodeByRule(n, rule, buffer)
}

func isNilNode(n *node32) bool {
	return n == nil
}

func isNilDSLNode(n DSLNode) bool {
	return n == nil
}

func convertNodeByRule(n *node32, rule string, buffer string) DSLNode {
	switch rule {
	case "Query", "Pipeline":
		return convertPipelineNode(n, buffer)
	case "Map":
		return convertMapNode(n, buffer)
	case "RMap":
		return convertRMapNode(n, buffer)
	case "Filter":
		return convertFilterNode(n, buffer)
	case "RFilter":
		return convertRFilterNode(n, buffer)
	case "Reduce":
		return convertReduceNode(n, buffer)
	case "FieldAccess":
		return convertFieldAccessNode(n, buffer)
	case "Literal":
		return convertLiteralNode(n, buffer)
	case "String":
		return convertStringNode(n, buffer)
	case "Number", "Boolean":
		return convertNumberBooleanNode(n, buffer)
	case "Comparison":
		return convertComparisonNode(n, buffer)
	case "OrExpr":
		return convertOrExprNode(n, buffer)
	case "AndExpr":
		return convertAndExprNode(n, buffer)
	case "Membership":
		return convertMembershipNode(n, buffer)
	default:
		return convertDefaultNode(n, buffer)
	}
}

func convertPipelineNode(n *node32, buffer string) DSLNode {
	stages := collectStages(n, buffer)
	if isSingleStage(stages) {
		return stages[0]
	}
	return &PipelineNode{Stages: stages}
}

func collectStages(n *node32, buffer string) []DSLNode {
	var stages []DSLNode
	for c := n.up; c != nil; c = c.next {
		stage := ConvertAST(c, buffer)
		if isNotNullStage(stage) {
			stages = append(stages, stage)
		}
	}
	return stages
}

func isNotNullStage(stage DSLNode) bool {
	return stage != nil
}

func isSingleStage(stages []DSLNode) bool {
	return len(stages) == 1
}

func convertMapNode(n *node32, buffer string) DSLNode {
	return &MapNode{Expr: ConvertAST(n.up, buffer)}
}

func convertRMapNode(n *node32, buffer string) DSLNode {
	return &RMapNode{Expr: ConvertAST(n.up, buffer)}
}

func convertFilterNode(n *node32, buffer string) DSLNode {
	return &FilterNode{Expr: ConvertAST(n.up, buffer)}
}

func convertRFilterNode(n *node32, buffer string) DSLNode {
	return &RFilterNode{Expr: ConvertAST(n.up, buffer)}
}

func convertReduceNode(n *node32, buffer string) DSLNode {
	if hasUpNode(n) {
		name := extractNodeName(n, buffer)
		return &ReduceNode{Expr: &CallNode{Name: name, Args: nil}}
	}
	return &ReduceNode{Expr: nil}
}

func hasUpNode(n *node32) bool {
	return n.up != nil
}

func extractNodeName(n *node32, buffer string) string {
	return buffer[n.up.begin:n.up.end]
}

func convertFieldAccessNode(n *node32, buffer string) DSLNode {
	for c := n.up; c != nil; c = c.next {
		if isIdentifierRule(c) {
			return &FieldNode{Name: buffer[c.begin:c.end]}
		}
	}
	return nil
}

func isIdentifierRule(c *node32) bool {
	return rul3s[c.pegRule] == "Identifier"
}

func convertLiteralNode(n *node32, buffer string) DSLNode {
	if hasUpNode(n) {
		val := ConvertAST(n.up, buffer)
		if isLiteralNode(val) {
			return val
		}
		return &LiteralNode{Value: val}
	}
	return nil
}

func isLiteralNode(val DSLNode) bool {
	_, ok := val.(*LiteralNode)
	return ok
}

func convertStringNode(n *node32, buffer string) DSLNode {
	val := buffer[n.begin:n.end]
	if isQuotedString(val) {
		val = removeQuotes(val)
	}
	return &LiteralNode{Value: val}
}

func isQuotedString(val string) bool {
	return len(val) >= 2 && (val[0] == '"' || val[0] == '\'')
}

func removeQuotes(val string) string {
	return val[1 : len(val)-1]
}

func convertNumberBooleanNode(n *node32, buffer string) DSLNode {
	return &LiteralNode{Value: buffer[n.begin:n.end]}
}

func convertComparisonNode(n *node32, buffer string) DSLNode {
	left, right, op := extractComparisonParts(n, buffer)
	left = wrapAsLiteralIfNeeded(left)
	right = wrapAsLiteralIfNeeded(right)
	return &CallNode{Name: op, Args: []DSLNode{left, right}}
}

func extractComparisonParts(n *node32, buffer string) (DSLNode, DSLNode, string) {
	var left, right DSLNode
	var op string
	valueCount := 0
	for c := n.up; c != nil; c = c.next {
		rule := rul3s[c.pegRule]
		switch rule {
		case "Value":
			if isFirstValue(valueCount) {
				left = ConvertAST(c, buffer)
				valueCount++
			} else if isSecondValue(valueCount) {
				right = ConvertAST(c, buffer)
				valueCount++
			}
		case "CompOp":
			op = buffer[c.begin:c.end]
		}
	}
	return left, right, op
}

func isFirstValue(valueCount int) bool {
	return valueCount == 0
}

func isSecondValue(valueCount int) bool {
	return valueCount == 1
}

func wrapAsLiteralIfNeeded(node DSLNode) DSLNode {
	if isStringNode(node) {
		return &LiteralNode{Value: node}
	}
	return node
}

func isStringNode(node DSLNode) bool {
	_, ok := node.(string)
	return ok
}

func convertOrExprNode(n *node32, buffer string) DSLNode {
	args := collectOrExprArgs(n, buffer)
	if isSingleArg(args) {
		return args[0]
	}
	return foldOrExprArgs(args)
}

func collectOrExprArgs(n *node32, buffer string) []DSLNode {
	var args []DSLNode
	for c := n.up; c != nil; c = c.next {
		child := ConvertAST(c, buffer)
		if isNotNullStage(child) {
			args = append(args, child)
		}
	}
	return args
}

func isSingleArg(args []DSLNode) bool {
	return len(args) == 1
}

func foldOrExprArgs(args []DSLNode) DSLNode {
	cur := args[0]
	for i := 1; i < len(args); i++ {
		cur = &CallNode{Name: "||", Args: []DSLNode{cur, args[i]}}
	}
	return cur
}

func convertAndExprNode(n *node32, buffer string) DSLNode {
	args := collectAndExprArgs(n, buffer)
	if isSingleArg(args) {
		return args[0]
	}
	return foldAndExprArgs(args)
}

func collectAndExprArgs(n *node32, buffer string) []DSLNode {
	var args []DSLNode
	for c := n.up; c != nil; c = c.next {
		child := ConvertAST(c, buffer)
		if isNotNullStage(child) {
			args = append(args, child)
		}
	}
	return args
}

func foldAndExprArgs(args []DSLNode) DSLNode {
	cur := args[0]
	for i := 1; i < len(args); i++ {
		cur = &CallNode{Name: "&&", Args: []DSLNode{cur, args[i]}}
	}
	return cur
}

func convertMembershipNode(n *node32, buffer string) DSLNode {
	left, right := extractMembershipParts(n, buffer)
	if isIncompleteMembership(left, right) {
		return nil
	}
	return &CallNode{Name: "has", Args: []DSLNode{left, right}}
}

func extractMembershipParts(n *node32, buffer string) (DSLNode, DSLNode) {
	var left, right DSLNode
	for c := n.up; c != nil; c = c.next {
		rule := rul3s[c.pegRule]
		if isFieldAccessRule(rule) && isNilDSLNode(left) {
			left = ConvertAST(c, buffer)
		} else if isValueRule(rule) && isNilDSLNode(right) {
			right = ConvertAST(c, buffer)
		}
	}
	return left, right
}

func isFieldAccessRule(rule string) bool {
	return rule == "FieldAccess"
}

func isValueRule(rule string) bool {
	return rule == "Value"
}

func isIncompleteMembership(left, right DSLNode) bool {
	return left == nil || right == nil
}

func convertDefaultNode(n *node32, buffer string) DSLNode {
	if hasUpNode(n) {
		return ConvertAST(n.up, buffer)
	}
	return nil
}
