package complexity

import (
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestComplexityAnalyzer_Basic(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	// Test basic functionality
	if analyzer.Name() != "complexity_analysis" {
		t.Errorf("Expected name 'complexity_analysis', got '%s'", analyzer.Name())
	}

	thresholds := analyzer.Thresholds()
	if len(thresholds) != 3 {
		t.Errorf("Expected 3 thresholds, got %d", len(thresholds))
	}

	// Test cyclomatic complexity thresholds
	cyclomaticThreshold, exists := thresholds["cyclomatic_complexity"]
	if !exists {
		t.Error("Expected 'cyclomatic_complexity' threshold to exist")
	}

	expectedValues := map[string]int{"green": 1, "yellow": 5, "red": 10}
	for key, expected := range expectedValues {
		if value, ok := cyclomaticThreshold[key]; !ok {
			t.Errorf("Expected threshold key '%s' to exist", key)
		} else if value != expected {
			t.Errorf("Expected threshold '%s' to be %d, got %v", key, expected, value)
		}
	}
}

func TestComplexityAnalyzer_NilRoot(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}
	result, err := analyzer.Analyze(nil)

	if err != nil {
		t.Errorf("Expected no error for nil root, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result for nil root")
		return
	}

	if total, ok := result["total_complexity"]; !ok {
		t.Error("Expected 'total_complexity' in result")
	} else if total != 0 {
		t.Errorf("Expected total_complexity to be 0, got %v", total)
	}

	if functions, ok := result["functions"]; !ok {
		t.Error("Expected 'functions' in result")
	} else if functions == nil {
		t.Error("Expected non-nil functions map")
	}

	if count, ok := result["function_count"]; !ok {
		t.Error("Expected 'function_count' in result")
	} else if count != 0 {
		t.Errorf("Expected function_count to be 0, got %v", count)
	}
}

func TestComplexityAnalyzer_SimpleFunction(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	// Create a simple function node
	functionNode := node.NewWithType(node.UASTFunction)
	functionNode.Roles = []node.Role{node.RoleFunction, node.RoleDeclaration}
	nameNode := node.NewNodeWithToken(node.UASTIdentifier, "testFunction")
	nameNode.Roles = []node.Role{node.RoleName}
	functionNode.AddChild(nameNode)

	// Create a root node with the function
	root := node.NewWithType(node.UASTFile)
	root.AddChild(functionNode)

	result, err := analyzer.Analyze(root)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	if total, ok := result["total_complexity"]; !ok {
		t.Error("Expected 'total_complexity' in result")
	} else if total != 1 {
		t.Errorf("Expected total_complexity to be 1 for simple function, got %v", total)
	}

	if count, ok := result["function_count"]; !ok {
		t.Error("Expected 'function_count' in result")
	} else if count != 1 {
		t.Errorf("Expected function_count to be 1, got %v", count)
	}
}

func TestComplexityAnalyzer_ExtractFunctionName(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	// Test function with Name role in children
	functionNode := node.NewWithType(node.UASTFunction)
	nameNode := node.NewNodeWithToken(node.UASTIdentifier, "testFunction")
	nameNode.Roles = []node.Role{node.RoleName}
	functionNode.AddChild(nameNode)

	name := analyzer.extractFunctionName(functionNode)
	if name != "testFunction" {
		t.Errorf("Expected function name 'testFunction', got '%s'", name)
	}

	// Test function with name in props
	functionNode2 := node.NewWithType(node.UASTFunction)
	functionNode2.Props = map[string]string{"name": "propFunction"}

	name = analyzer.extractFunctionName(functionNode2)
	if name != "propFunction" {
		t.Errorf("Expected function name 'propFunction', got '%s'", name)
	}

	// Test function with no name (should return anonymous)
	functionNode3 := node.NewWithType(node.UASTFunction)
	name = analyzer.extractFunctionName(functionNode3)
	if name != "anonymous" {
		t.Errorf("Expected function name 'anonymous', got '%s'", name)
	}
}

func TestComplexityAnalyzer_IsDecisionPoint(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	// Test decision point types
	decisionTypes := []string{
		node.UASTIf, node.UASTLoop, node.UASTSwitch, node.UASTCase,
		node.UASTTry, node.UASTCatch, node.UASTThrow, node.UASTBreak,
		node.UASTContinue, node.UASTReturn,
	}

	for _, nodeType := range decisionTypes {
		testNode := node.NewWithType(nodeType)
		testNode.Roles = []node.Role{node.RoleCondition}

		if !analyzer.isDecisionPoint(testNode) {
			t.Errorf("Expected node type '%s' to be a decision point", nodeType)
		}
	}

	// Test non-decision point type
	regularNode := node.NewWithType(node.UASTIdentifier)
	if analyzer.isDecisionPoint(regularNode) {
		t.Error("Expected identifier node to not be a decision point")
	}

	// Test logical operators
	logicalOpNode := node.NewWithType(node.UASTBinaryOp)
	logicalOpNode.Props = map[string]string{"operator": "&&"}
	if !analyzer.isDecisionPoint(logicalOpNode) {
		t.Error("Expected binary op with '&&' to be a decision point")
	}

	// Test non-logical operator
	arithmeticOpNode := node.NewWithType(node.UASTBinaryOp)
	arithmeticOpNode.Props = map[string]string{"operator": "+"}
	if analyzer.isDecisionPoint(arithmeticOpNode) {
		t.Error("Expected binary op with '+' to not be a decision point")
	}
}

func TestComplexityAnalyzer_WithIfStatement(t *testing.T) {
	analyzer := &ComplexityAnalyzer{}

	// Create a function with an if statement
	functionNode := node.NewWithType(node.UASTFunction)
	functionNode.Roles = []node.Role{node.RoleFunction, node.RoleDeclaration}

	// Add function name
	nameNode := node.NewNodeWithToken(node.UASTIdentifier, "testFunction")
	nameNode.Roles = []node.Role{node.RoleName}
	functionNode.AddChild(nameNode)

	// Add if statement
	ifNode := node.NewWithType(node.UASTIf)
	ifNode.Roles = []node.Role{node.RoleCondition}
	functionNode.AddChild(ifNode)

	// Create a root node with the function
	root := node.NewWithType(node.UASTFile)
	root.AddChild(functionNode)

	result, err := analyzer.Analyze(root)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
		return
	}

	if total, ok := result["total_complexity"]; !ok {
		t.Error("Expected 'total_complexity' in result")
	} else if total != 2 {
		t.Errorf("Expected total_complexity to be 2 for function with if, got %v", total)
	}
}

func TestComplexityAnalyzer_LegacyCompatibility(t *testing.T) {
	analyzer := &CyclomaticComplexityAnalyzer{}

	// Test basic functionality
	if analyzer.Name() != "cyclomatic_complexity" {
		t.Errorf("Expected name 'cyclomatic_complexity', got '%s'", analyzer.Name())
	}

	thresholds := analyzer.Thresholds()
	if len(thresholds) != 1 {
		t.Errorf("Expected 1 threshold, got %d", len(thresholds))
	}

	complexityThreshold, exists := thresholds["complexity"]
	if !exists {
		t.Error("Expected 'complexity' threshold to exist")
	}

	expectedValues := map[string]int{"green": 1, "yellow": 5, "red": 10}
	for key, expected := range expectedValues {
		if value, ok := complexityThreshold[key]; !ok {
			t.Errorf("Expected threshold key '%s' to exist", key)
		} else if value != expected {
			t.Errorf("Expected threshold '%s' to be %d, got %v", key, expected, value)
		}
	}
}
