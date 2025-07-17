package mapping

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MappingParser parses the mapping DSL and returns validated mapping rules.
type MappingParser struct{}

// ParseMapping parses the mapping DSL input and returns mapping rules.
func (p *MappingParser) ParseMapping(input string) ([]MappingRule, error) {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	ast, err := parseMappingDSL(input)
	if err != nil {
		fmt.Println("DEBUG: DSL input:")
		fmt.Println(input)
		fmt.Printf("DEBUG: parse error: %v\n", err)
		return nil, err
	}
	rules, err := buildMappingRulesFromAST(ast)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// parseMappingDSL uses the generated PEG parser to parse the input DSL.
func parseMappingDSL(input string) (any, error) {
	nodeTextBuffer = input
	parser := &MappingDSL{Buffer: input}
	parser.Init()
	if err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("mapping DSL parse error: %w", err)
	}
	return parser.AST(), nil
}

// buildMappingRulesFromAST converts the PEG AST to []MappingRule.
func buildMappingRulesFromAST(ast interface{}) ([]MappingRule, error) {
	var rules []MappingRule

	var walk func(n *node32)
	walk = func(n *node32) {
		if n == nil {
			return
		}
		if n.pegRule == ruleRule {
			rule, err := extractMappingRule(n)

			rules = append(rules, rule)
			if err != nil {
				//
			}
		}
		for child := n.up; child != nil; child = child.next {
			walk(child)
		}
	}

	switch n := ast.(type) {
	case *node32:
		walk(n)
	case []*node32:
		for _, child := range n {
			walk(child)
		}
	default:
		fmt.Printf("DEBUG: AST root type: %T, value: %#v\n", ast, ast)
		return nil, fmt.Errorf("invalid AST root type: %T", ast)
	}

	if len(rules) == 0 {
		if n, ok := ast.(*node32); ok && n != nil {
			fmt.Println("DEBUG: AST structure:")
			n.Print(os.Stdout, nodeTextBuffer)
		}
		return nil, fmt.Errorf("no mapping rules found in DSL")
	}
	return rules, nil
}

func extractMappingRule(ruleNode *node32) (MappingRule, error) {
	var rule MappingRule
	nameNode := findChild(ruleNode, ruleIdentifier)
	patternNode := findChild(ruleNode, rulePattern)
	uastSpecNode := findChild(ruleNode, ruleUASTSpec)
	whenNode := findChild(ruleNode, ruleConditionList)

	var inheritanceNode *node32
	for child := ruleNode.up; child != nil; child = child.next {
		if child.pegRule == ruleInheritanceComment {
			inheritanceNode = child
			break
		}
	}

	if inheritanceNode == nil && ruleNode.next != nil && ruleNode.next.pegRule == ruleInheritanceComment {
		inheritanceNode = ruleNode.next
	}

	extends := ""
	inheritanceConditions := []Condition{}
	if inheritanceNode != nil {
		extends, inheritanceConditions = extractInheritanceAndConditions(inheritanceNode)
	}

	if nameNode != nil {
		rule.Name = extractText(nameNode)
	}
	if patternNode != nil {
		rule.Pattern = extractPattern(patternNode)
	}
	if uastSpecNode != nil {
		spec, err := extractUASTSpec(uastSpecNode)
		if err == nil {
			rule.UASTSpec = spec
		}
	}
	allConditions := []Condition{}
	if whenNode != nil {
		allConditions = append(allConditions, extractConditions(whenNode)...)
	}
	if len(inheritanceConditions) > 0 {
		allConditions = append(allConditions, inheritanceConditions...)
	}
	rule.Conditions = allConditions
	rule.Extends = extends

	broken := rule.Name == "" || rule.Pattern == "" || rule.UASTSpec.Type == ""
	if broken {
		return rule, fmt.Errorf("invalid mapping rule")
	}
	return rule, nil
}

func extractConditions(condNode *node32) []Condition {
	var conds []Condition
	for c := condNode.up; c != nil; c = c.next {
		if c.pegRule == ruleCondition {
			cond := Condition{Expr: extractText(c)}
			conds = append(conds, cond)
		}
	}
	return conds
}

func extractInheritanceAndConditions(inheritanceNode *node32) (string, []Condition) {
	// Format: # Extends base_rule [when field == "val" and other_field != "bad"]
	text := extractText(inheritanceNode)
	base := ""
	conds := []Condition{}
	if strings.HasPrefix(strings.TrimSpace(text), "# Extends ") {
		parts := strings.Fields(strings.TrimSpace(text))
		if len(parts) >= 3 {
			base = parts[2]
		}
		// Look for 'when' and condition expressions
		whenIdx := strings.Index(text, "when ")
		if whenIdx != -1 {
			condExpr := strings.TrimSpace(text[whenIdx+5:])
			if condExpr != "" {
				// Split on 'and' for multiple conditions
				for _, c := range strings.Split(condExpr, " and ") {
					c = strings.TrimSpace(c)
					if c != "" {
						conds = append(conds, Condition{Expr: c})
					}
				}
			}
		}
	}
	return base, conds
}

func findChild(node *node32, target pegRule) *node32 {
	for child := node.up; child != nil; child = child.next {
		if child.pegRule == target {
			return child
		}
	}
	return nil
}

func extractText(node *node32) string {
	if node == nil {
		return ""
	}
	return string([]rune(nodeTextBuffer)[node.begin:node.end])
}

// nodeTextBuffer is set by parseMappingDSL for text extraction
var nodeTextBuffer string

func extractPattern(patternNode *node32) string {
	return extractText(patternNode)
}

func extractUASTField(fieldNode *node32) (string, []string) {
	var fname string
	var fvals []string

	for child := fieldNode.up; child != nil; child = child.next {
		switch child.pegRule {
		case ruleUASTFieldName:
			fname = extractText(child)
		case ruleUASTFieldValue:
			for valNode := child.up; valNode != nil; valNode = valNode.next {
				switch valNode.pegRule {
				case ruleCapture, ruleIdentifier:
					val := extractText(valNode)
					fvals = append(fvals, val)
				case ruleString:
					val := extractText(valNode)
					if unq, err := strconv.Unquote(val); err == nil {
						val = unq
					}
					fvals = append(fvals, val)
				case ruleMultipleCaptures:
					for valNode := valNode.up; valNode != nil; valNode = valNode.next {
						if valNode.pegRule == ruleCapture {
							val := extractText(valNode)
							fvals = append(fvals, val)
						}
					}
				case ruleMultipleStrings:
					for valNode := valNode.up; valNode != nil; valNode = valNode.next {
						if valNode.pegRule == ruleString {
							val := extractText(valNode)
							if unq, err := strconv.Unquote(val); err == nil {
								val = unq
							}
							fvals = append(fvals, val)
						}
					}
				}
			}
		}
	}
	return fname, fvals
}

func extractUASTSpec(uastSpecNode *node32) (UASTSpec, error) {
	var spec UASTSpec
	fieldsNode := findChild(uastSpecNode, ruleUASTFields)
	if fieldsNode == nil {
		return spec, fmt.Errorf("missing UAST field list")
	}
	for entryNode := fieldsNode.up; entryNode != nil; entryNode = entryNode.next {
		if entryNode.pegRule != ruleUASTField {
			continue
		}
		fieldNode := entryNode
		fname, fvals := extractUASTField(fieldNode)
		switch fname {
		case "type":
			if len(fvals) > 0 {
				spec.Type = fvals[0]
			}
		case "token":
			if len(fvals) > 0 {
				spec.Token = fvals[0]
			}
		case "roles":
			spec.Roles = append(spec.Roles, fvals...)
		case "props":
			if spec.Props == nil {
				spec.Props = make(map[string]string)
			}
			if len(fvals) > 0 {
				spec.Props[fname] = fvals[0]
			}
		case "children":
			spec.Children = append(spec.Children, fvals...)
		default:
			if spec.Props == nil {
				spec.Props = make(map[string]string)
			}
			if len(fvals) > 0 {
				spec.Props[fname] = fvals[0]
			}
		}
	}
	return spec, nil
}
