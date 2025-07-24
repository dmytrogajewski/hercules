package halstead

import (
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestHalsteadAnalyzer_Basic(t *testing.T) {
	analyzer := &HalsteadAnalyzer{}

	// Test empty result
	result, err := analyzer.Analyze(nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["file_count"] != 0 {
		t.Errorf("Expected file_count 0, got %v", result["file_count"])
	}
}

func TestHalsteadAnalyzer_SimpleFunction(t *testing.T) {
	analyzer := &HalsteadAnalyzer{}

	// Create a simple function node
	functionNode := node.New("func1", "FunctionDecl", "", []node.Role{node.RoleFunction}, nil, nil)
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

	if result["file_count"] != 1 {
		t.Errorf("Expected file_count 1, got %v", result["file_count"])
	}

	functions := result["functions"].(map[string]*FunctionHalsteadMetrics)
	if len(functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(functions))
	}

	funcMetrics := functions["simpleFunction"]
	if funcMetrics == nil {
		t.Fatal("Expected function metrics for 'simpleFunction'")
	}

	// Check basic metrics
	if funcMetrics.DistinctOperators < 1 {
		t.Errorf("Expected at least 1 distinct operator, got %d", funcMetrics.DistinctOperators)
	}

	if funcMetrics.DistinctOperands < 2 {
		t.Errorf("Expected at least 2 distinct operands, got %d", funcMetrics.DistinctOperands)
	}
}

func TestHalsteadAnalyzer_CalculateMetrics(t *testing.T) {
	analyzer := &HalsteadAnalyzer{}

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
	analyzer := &HalsteadAnalyzer{}
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

func TestHalsteadAnalyzer_Configuration(t *testing.T) {
	analyzer := &HalsteadAnalyzer{}
	config := analyzer.DefaultConfig()

	// Check default configuration
	if !config.IncludeFunctionBreakdown {
		t.Error("Expected IncludeFunctionBreakdown to be true by default")
	}

	if !config.IncludeTimeEstimate {
		t.Error("Expected IncludeTimeEstimate to be true by default")
	}

	if !config.IncludeBugEstimate {
		t.Error("Expected IncludeBugEstimate to be true by default")
	}
}
