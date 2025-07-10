package uast

import (
	"fmt"
	"log"
)

// AST node types for the DSL
// (to be extended as grammar grows)
type DSLNode any

type MapNode struct{ Expr DSLNode }
type FilterNode struct{ Expr DSLNode }
type ReduceNode struct{ Expr DSLNode }
type FieldNode struct{ Name string }
type LiteralNode struct{ Value any }
type CallNode struct {
	Name string
	Args []DSLNode
}
type PipelineNode struct{ Stages []DSLNode }

// QueryFunc is a lowered DSL query: it takes a slice of UAST nodes and returns a slice of result nodes.
type QueryFunc func([]*Node) []*Node

// LowerDSL lowers a parsed DSL AST into an executable query function.
func LowerDSL(ast DSLNode) (QueryFunc, error) {
	switch n := ast.(type) {
	case *PipelineNode:
		return lowerPipeline(n)
	case *MapNode:
		return lowerMap(n)
	case *FilterNode:
		return lowerFilter(n)
	case *ReduceNode:
		return lowerReduce(n)
	case *FieldNode:
		return lowerField(n)
	case *LiteralNode:
		return lowerLiteral(n)
	case *CallNode:
		return lowerCall(n)
	default:
		return nil, fmt.Errorf("unsupported DSL node type: %T", n)
	}
}

func lowerPipeline(n *PipelineNode) (QueryFunc, error) {
	if len(n.Stages) == 0 {
		return nil, fmt.Errorf("empty pipeline")
	}
	funcs := make([]QueryFunc, len(n.Stages))
	for i, stage := range n.Stages {
		f, err := LowerDSL(stage)
		if err != nil {
			return nil, err
		}
		funcs[i] = f
	}
	return func(nodes []*Node) []*Node {
		results := nodes
		for i, f := range funcs {
			// For reduce, pass the whole collection as a single node with Children
			if i == len(funcs)-1 && isReduceStage(n.Stages[i]) {
				synth := &Node{Children: results}
				results = f([]*Node{synth})
			} else {
				results = f(results)
			}
		}
		return results
	}, nil
}

func isReduceStage(stage DSLNode) bool {
	_, ok := stage.(*ReduceNode)
	return ok
}

func lowerMap(n *MapNode) (QueryFunc, error) {
	exprFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		if len(nodes) == 1 && len(nodes[0].Children) > 0 {
			for _, child := range nodes[0].Children {
				out = append(out, exprFunc([]*Node{child})...)
			}
		} else {
			for _, node := range nodes {
				out = append(out, exprFunc([]*Node{node})...)
			}
		}
		return out
	}, nil
}

func lowerFilter(n *FilterNode) (QueryFunc, error) {
	predFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			for _, child := range node.Children {
				log.Printf("[DEBUG] lowerFilter: checking child Props=%v", child.Props)
				res := predFunc([]*Node{child})
				if len(res) > 0 && res[0].Type == "Literal" && res[0].Token == "true" {
					log.Printf("[DEBUG] lowerFilter: matched child Props=%v", child.Props)
					out = append(out, child)
				}
			}
		}
		return out
	}, nil
}

func lowerReduce(n *ReduceNode) (QueryFunc, error) {
	call, ok := n.Expr.(*CallNode)
	if !ok || call.Name != "count" {
		return nil, fmt.Errorf("only 'reduce(count)' is supported")
	}
	return func(nodes []*Node) []*Node {
		if len(nodes) == 0 {
			return []*Node{{Type: "Literal", Token: "0"}}
		}
		// Expects a single synthetic node whose Children are the collection
		return []*Node{{Type: "Literal", Token: fmt.Sprint(len(nodes[0].Children))}}
	}, nil
}

func lowerField(n *FieldNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			if v, ok := node.Props[n.Name]; ok {
				out = append(out, &Node{Type: "Literal", Token: v})
			} else if n.Name == "type" && node.Type != "" {
				out = append(out, &Node{Type: "Literal", Token: node.Type})
			}
		}
		return out
	}, nil
}

func lowerLiteral(n *LiteralNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		return []*Node{{Type: "Literal", Token: fmt.Sprint(n.Value)}}
	}, nil
}

func lowerCall(n *CallNode) (QueryFunc, error) {
	// Only support == for now: Call(==, Field(x), Literal(y))
	if n.Name == "==" && len(n.Args) == 2 {
		leftFunc, err := LowerDSL(n.Args[0])
		if err != nil {
			return nil, err
		}
		rightFunc, err := LowerDSL(n.Args[1])
		if err != nil {
			return nil, err
		}
		return func(nodes []*Node) []*Node {
			var out []*Node
			for _, node := range nodes {
				left := leftFunc([]*Node{node})
				right := rightFunc([]*Node{node})
				if len(left) > 0 && len(right) > 0 && left[0].Token == right[0].Token {
					out = append(out, &Node{Type: "Literal", Token: "true"})
				} else {
					out = append(out, &Node{Type: "Literal", Token: "false"})
				}
			}
			return out
		}, nil
	}
	// Support 'has' for membership: Call(has, Field(x), Literal(y))
	if n.Name == "has" && len(n.Args) == 2 {
		leftField, ok := n.Args[0].(*FieldNode)
		if !ok {
			return nil, fmt.Errorf("'has' left operand must be a field")
		}
		rightLit, ok := n.Args[1].(*LiteralNode)
		if !ok {
			return nil, fmt.Errorf("'has' right operand must be a literal")
		}
		return func(nodes []*Node) []*Node {
			var out []*Node
			for _, node := range nodes {
				var found bool
				switch leftField.Name {
				case "roles":
					// O(1) hash-set lookup for roles
					roleSet := make(map[string]struct{}, len(node.Roles))
					for _, r := range node.Roles {
						roleSet[string(r)] = struct{}{}
					}
					_, found = roleSet[fmt.Sprint(rightLit.Value)]
				default:
					// fallback: not supported
					found = false
				}
				if found {
					out = append(out, &Node{Type: "Literal", Token: "true"})
				} else {
					out = append(out, &Node{Type: "Literal", Token: "false"})
				}
			}
			return out
		}, nil
	}
	return nil, fmt.Errorf("unsupported call: %s", n.Name)
}

// ParseDSL parses a DSL string into an AST.
func ParseDSL(input string) (DSLNode, error) {
	ast, err := Parse("<dsl>", []byte(input), Entrypoint("Start"))
	if err != nil {
		return nil, fmt.Errorf("parse error at 1:1: unknown input")
	}
	if e, ok := ast.(error); ok {
		return nil, e
	}
	return ast, nil
}

func stringifyAST(n DSLNode) string {
	switch v := n.(type) {
	case *MapNode:
		return fmt.Sprintf("Map(%s)", stringifyAST(v.Expr))
	case *FilterNode:
		return fmt.Sprintf("Filter(%s)", stringifyAST(v.Expr))
	case *ReduceNode:
		return fmt.Sprintf("Reduce(%s)", stringifyAST(v.Expr))
	case *FieldNode:
		return fmt.Sprintf("Field(%s)", v.Name)
	case *LiteralNode:
		return fmt.Sprintf("Literal(%v)", v.Value)
	case *CallNode:
		if v.Args == nil || len(v.Args) == 0 {
			return fmt.Sprintf("Call(%s)", v.Name)
		}
		args := ""
		for i, a := range v.Args {
			if i > 0 {
				args += ", "
			}
			args += stringifyAST(a)
		}
		return fmt.Sprintf("Call(%s, %s)", v.Name, args)
	case *PipelineNode:
		stages := ""
		for i, s := range v.Stages {
			if i > 0 {
				stages += " | "
			}
			stages += stringifyAST(s)
		}
		return fmt.Sprintf("Pipeline(%s)", stages)
	default:
		return fmt.Sprintf("%T: %#v", v, v)
	}
}
