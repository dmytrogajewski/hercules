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
				log.Printf("[DEBUG] lowerFilter: checking child Type=%s Props=%v", child.Type, child.Props)
				res := predFunc([]*Node{child})
				if len(res) > 0 {
					log.Printf("[DEBUG] lowerFilter: predicate result for child Type=%s: %v", child.Type, res[0].Token)
				}
				if len(res) > 0 && res[0].Type == "Literal" && res[0].Token == "true" {
					log.Printf("[DEBUG] lowerFilter: matched child Type=%s Props=%v", child.Type, child.Props)
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
			if n.Name == "roles" {
				for _, r := range node.Roles {
					out = append(out, &Node{Type: "Literal", Token: string(r)})
				}
			} else if v, ok := node.Props[n.Name]; ok {
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
	// Logical OR (||)
	if n.Name == "||" && len(n.Args) == 2 {
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
				l := leftFunc([]*Node{node})
				r := rightFunc([]*Node{node})
				lTrue := len(l) > 0 && l[0].Type == "Literal" && l[0].Token == "true"
				rTrue := len(r) > 0 && r[0].Type == "Literal" && r[0].Token == "true"
				if lTrue || rTrue {
					out = append(out, &Node{Type: "Literal", Token: "true"})
				} else {
					out = append(out, &Node{Type: "Literal", Token: "false"})
				}
			}
			return out
		}, nil
	}
	// Logical AND (&&)
	if n.Name == "&&" && len(n.Args) == 2 {
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
				l := leftFunc([]*Node{node})
				r := rightFunc([]*Node{node})
				lTrue := len(l) > 0 && l[0].Type == "Literal" && l[0].Token == "true"
				rTrue := len(r) > 0 && r[0].Type == "Literal" && r[0].Token == "true"
				var lTokens, rTokens []string
				for _, n := range l {
					lTokens = append(lTokens, n.Token)
				}
				for _, n := range r {
					rTokens = append(rTokens, n.Token)
				}
				log.Printf("[DEBUG] lowerCall &&: node.Type=%s lTrue=%v rTrue=%v lTokens=%v rTokens=%v", node.Type, lTrue, rTrue, lTokens, rTokens)
				if lTrue && rTrue {
					out = append(out, &Node{Type: "Literal", Token: "true"})
				} else {
					out = append(out, &Node{Type: "Literal", Token: "false"})
				}
			}
			return out
		}, nil
	}
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
				var lval, rval string
				if len(left) > 0 {
					lval = left[0].Token
				}
				if len(right) > 0 {
					rval = right[0].Token
				}
				log.Printf("[DEBUG] lowerCall ==: node.Type=%s left=%q right=%q", node.Type, lval, rval)
				if len(left) > 0 && len(right) > 0 && lval == rval {
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
				leftVals := leftFunc([]*Node{node})
				rightVals := rightFunc(nil)
				if len(leftVals) == 0 || len(rightVals) == 0 {
					log.Printf("[DEBUG] lowerCall has: leftVals or rightVals is empty for node.Type=%s", node.Type)
					out = append(out, &Node{Type: "Literal", Token: "false"})
					continue
				}
				rval := rightVals[0].Token
				found := false
				for _, l := range leftVals {
					if l.Token == rval {
						found = true
						break
					}
				}
				log.Printf("[DEBUG] lowerCall has: node.Token=%q leftVals=%v rightVals=%v found=%v", node.Token, leftVals, rightVals, found)
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
