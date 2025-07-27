package halstead

import (
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestHalsteadAnalyzer_Basic(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Test empty result
	result, err := analyzer.Analyze(nil)
	if err == nil {
		t.Fatalf("Expected error for nil root, got nil")
	}

	// Test with empty root
	emptyNode := node.New("empty", "File", "", nil, nil, nil)
	result, err = analyzer.Analyze(emptyNode)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["total_functions"] != 0 {
		t.Errorf("Expected total_functions 0, got %v", result["total_functions"])
	}
}

func TestHalsteadAnalyzer_SimpleFunction(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Create a simple function node
	functionNode := node.New("func1", "Function", "", []node.Role{node.RoleFunction, node.RoleDeclaration}, nil, map[string]string{"name": "simpleFunction"})
	nameNode := node.New("name1", "Identifier", "simpleFunction", []node.Role{node.RoleName}, nil, map[string]string{"name": "simpleFunction"})

	// Add some basic operators and operands
	assignmentNode := node.New("assign1", "Assignment", "=", []node.Role{node.RoleAssignment}, nil, map[string]string{"operator": "="})
	identifierNode := node.New("id1", "Identifier", "x", []node.Role{node.RoleVariable}, nil, map[string]string{"name": "x"})
	literalNode := node.New("lit1", "Literal", "5", []node.Role{node.RoleLiteral}, nil, map[string]string{"value": "5"})

	functionNode.AddChild(nameNode)
	functionNode.AddChild(assignmentNode)
	assignmentNode.AddChild(identifierNode)
	assignmentNode.AddChild(literalNode)

	result, err := analyzer.Analyze(functionNode)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["total_functions"] != 1 {
		t.Errorf("Expected total_functions 1, got %v", result["total_functions"])
	}

	functions := result["functions"].([]map[string]interface{})
	if len(functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(functions))
	}

	funcMetrics := functions[0]
	if funcMetrics["name"] != "simpleFunction" {
		t.Errorf("Expected function name 'simpleFunction', got %v", funcMetrics["name"])
	}

	// Check basic metrics
	if volume, ok := funcMetrics["volume"].(float64); !ok || volume <= 0 {
		t.Errorf("Expected positive volume, got %v", volume)
	}

	if difficulty, ok := funcMetrics["difficulty"].(float64); !ok || difficulty <= 0 {
		t.Errorf("Expected positive difficulty, got %v", difficulty)
	}

	if effort, ok := funcMetrics["effort"].(float64); !ok || effort <= 0 {
		t.Errorf("Expected positive effort, got %v", effort)
	}
}

func TestHalsteadAnalyzer_CalculateMetrics(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Test metrics calculation with known values
	metrics := &FunctionHalsteadMetrics{
		DistinctOperators: 3,
		DistinctOperands:  4,
		TotalOperators:    6,
		TotalOperands:     8,
	}

	analyzer.calculateHalsteadMetrics(metrics)

	// Check calculated values
	expectedVocabulary := 7 // 3 + 4
	if metrics.Vocabulary != expectedVocabulary {
		t.Errorf("Expected vocabulary %d, got %d", expectedVocabulary, metrics.Vocabulary)
	}

	expectedLength := 14 // 6 + 8
	if metrics.Length != expectedLength {
		t.Errorf("Expected length %d, got %d", expectedLength, metrics.Length)
	}

	// Volume should be positive
	if metrics.Volume <= 0 {
		t.Errorf("Expected positive volume, got %f", metrics.Volume)
	}

	// Difficulty should be positive
	if metrics.Difficulty <= 0 {
		t.Errorf("Expected positive difficulty, got %f", metrics.Difficulty)
	}

	// Effort should be positive
	if metrics.Effort <= 0 {
		t.Errorf("Expected positive effort, got %f", metrics.Effort)
	}
}

func TestHalsteadAnalyzer_Thresholds(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()
	thresholds := analyzer.Thresholds()

	// Check that thresholds exist
	if thresholds["volume"] == nil {
		t.Error("Expected volume thresholds")
	}

	if thresholds["difficulty"] == nil {
		t.Error("Expected difficulty thresholds")
	}

	if thresholds["effort"] == nil {
		t.Error("Expected effort thresholds")
	}

	// Check threshold values are reasonable
	if volumeThresholds, ok := thresholds["volume"]; ok {
		if green, ok := volumeThresholds["green"].(int); ok {
			if green <= 0 {
				t.Error("Expected positive green threshold for volume")
			}
		}
	}

	if difficultyThresholds, ok := thresholds["difficulty"]; ok {
		if green, ok := difficultyThresholds["green"].(int); ok {
			if green <= 0 {
				t.Error("Expected positive green threshold for difficulty")
			}
		}
	}

	if effortThresholds, ok := thresholds["effort"]; ok {
		if green, ok := effortThresholds["green"].(int); ok {
			if green <= 0 {
				t.Error("Expected positive green threshold for effort")
			}
		}
	}
}

func TestHalsteadAnalyzer_MessageGeneration(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Test excellent complexity
	message := analyzer.getHalsteadMessage(50, 3, 500)
	if message == "" {
		t.Error("Expected non-empty message for excellent complexity")
	}

	// Test high complexity
	message = analyzer.getHalsteadMessage(6000, 35, 60000)
	if message == "" {
		t.Error("Expected non-empty message for high complexity")
	}
}

func TestHalsteadAnalyzer_OperatorOperandDetection(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Test operator detection
	operatorNode := node.New("op", "BinaryOp", "+", []node.Role{node.RoleOperator}, nil, map[string]string{"operator": "+"})
	if !analyzer.isOperator(operatorNode) {
		t.Error("Expected binary operator to be detected as operator")
	}

	// Test operand detection
	operandNode := node.New("var", "Identifier", "x", []node.Role{node.RoleVariable}, nil, map[string]string{"name": "x"})
	if !analyzer.isOperand(operandNode) {
		t.Error("Expected identifier to be detected as operand")
	}

	// Test literal operand
	literalNode := node.New("lit", "Literal", "42", []node.Role{node.RoleLiteral}, nil, map[string]string{"value": "42"})
	if !analyzer.isOperand(literalNode) {
		t.Error("Expected literal to be detected as operand")
	}
}

func TestHalsteadAnalyzer_FunctionNameExtraction(t *testing.T) {
	analyzer := NewHalsteadAnalyzer()

	// Test function with name in properties
	functionNode := node.New("func", "FunctionDecl", "", []node.Role{node.RoleFunction}, nil, map[string]string{"name": "testFunction"})
	name := analyzer.extractFunctionName(functionNode)
	if name != "testFunction" {
		t.Errorf("Expected function name 'testFunction', got '%s'", name)
	}

	// Test function with name in child identifier
	functionNode = node.New("func", "FunctionDecl", "", []node.Role{node.RoleFunction}, nil, nil)
	nameNode := node.New("name", "Identifier", "childFunction", []node.Role{node.RoleName}, nil, map[string]string{"name": "childFunction"})
	functionNode.AddChild(nameNode)

	name = analyzer.extractFunctionName(functionNode)
	if name != "childFunction" {
		t.Errorf("Expected function name 'childFunction', got '%s'", name)
	}
}
