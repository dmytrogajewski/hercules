package mapping

import (
	"context"
	"strings"
	"testing"

	gositter "github.com/alexaandru/go-sitter-forest/go"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

// Rename contains to containsStringSlice to avoid redeclaration issues.
func containsStringSlice(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestParseNodeTypes(t *testing.T) {}

func TestBuildNodeRegistry(t *testing.T) {}

func TestApplyHeuristicClassification(t *testing.T) {}

func TestParseMapping(t *testing.T) {}

func TestCompileQuery(t *testing.T) {}

func TestCacheCompiledQueries(t *testing.T) {}

func TestMatchPattern(t *testing.T) {
	pm := NewPatternMatcher(nil) // legacy stub
	_, err := pm.MatchPattern(nil, nil, nil)
	if err == nil || err.Error() != "query or node is nil" {
		t.Errorf("expected query or node is nil error, got %v", err)
	}
}

func TestParseMappingRule(t *testing.T) {
	input := `function_declaration <- (function_declaration name: (identifier) @name body: (block) @body) => uast(
    type: "Function",
    token: @name,
    roles: "Declaration",
    children: @body
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	rule := rules[0]
	if rule.Name != "function_declaration" {
		t.Errorf("expected rule name 'function_declaration', got '%s'", rule.Name)
	}
	if rule.UASTSpec.Type != "Function" {
		t.Errorf("expected UAST type 'Function', got '%s'", rule.UASTSpec.Type)
	}
	if rule.UASTSpec.Token != "@name" {
		t.Errorf("expected token '@name', got '%s'", rule.UASTSpec.Token)
	}
	if len(rule.UASTSpec.Roles) == 0 || rule.UASTSpec.Roles[0] != "Declaration" {
		t.Errorf("expected roles to contain 'Declaration', got %v", rule.UASTSpec.Roles)
	}
	if len(rule.UASTSpec.Children) == 0 || rule.UASTSpec.Children[0] != "@body" {
		t.Errorf("expected children to contain '@body', got %v", rule.UASTSpec.Children)
	}
}

func TestParseMappingRule_MultiRoles(t *testing.T) {
	input := `base_rule <- (base_node) => uast(
    type: "Base",
    roles: "ChildRole", "ExtraRole"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rules, got %d", len(rules))
	}
	base := rules[0]
	if base.Name != "base_rule" {
		t.Errorf("unexpected rule names: %v", base.Name)
	}
	if len(base.UASTSpec.Roles) != 2 || base.UASTSpec.Roles[0] != "ChildRole" || base.UASTSpec.Roles[1] != "ExtraRole" {
		t.Errorf("expected roles [ChildRole, ExtraRole], got %v", base.UASTSpec.Roles)
	}
}

func TestParseMappingRule_InheritanceAndConditions(t *testing.T) {
	input := `base_rule <- (base_node) => uast(
    type: "Base",
    roles: "BaseRole"
)

child_rule <- (child_node) => uast(
    type: "Child",
    roles: "ChildRole", "ExtraRole"
) when some_field == "value" and other_field != "bad"

inherited_rule <- (inherited_node) => uast(
    type: "Inherited"
) # Extends base_rule
`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	base := rules[0]
	child := rules[1]
	inherited := rules[2]
	if base.Name != "base_rule" || child.Name != "child_rule" || inherited.Name != "inherited_rule" {
		t.Errorf("unexpected rule names: %v, %v, %v", base.Name, child.Name, inherited.Name)
	}
	if len(child.UASTSpec.Roles) != 2 || child.UASTSpec.Roles[0] != "ChildRole" || child.UASTSpec.Roles[1] != "ExtraRole" {
		t.Errorf("expected roles [ChildRole, ExtraRole], got %v", child.UASTSpec.Roles)
	}
	if len(child.Conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(child.Conditions))
	}
	if !strings.Contains(child.Conditions[0].Expr, "some_field") {
		t.Errorf("expected first condition to mention 'some_field', got %v", child.Conditions[0].Expr)
	}
}

func TestPatternMatcher_CompileAndCache(t *testing.T) {
	pm := NewPatternMatcher(nil) // legacy stub
	_, err := pm.CompileAndCache("(identifier) @id")
	if err == nil {
		t.Errorf("expected error for nil language, got nil")
	}
	// Should cache the error result
	_, err2 := pm.CompileAndCache("(identifier) @id")
	if err2 == nil {
		t.Errorf("expected cached error for nil language, got nil")
	}
}

func TestPatternMatcher_RealGoFunctionMatch(t *testing.T) {
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	pm := NewPatternMatcher(lang)

	source := []byte(`package main
func Hello() {}`)
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseString(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}
	root := tree.RootNode()

	pattern := `(
	  (function_declaration
	    name: (identifier) @funcname
	  )
	)`
	query, err := pm.CompileAndCache(pattern)
	if err != nil {
		t.Fatalf("CompileAndCache failed: %v", err)
	}
	captures, err := pm.MatchPattern(query, &root, source)
	if err != nil {
		t.Fatalf("MatchPattern failed: %v", err)
	}
	name, ok := captures["funcname"]
	if !ok {
		t.Fatalf("Expected capture 'funcname' not found in %v", captures)
	}
	if name != "Hello" {
		t.Errorf("Expected function name 'Hello', got '%s'", name)
	}
}

func TestLoadMappings_Valid(t *testing.T) {
	input := `function_declaration <- (function_declaration name: (identifier) @name body: (block) @body) => uast(
    type: "Function",
    token: @name,
    roles: "Declaration",
    children: @body
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "function_declaration" {
		t.Errorf("expected rule name 'function_declaration', got '%s'", rules[0].Name)
	}
}

func TestLoadMappings_Invalid(t *testing.T) {
	input := `function_declaration <- (function_declaration) => uast()` // missing fields
	parser := &MappingParser{}
	_, err := parser.ParseMapping(input)
	if err == nil {
		t.Error("expected error for invalid mapping DSL, got nil")
	}
}

func TestParseMappingRule_ChildrenDeduplication(t *testing.T) {
	input := `complex_node <- (complex_node field1: (child1) @c1 (child2) @c2 (child1) @c1) => uast(type: "Complex", children: @c1, @c2)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	children := rules[0].UASTSpec.Children
	if len(children) != 2 || !containsStringSlice(children, "@c1") || !containsStringSlice(children, "@c2") {
		t.Errorf("expected children to contain '@c1' and '@c2', got %v", children)
	}
}

func TestParseMappingRule_InheritanceAndConditions_Advanced(t *testing.T) {
	input := `base_rule <- (base_node) => uast(
    type: "Base",
    roles: "BaseRole"
)

child_rule <- (child_node) => uast(
    type: "Child",
    roles: "ChildRole"
) # Extends base when field == "val"`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	child := rules[1]
	if child.Extends != "base" {
		t.Errorf("expected Extends 'base', got '%s'", child.Extends)
	}
	if len(child.Conditions) != 1 || !strings.Contains(child.Conditions[0].Expr, "field == \"val\"") {
		t.Errorf("expected condition on 'field', got %v", child.Conditions)
	}
}

// Tests for advanced token extraction features
func TestParseMappingRule_TokenExtractionStrategies(t *testing.T) {
	input := `self_token <- (identifier) => uast(
    type: "Identifier",
    token: "self"
)

child_token <- (function_call function: (identifier) @func) => uast(
    type: "Call",
    token: "child:identifier"
)

descendant_token <- (complex_expression) => uast(
    type: "Expression",
    token: "descendant:identifier"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	// Test self token extraction
	selfRule := rules[0]
	if selfRule.UASTSpec.Token != "self" {
		t.Errorf("expected self token extraction, got '%s'", selfRule.UASTSpec.Token)
	}

	// Test child token extraction
	childRule := rules[1]
	if childRule.UASTSpec.Token != "child:identifier" {
		t.Errorf("expected child token extraction, got '%s'", childRule.UASTSpec.Token)
	}

	// Test descendant token extraction
	descRule := rules[2]
	if descRule.UASTSpec.Token != "descendant:identifier" {
		t.Errorf("expected descendant token extraction, got '%s'", descRule.UASTSpec.Token)
	}
}

// Tests for property mapping
func TestParseMappingRule_PropertyMapping(t *testing.T) {
	input := `function_declaration <- (function_declaration name: (identifier) @name params: (parameter_list) @params body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration",
    children: "@params", "@body",
    name: "@name",
    params: "@params",
    body: "@body"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	expectedProps := map[string]string{
		"name":   "@name",
		"params": "@params",
		"body":   "@body",
	}

	for key, expectedValue := range expectedProps {
		if actualValue, exists := rule.UASTSpec.Props[key]; !exists {
			t.Errorf("expected property '%s', not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("expected property '%s' to be '%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}

// Tests for real-world Go language mapping examples
func TestParseMappingRule_GoLanguageMapping(t *testing.T) {
	input := `function_declaration <- (function_declaration name: (identifier) @name params: (parameter_list) @params body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration",
    children: "@params", "@body",
    name: "@name",
    params: "@params"
)

method_declaration <- (method_declaration name: (identifier) @name receiver: (parameter_list) @receiver params: (parameter_list) @params body: (block) @body) => uast(
    type: "Method",
    token: "@name",
    roles: "Declaration", "Method"
) # Extends function_declaration

var_declaration <- (var_declaration name: (identifier) @name type: (type_annotation) @type value: (expression) @value) => uast(
    type: "Variable",
    token: "@name",
    roles: "Declaration", "Variable",
    children: "@type", "@value",
    name: "@name",
    type_info: @type
)

if_statement <- (if_statement condition: (expression) @cond consequence: (block) @conseq alternative: (block) @alt) => uast(
    type: "If",
    roles: "Statement", "Conditional",
    children: "@cond", "@conseq", "@alt"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}

	// Test function declaration
	funcRule := rules[0]
	if funcRule.Name != "function_declaration" {
		t.Errorf("expected function_declaration rule, got '%s'", funcRule.Name)
	}
	if funcRule.UASTSpec.Type != "Function" {
		t.Errorf("expected Function type, got '%s'", funcRule.UASTSpec.Type)
	}
	if len(funcRule.UASTSpec.Roles) != 1 || funcRule.UASTSpec.Roles[0] != "Declaration" {
		t.Errorf("expected Declaration role, got %v", funcRule.UASTSpec.Roles)
	}

	// Test method declaration inheritance
	methodRule := rules[1]
	if methodRule.Name != "method_declaration" {
		t.Errorf("expected method_declaration rule, got '%s'", methodRule.Name)
	}
	if methodRule.Extends != "function_declaration" {
		t.Errorf("expected inheritance from function_declaration, got '%s'", methodRule.Extends)
	}
	if len(methodRule.UASTSpec.Roles) != 2 || !containsStringSlice(methodRule.UASTSpec.Roles, "Declaration") || !containsStringSlice(methodRule.UASTSpec.Roles, "Method") {
		t.Errorf("expected Declaration and Method roles, got %v", methodRule.UASTSpec.Roles)
	}

	// Test variable declaration with descendant property
	varRule := rules[2]
	if varRule.Name != "var_declaration" {
		t.Errorf("expected var_declaration rule, got '%s'", varRule.Name)
	}
	if varRule.UASTSpec.Type != "Variable" {
		t.Errorf("expected Variable type, got '%s'", varRule.UASTSpec.Type)
	}
	if len(varRule.UASTSpec.Roles) != 2 || !containsStringSlice(varRule.UASTSpec.Roles, "Declaration") || !containsStringSlice(varRule.UASTSpec.Roles, "Variable") {
		t.Errorf("expected Declaration and Variable roles, got %v", varRule.UASTSpec.Roles)
	}

	// Test if statement
	ifRule := rules[3]
	if ifRule.Name != "if_statement" {
		t.Errorf("expected if_statement rule, got '%s'", ifRule.Name)
	}
	if ifRule.UASTSpec.Type != "If" {
		t.Errorf("expected If type, got '%s'", ifRule.UASTSpec.Type)
	}
	if len(ifRule.UASTSpec.Roles) != 2 || !containsStringSlice(ifRule.UASTSpec.Roles, "Statement") || !containsStringSlice(ifRule.UASTSpec.Roles, "Conditional") {
		t.Errorf("expected Statement and Conditional roles, got %v", ifRule.UASTSpec.Roles)
	}
}

// Tests for Python mapping examples
func TestParseMappingRule_PythonMapping(t *testing.T) {
	input := `function_definition <- (function_definition name: (identifier) @name parameters: (parameters) @params body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration", "Function",
    children: "@params", "@body"
)

class_definition <- (class_definition name: (identifier) @name body: (block) @body) => uast(
    type: "Class",
    token: "@name",
    roles: "Declaration", "Class",
    children: "@body"
)

base_expression <- (expression) => uast(
    type: "Expression",
    roles: "Expression"
)

comparison_expression <- (comparison_expression left: (expression) @left operator: (comparison_operator) @op right: (expression) @right) => uast(
    type: "Comparison",
    token: "@op",
    roles: "Expression", "Comparison"
) # Extends base_expression`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}

	// Test function definition
	funcDefRule := rules[0]
	if funcDefRule.Name != "function_definition" {
		t.Errorf("expected function_definition rule, got '%s'", funcDefRule.Name)
	}
	if funcDefRule.UASTSpec.Type != "Function" {
		t.Errorf("expected Function type, got '%s'", funcDefRule.UASTSpec.Type)
	}
	if len(funcDefRule.UASTSpec.Roles) != 2 || !containsStringSlice(funcDefRule.UASTSpec.Roles, "Declaration") || !containsStringSlice(funcDefRule.UASTSpec.Roles, "Function") {
		t.Errorf("expected Declaration and Function roles, got %v", funcDefRule.UASTSpec.Roles)
	}

	// Test class definition
	classDefRule := rules[1]
	if classDefRule.Name != "class_definition" {
		t.Errorf("expected class_definition rule, got '%s'", classDefRule.Name)
	}
	if classDefRule.UASTSpec.Type != "Class" {
		t.Errorf("expected Class type, got '%s'", classDefRule.UASTSpec.Type)
	}

	// Test base expression
	baseRule := rules[2]
	if baseRule.Name != "base_expression" {
		t.Errorf("expected base_expression rule, got '%s'", baseRule.Name)
	}
	if baseRule.UASTSpec.Type != "Expression" {
		t.Errorf("expected Expression type, got '%s'", baseRule.UASTSpec.Type)
	}

	// Test comparison expression with inheritance
	compRule := rules[3]
	if compRule.Name != "comparison_expression" {
		t.Errorf("expected comparison_expression rule, got '%s'", compRule.Name)
	}
	if compRule.Extends != "base_expression" {
		t.Errorf("expected inheritance from base_expression, got '%s'", compRule.Extends)
	}
}

// Tests for recipes and best practices
func TestParseMappingRule_LanguageAgnosticFunctionMapping(t *testing.T) {
	input := `function_base <- (function_declaration name: (identifier) @name body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration", "Function",
    children: "@body"
)

go_function <- (function_declaration name: (identifier) @name params: (parameter_list) @params body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration", "Function",
    children: "@params", "@body"
) # Extends function_base

js_function <- (function_declaration name: (identifier) @name params: (formal_parameters) @params body: (statement_block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration", "Function",
    children: "@params", "@body"
) # Extends function_base`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	// Test base function pattern
	baseRule := rules[0]
	if baseRule.Name != "function_base" {
		t.Errorf("expected function_base rule, got '%s'", baseRule.Name)
	}
	if len(baseRule.UASTSpec.Roles) != 2 || !containsStringSlice(baseRule.UASTSpec.Roles, "Declaration") || !containsStringSlice(baseRule.UASTSpec.Roles, "Function") {
		t.Errorf("expected Declaration and Function roles, got %v", baseRule.UASTSpec.Roles)
	}

	// Test Go-specific function
	goRule := rules[1]
	if goRule.Name != "go_function" {
		t.Errorf("expected go_function rule, got '%s'", goRule.Name)
	}
	if goRule.Extends != "function_base" {
		t.Errorf("expected inheritance from function_base, got '%s'", goRule.Extends)
	}

	// Test JavaScript function
	jsRule := rules[2]
	if jsRule.Name != "js_function" {
		t.Errorf("expected js_function rule, got '%s'", jsRule.Name)
	}
	if jsRule.Extends != "function_base" {
		t.Errorf("expected inheritance from function_base, got '%s'", jsRule.Extends)
	}
}

// Tests for conditional type mapping
func TestParseMappingRule_ConditionalTypeMapping(t *testing.T) {
	input := `expression <- (expression) => uast(
    type: "Expression",
    roles: "Expression"
)

arithmetic_expression <- (binary_expression left: (expression) @left operator: (arithmetic_operator) @op right: (expression) @right) => uast(
    type: "ArithmeticExpression",
    token: "@op",
    roles: "Expression", "Arithmetic"
)

logical_expression <- (binary_expression left: (expression) @left operator: (logical_operator) @op right: (expression) @right) => uast(
    type: "LogicalExpression",
    token: "@op",
    roles: "Expression", "Logical"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	// Test base expression
	baseRule := rules[0]
	if baseRule.Name != "expression" {
		t.Errorf("expected expression rule, got '%s'", baseRule.Name)
	}
	if baseRule.UASTSpec.Type != "Expression" {
		t.Errorf("expected Expression type, got '%s'", baseRule.UASTSpec.Type)
	}

	// Test arithmetic expression
	arithRule := rules[1]
	if arithRule.Name != "arithmetic_expression" {
		t.Errorf("expected arithmetic_expression rule, got '%s'", arithRule.Name)
	}
	if arithRule.UASTSpec.Type != "ArithmeticExpression" {
		t.Errorf("expected ArithmeticExpression type, got '%s'", arithRule.UASTSpec.Type)
	}

	// Test logical expression
	logicRule := rules[2]
	if logicRule.Name != "logical_expression" {
		t.Errorf("expected logical_expression rule, got '%s'", logicRule.Name)
	}
	if logicRule.UASTSpec.Type != "LogicalExpression" {
		t.Errorf("expected LogicalExpression type, got '%s'", logicRule.UASTSpec.Type)
	}
}

// Tests for advanced token extraction recipes
func TestParseMappingRule_AdvancedTokenExtraction(t *testing.T) {
	input := `function_call <- (call_expression function: (identifier) @func) => uast(
    type: "Call",
    token: "child:identifier",
    roles: "Expression", "Call",
    function: "child:identifier"
)

typed_variable <- (variable_declaration name: (identifier) @name type: (type_annotation) @type) => uast(
    type: "Variable",
    token: "@name",
    roles: "Declaration", "Variable",
    name: "child:identifier",
    type_info: "descendant:type_annotation"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	// Test function call with child token extraction
	callRule := rules[0]
	if callRule.Name != "function_call" {
		t.Errorf("expected function_call rule, got '%s'", callRule.Name)
	}
	if callRule.UASTSpec.Token != "child:identifier" {
		t.Errorf("expected child:identifier token, got '%s'", callRule.UASTSpec.Token)
	}
	if callRule.UASTSpec.Props["function"] != "child:identifier" {
		t.Errorf("expected function property, got '%s'", callRule.UASTSpec.Props["function"])
	}

	// Test typed variable with descendant property
	varRule := rules[1]
	if varRule.Name != "typed_variable" {
		t.Errorf("expected typed_variable rule, got '%s'", varRule.Name)
	}
	if varRule.UASTSpec.Props["type_info"] != "descendant:type_annotation" {
		t.Errorf("expected descendant type property, got '%s'", varRule.UASTSpec.Props["type_info"])
	}
}

// Tests for error handling and validation recipes
func TestParseMappingRule_ErrorHandlingAndValidation(t *testing.T) {
	input := `safe_property <- (object_property key: (property_identifier) @key value: (expression) @value) => uast(
    type: "Property",
    token: "@key",
    roles: "Property",
    key: "@key",
    value: "@value"
)

conditional_role <- (identifier) => uast(
    type: "Identifier",
    token: "self",
    roles: "Name"
)`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	// Test safe property
	safeRule := rules[0]
	if safeRule.Name != "safe_property" {
		t.Errorf("expected safe_property rule, got '%s'", safeRule.Name)
	}
	if safeRule.UASTSpec.Type != "Property" {
		t.Errorf("expected Property type, got '%s'", safeRule.UASTSpec.Type)
	}

	// Test conditional role assignment
	condRule := rules[1]
	if condRule.Name != "conditional_role" {
		t.Errorf("expected conditional_role rule, got '%s'", condRule.Name)
	}
	if condRule.UASTSpec.Token != "self" {
		t.Errorf("expected self token, got '%s'", condRule.UASTSpec.Token)
	}
}

// Tests for complex inheritance chains
func TestParseMappingRule_ComplexInheritanceChain(t *testing.T) {
	input := `base_expression <- (expression) => uast(
    type: "Expression",
    roles: "Expression"
)

binary_expression <- (binary_expression left: (expression) @left op: (operator) @op right: (expression) @right) => uast(
    type: "BinaryExpression",
    token: "@op",
    roles: "Expression", "Binary"
) # Extends base_expression

arithmetic_expression <- (arithmetic_expression left: (expression) @left op: (arithmetic_operator) @op right: (expression) @right) => uast(
    type: "ArithmeticExpression",
    token: "@op",
    roles: "Expression", "Arithmetic"
) # Extends binary_expression`
	parser := &MappingParser{}
	rules, err := parser.ParseMapping(input)
	if err != nil {
		t.Fatalf("ParseMapping failed: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	// Test base expression
	baseRule := rules[0]
	if baseRule.Name != "base_expression" {
		t.Errorf("expected base_expression rule, got '%s'", baseRule.Name)
	}

	// Test binary expression inheritance
	binaryRule := rules[1]
	if binaryRule.Name != "binary_expression" {
		t.Errorf("expected binary_expression rule, got '%s'", binaryRule.Name)
	}
	if binaryRule.Extends != "base_expression" {
		t.Errorf("expected inheritance from base_expression, got '%s'", binaryRule.Extends)
	}

	// Test arithmetic expression with inheritance
	arithRule := rules[2]
	if arithRule.Name != "arithmetic_expression" {
		t.Errorf("expected arithmetic_expression rule, got '%s'", arithRule.Name)
	}
	if arithRule.Extends != "binary_expression" {
		t.Errorf("expected inheritance from binary_expression, got '%s'", arithRule.Extends)
	}
}
