package uast

import (
	"fmt"
	"strings"
)

// ParseDSL parses a DSL string into a legacy DSLNode AST using the generated QueryDSL parser.
func ParseDSL(input string) (DSLNode, error) {
	parser := &QueryDSL{Buffer: input}
	if err := parser.Init(); err != nil {
		return nil, fmt.Errorf("parser initialization failed: %w", err)
	}
	if err := parser.Parse(); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "parse error near Unknown") {
			return nil, fmt.Errorf("parse error at 1:1: unknown input")
		}
		return nil, fmt.Errorf("parse error: %w", err)
	}
	ast := parser.tokens32.AST()
	return ConvertAST(ast, input), nil
}

// stringifyAST returns a compact, function-call style string representation of the legacy DSLNode AST.
func stringifyAST(n DSLNode) string {
	switch v := n.(type) {
	case *PipelineNode:
		var stages []string
		for _, s := range v.Stages {
			stages = append(stages, stringifyAST(s))
		}
		return "Pipeline(" + strings.Join(stages, " | ") + ")"
	case *MapNode:
		return "Map(" + stringifyAST(v.Expr) + ")"
	case *FilterNode:
		return "Filter(" + stringifyAST(v.Expr) + ")"
	case *ReduceNode:
		return "Reduce(" + stringifyAST(v.Expr) + ")"
	case *FieldNode:
		return "Field(" + v.Name + ")"
	case *LiteralNode:
		return "Literal(" + stringifyAST(v.Value) + ")"
	case *CallNode:
		if v.Args == nil || len(v.Args) == 0 {
			return "Call(" + v.Name + ")"
		}
		var args []string
		for _, a := range v.Args {
			args = append(args, stringifyAST(a))
		}
		return "Call(" + v.Name + ", " + strings.Join(args, ", ") + ")"
	case string:
		return v
	case nil:
		return "<nil>"
	default:
		return "<unknown>"
	}
}

// ConvertAST converts a *node32 parse tree to the legacy DSLNode AST.
func ConvertAST(n *node32, buffer string) DSLNode {
	if n == nil {
		return nil
	}
	rule := rul3s[n.pegRule]
	switch rule {
	case "Query", "Pipeline":
		var stages []DSLNode
		for c := n.up; c != nil; c = c.next {
			stage := ConvertAST(c, buffer)
			if stage != nil {
				stages = append(stages, stage)
			}
		}
		if len(stages) == 1 {
			return stages[0]
		}
		return &PipelineNode{Stages: stages}
	case "Map":
		return &MapNode{Expr: ConvertAST(n.up, buffer)}
	case "Filter":
		return &FilterNode{Expr: ConvertAST(n.up, buffer)}
	case "Reduce":
		if n.up != nil {
			name := buffer[n.up.begin:n.up.end]
			return &ReduceNode{Expr: &CallNode{Name: name, Args: nil}}
		}
		return &ReduceNode{Expr: nil}
	case "FieldAccess":
		for c := n.up; c != nil; c = c.next {
			if rul3s[c.pegRule] == "Identifier" {
				return &FieldNode{Name: buffer[c.begin:c.end]}
			}
		}
		return nil
	case "Literal":
		if n.up != nil {
			val := ConvertAST(n.up, buffer)
			if _, ok := val.(*LiteralNode); ok {
				return val
			}
			return &LiteralNode{Value: val}
		}
		return nil
	case "String":
		val := buffer[n.begin:n.end]
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') {
			val = val[1 : len(val)-1]
		}
		return &LiteralNode{Value: val}
	case "Number", "Boolean":
		return &LiteralNode{Value: buffer[n.begin:n.end]}
	case "Comparison":
		// Comparison(left, op, right) -> CallNode{Name: op, Args: [left, right]}
		var left, right DSLNode
		var op string
		valueCount := 0
		for c := n.up; c != nil; c = c.next {
			rule := rul3s[c.pegRule]
			switch rule {
			case "Value":
				if valueCount == 0 {
					left = ConvertAST(c, buffer)
					valueCount++
				} else if valueCount == 1 {
					right = ConvertAST(c, buffer)
					valueCount++
				}
			case "CompOp":
				op = buffer[c.begin:c.end]
			}
		}
		// Wrap left/right as LiteralNode if they are string/number/boolean
		if s, ok := left.(string); ok {
			left = &LiteralNode{Value: s}
		}
		if s, ok := right.(string); ok {
			right = &LiteralNode{Value: s}
		}
		return &CallNode{Name: op, Args: []DSLNode{left, right}}
	case "OrExpr":
		// Fold OrExpr: AndExpr ( '||' AndExpr )*
		var args []DSLNode
		for c := n.up; c != nil; c = c.next {
			child := ConvertAST(c, buffer)
			if child != nil {
				args = append(args, child)
			}
		}
		if len(args) == 1 {
			return args[0]
		}
		// Left-associative folding
		cur := args[0]
		for i := 1; i < len(args); i++ {
			cur = &CallNode{Name: "||", Args: []DSLNode{cur, args[i]}}
		}
		return cur
	case "AndExpr":
		// Fold AndExpr: NotExpr ( '&&' NotExpr )*
		var args []DSLNode
		for c := n.up; c != nil; c = c.next {
			child := ConvertAST(c, buffer)
			if child != nil {
				args = append(args, child)
			}
		}
		if len(args) == 1 {
			return args[0]
		}
		cur := args[0]
		for i := 1; i < len(args); i++ {
			cur = &CallNode{Name: "&&", Args: []DSLNode{cur, args[i]}}
		}
		return cur
	case "Membership":
		// Membership(left, right) -> CallNode{Name: "has", Args: [left, right]}
		var left, right DSLNode
		for c := n.up; c != nil; c = c.next {
			rule := rul3s[c.pegRule]
			if rule == "FieldAccess" && left == nil {
				left = ConvertAST(c, buffer)
			} else if rule == "Value" && right == nil {
				right = ConvertAST(c, buffer)
			}
		}
		if left == nil || right == nil {
			return nil
		}
		return &CallNode{Name: "has", Args: []DSLNode{left, right}}
	default:
		if n.up != nil {
			return ConvertAST(n.up, buffer)
		}
		return nil
	}
}
