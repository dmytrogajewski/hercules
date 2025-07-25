package cohesion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestCohesionAnalyzer_Name(t *testing.T) {
	analyzer := &CohesionAnalyzer{}
	expected := "cohesion"
	if got := analyzer.Name(); got != expected {
		t.Errorf("CohesionAnalyzer.Name() = %v, want %v", got, expected)
	}
}

func TestCohesionAnalyzer_Thresholds(t *testing.T) {
	analyzer := &CohesionAnalyzer{}
	thresholds := analyzer.Thresholds()

	// Check that all expected metrics are present
	expectedMetrics := []string{"lcom", "cohesion_score", "function_cohesion"}
	for _, metric := range expectedMetrics {
		if _, exists := thresholds[metric]; !exists {
			t.Errorf("Expected threshold for metric '%s' not found", metric)
		}
	}

	// Check specific threshold values
	if lcom, exists := thresholds["lcom"]; exists {
		if red, ok := lcom["red"].(float64); !ok || red != 4.0 {
			t.Errorf("Expected LCOM red threshold to be 4.0, got %v", red)
		}
		if yellow, ok := lcom["yellow"].(float64); !ok || yellow != 2.0 {
			t.Errorf("Expected LCOM yellow threshold to be 2.0, got %v", yellow)
		}
		if green, ok := lcom["green"].(float64); !ok || green != 1.0 {
			t.Errorf("Expected LCOM green threshold to be 1.0, got %v", green)
		}
	}
}

func TestCohesionAnalyzer_Analyze_NilRoot(t *testing.T) {
	analyzer := &CohesionAnalyzer{}
	_, err := analyzer.Analyze(nil)
	if err == nil {
		t.Error("Expected error when root node is nil")
	}
	if !strings.Contains(err.Error(), "root node is nil") {
		t.Errorf("Expected error message to contain 'root node is nil', got: %v", err.Error())
	}
}

func TestCohesionAnalyzer_Analyze_NoFunctions(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a simple UAST with no functions
	root := &node.Node{
		Type: "File",
		Children: []*node.Node{
			{
				Type:  "Comment",
				Token: "// This is a comment",
			},
		},
	}

	report, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check expected values for no functions
	if totalFunctions, ok := report["total_functions"].(int); !ok || totalFunctions != 0 {
		t.Errorf("Expected total_functions to be 0, got %v", totalFunctions)
	}
	if lcom, ok := report["lcom"].(float64); !ok || lcom != 0.0 {
		t.Errorf("Expected lcom to be 0.0, got %v", lcom)
	}
	if cohesionScore, ok := report["cohesion_score"].(float64); !ok || cohesionScore != 1.0 {
		t.Errorf("Expected cohesion_score to be 1.0, got %v", cohesionScore)
	}
	if message, ok := report["message"].(string); !ok || message != "No functions found" {
		t.Errorf("Expected message to be 'No functions found', got %v", message)
	}
}

func TestCohesionAnalyzer_Analyze_SingleFunction(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a UAST with a single function
	root := &node.Node{
		Type: "File",
		Children: []*node.Node{
			{
				Type:  "Function",
				Roles: []node.Role{"Function", "Declaration"},
				Props: map[string]string{"name": "simpleFunction"},
				Children: []*node.Node{
					{
						Type:  "Parameter",
						Roles: []node.Role{"Parameter"},
						Props: map[string]string{"name": "x"},
						Children: []*node.Node{
							{
								Type:  "Identifier",
								Token: "x",
								Roles: []node.Role{"Name"},
							},
						},
					},
					{
						Type:  "Block",
						Roles: []node.Role{"Body"},
						Children: []*node.Node{
							{
								Type:  "Return",
								Roles: []node.Role{"Return"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "x",
										Roles: []node.Role{"Name"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	report, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check expected values for single function
	if totalFunctions, ok := report["total_functions"].(int); !ok || totalFunctions != 1 {
		t.Errorf("Expected total_functions to be 1, got %v", totalFunctions)
	}
	if lcom, ok := report["lcom"].(float64); !ok || lcom != 0.0 {
		t.Errorf("Expected lcom to be 0.0 for single function, got %v", lcom)
	}
	if cohesionScore, ok := report["cohesion_score"].(float64); !ok || cohesionScore != 1.0 {
		t.Errorf("Expected cohesion_score to be 1.0 for single function, got %v", cohesionScore)
	}
}

func TestCohesionAnalyzer_Analyze_MultipleFunctions(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a UAST with multiple functions that share variables
	root := &node.Node{
		Type: "File",
		Children: []*node.Node{
			// Function 1
			{
				Type:  "Function",
				Roles: []node.Role{"Function", "Declaration"},
				Props: map[string]string{"name": "function1"},
				Children: []*node.Node{
					{
						Type:  "Parameter",
						Roles: []node.Role{"Parameter"},
						Props: map[string]string{"name": "sharedVar"},
						Children: []*node.Node{
							{
								Type:  "Identifier",
								Token: "sharedVar",
								Roles: []node.Role{"Name"},
							},
						},
					},
					{
						Type:  "Variable",
						Roles: []node.Role{"Variable", "Declaration"},
						Props: map[string]string{"name": "localVar1"},
						Children: []*node.Node{
							{
								Type:  "Identifier",
								Token: "localVar1",
								Roles: []node.Role{"Name"},
							},
						},
					},
				},
			},
			// Function 2
			{
				Type:  "Function",
				Roles: []node.Role{"Function", "Declaration"},
				Props: map[string]string{"name": "function2"},
				Children: []*node.Node{
					{
						Type:  "Parameter",
						Roles: []node.Role{"Parameter"},
						Props: map[string]string{"name": "sharedVar"},
						Children: []*node.Node{
							{
								Type:  "Identifier",
								Token: "sharedVar",
								Roles: []node.Role{"Name"},
							},
						},
					},
					{
						Type:  "Variable",
						Roles: []node.Role{"Variable", "Declaration"},
						Props: map[string]string{"name": "localVar2"},
						Children: []*node.Node{
							{
								Type:  "Identifier",
								Token: "localVar2",
								Roles: []node.Role{"Name"},
							},
						},
					},
				},
			},
		},
	}

	report, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check expected values for multiple functions
	if totalFunctions, ok := report["total_functions"].(int); !ok || totalFunctions != 2 {
		t.Errorf("Expected total_functions to be 2, got %v", totalFunctions)
	}

	// Functions share a variable, so LCOM should be lower (can be negative when functions share many variables)
	if lcom, ok := report["lcom"].(float64); !ok {
		t.Errorf("Expected lcom to be a float64, got %v", lcom)
	}
	// LCOM can be negative when functions share many variables (good cohesion)
	// LCOM can be positive when functions don't share variables (poor cohesion)
}

func TestCohesionAnalyzer_FormatReport(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a test report
	report := analyze.Report{
		"total_functions":   2,
		"lcom":              1.5,
		"cohesion_score":    0.7,
		"function_cohesion": 0.6,
		"message":           "Good cohesion - functions are generally well-organized",
		"functions": []map[string]interface{}{
			{
				"name":           "testFunction1",
				"line_count":     5,
				"variable_count": 3,
				"cohesion":       0.8,
			},
			{
				"name":           "testFunction2",
				"line_count":     8,
				"variable_count": 6,
				"cohesion":       0.4,
			},
		},
	}

	var buf bytes.Buffer
	err := analyzer.FormatReport(report, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Check that the report contains expected sections
	expectedSections := []string{
		"Cohesion Analysis Report",
		"Overall Metrics:",
		"Total Functions: 2",
		"LCOM (Lack of Cohesion): 1.50",
		"Cohesion Score: 0.70",
		"Function Cohesion: 0.60",
		"Function Details:",
		"testFunction1",
		"testFunction2",
		"Recommendations:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain '%s', but it was not found", section)
		}
	}
}

func TestCohesionAnalyzer_FormatReportJSON(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a test report
	report := analyze.Report{
		"total_functions":   1,
		"lcom":              0.0,
		"cohesion_score":    1.0,
		"function_cohesion": 1.0,
		"message":           "Excellent cohesion",
		"functions": []map[string]interface{}{
			{
				"name":           "simpleFunction",
				"line_count":     3,
				"variable_count": 1,
				"cohesion":       1.0,
			},
		},
	}

	var buf bytes.Buffer
	err := analyzer.FormatReportJSON(report, &buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		t.Errorf("Generated output is not valid JSON: %v", err)
	}

	// Check that the JSON contains expected fields
	expectedFields := []string{"analyzer", "total_functions", "lcom", "cohesion_score", "function_cohesion", "message", "functions"}
	for _, field := range expectedFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Expected JSON to contain field '%s'", field)
		}
	}

	// Check specific values
	if analyzerName, ok := jsonData["analyzer"].(string); !ok || analyzerName != "cohesion" {
		t.Errorf("Expected analyzer to be 'cohesion', got %v", analyzerName)
	}
}

func TestCohesionAnalyzer_CreateAggregator(t *testing.T) {
	analyzer := &CohesionAnalyzer{}
	aggregator := analyzer.CreateAggregator()

	if aggregator == nil {
		t.Error("Expected CreateAggregator to return a non-nil aggregator")
	}

	// Check that it's the right type
	if _, ok := aggregator.(*CohesionAggregator); !ok {
		t.Error("Expected CreateAggregator to return a CohesionAggregator")
	}
}

func TestCohesionAggregator_Aggregate(t *testing.T) {
	aggregator := &CohesionAggregator{
		combinedFunctions: make(map[string]interface{}),
		totalLCOM:         0.0,
		totalCohesion:     0.0,
		functionCount:     0,
	}

	// Create test results
	results := map[string]analyze.Report{
		"file1": {
			"total_functions":   2,
			"lcom":              1.0,
			"cohesion_score":    0.8,
			"function_cohesion": 0.7,
			"functions": []map[string]interface{}{
				{
					"name":           "function1",
					"line_count":     5,
					"variable_count": 2,
					"cohesion":       0.8,
				},
				{
					"name":           "function2",
					"line_count":     8,
					"variable_count": 4,
					"cohesion":       0.6,
				},
			},
		},
		"file2": {
			"total_functions":   1,
			"lcom":              0.0,
			"cohesion_score":    1.0,
			"function_cohesion": 1.0,
			"functions": []map[string]interface{}{
				{
					"name":           "function3",
					"line_count":     3,
					"variable_count": 1,
					"cohesion":       1.0,
				},
			},
		},
	}

	aggregator.Aggregate(results)

	// Check aggregated values
	if aggregator.functionCount != 3 {
		t.Errorf("Expected functionCount to be 3, got %d", aggregator.functionCount)
	}
	if aggregator.totalLCOM != 1.0 {
		t.Errorf("Expected totalLCOM to be 1.0, got %f", aggregator.totalLCOM)
	}
	if aggregator.totalCohesion != 1.8 {
		t.Errorf("Expected totalCohesion to be 1.8, got %f", aggregator.totalCohesion)
	}
	if len(aggregator.combinedFunctions) != 3 {
		t.Errorf("Expected 3 combined functions, got %d", len(aggregator.combinedFunctions))
	}
}

func TestCohesionAggregator_GetResult(t *testing.T) {
	aggregator := &CohesionAggregator{
		combinedFunctions: make(map[string]interface{}),
		totalLCOM:         2.0,
		totalCohesion:     1.5,
		functionCount:     2,
	}

	// Add some functions to combinedFunctions
	aggregator.combinedFunctions["func1"] = map[string]interface{}{
		"name":           "func1",
		"line_count":     5,
		"variable_count": 2,
		"cohesion":       0.8,
	}
	aggregator.combinedFunctions["func2"] = map[string]interface{}{
		"name":           "func2",
		"line_count":     8,
		"variable_count": 4,
		"cohesion":       0.7,
	}

	result := aggregator.GetResult()

	// Check result structure
	if totalFunctions, ok := result["total_functions"].(int); !ok || totalFunctions != 2 {
		t.Errorf("Expected total_functions to be 2, got %v", totalFunctions)
	}
	if lcom, ok := result["lcom"].(float64); !ok || lcom != 1.0 {
		t.Errorf("Expected lcom to be 1.0, got %v", lcom)
	}
	if cohesionScore, ok := result["cohesion_score"].(float64); !ok || cohesionScore != 0.75 {
		t.Errorf("Expected cohesion_score to be 0.75, got %v", cohesionScore)
	}
	if functions, ok := result["functions"].([]map[string]interface{}); !ok || len(functions) != 2 {
		t.Errorf("Expected 2 functions in result, got %d", len(functions))
	}
}

func TestCohesionAggregator_GetResult_NoFunctions(t *testing.T) {
	aggregator := &CohesionAggregator{
		combinedFunctions: make(map[string]interface{}),
		totalLCOM:         0.0,
		totalCohesion:     0.0,
		functionCount:     0,
	}

	result := aggregator.GetResult()

	// Check expected values for no functions
	if totalFunctions, ok := result["total_functions"].(int); !ok || totalFunctions != 0 {
		t.Errorf("Expected total_functions to be 0, got %v", totalFunctions)
	}
	if lcom, ok := result["lcom"].(float64); !ok || lcom != 0.0 {
		t.Errorf("Expected lcom to be 0.0, got %v", lcom)
	}
	if cohesionScore, ok := result["cohesion_score"].(float64); !ok || cohesionScore != 1.0 {
		t.Errorf("Expected cohesion_score to be 1.0, got %v", cohesionScore)
	}
}

func TestCohesionAnalyzer_HelperFunctions(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Test haveSharedVariables
	fn1 := Function{
		Name:      "func1",
		Variables: []string{"a", "b", "c"},
	}
	fn2 := Function{
		Name:      "func2",
		Variables: []string{"b", "d", "e"},
	}
	fn3 := Function{
		Name:      "func3",
		Variables: []string{"x", "y", "z"},
	}

	if !analyzer.haveSharedVariables(fn1, fn2) {
		t.Error("Expected fn1 and fn2 to have shared variables")
	}
	if analyzer.haveSharedVariables(fn1, fn3) {
		t.Error("Expected fn1 and fn3 to not have shared variables")
	}

	// Test calculateCohesionScore
	score1 := analyzer.calculateCohesionScore(0.0, 1)
	if score1 != 1.0 {
		t.Errorf("Expected cohesion score to be 1.0 for single function, got %f", score1)
	}

	score2 := analyzer.calculateCohesionScore(2.0, 3)
	if score2 <= 0.0 || score2 > 1.0 {
		t.Errorf("Expected cohesion score to be between 0 and 1, got %f", score2)
	}

	// Test calculateFunctionCohesion
	functions := []Function{
		{Cohesion: 0.8},
		{Cohesion: 0.6},
		{Cohesion: 1.0},
	}
	avgCohesion := analyzer.calculateFunctionCohesion(functions)
	expected := (0.8 + 0.6 + 1.0) / 3.0
	if math.Abs(avgCohesion-expected) > 0.0001 {
		t.Errorf("Expected average cohesion to be %f, got %f", expected, avgCohesion)
	}

	// Test getCohesionMessage
	message1 := analyzer.getCohesionMessage(0.9)
	if !strings.Contains(message1, "Excellent") {
		t.Errorf("Expected excellent message for score 0.9, got: %s", message1)
	}

	message2 := analyzer.getCohesionMessage(0.2)
	if !strings.Contains(message2, "Poor") {
		t.Errorf("Expected poor message for score 0.2, got: %s", message2)
	}

	// Test getSeverityEmoji
	emoji1 := analyzer.getSeverityEmoji(0.9, "cohesion_score")
	if emoji1 != "🟢" {
		t.Errorf("Expected green emoji for high score, got: %s", emoji1)
	}

	emoji2 := analyzer.getSeverityEmoji(0.4, "cohesion_score")
	if emoji2 != "🔴" {
		t.Errorf("Expected red emoji for low score, got: %s", emoji2)
	}
}

func TestCohesionAnalyzer_EdgeCases(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Test with empty function list
	lcom := analyzer.calculateLCOM([]Function{})
	if lcom != 0.0 {
		t.Errorf("Expected LCOM to be 0.0 for empty function list, got %f", lcom)
	}

	// Test with single function
	lcom = analyzer.calculateLCOM([]Function{{Name: "single"}})
	if lcom != 0.0 {
		t.Errorf("Expected LCOM to be 0.0 for single function, got %f", lcom)
	}

	// Test function-level cohesion with zero lines
	fn := Function{LineCount: 0, Variables: []string{"a", "b"}}
	cohesion := analyzer.calculateFunctionLevelCohesion(fn)
	if cohesion != 1.0 {
		t.Errorf("Expected cohesion to be 1.0 for zero lines, got %f", cohesion)
	}

	// Test function-level cohesion with high variable density
	fn = Function{LineCount: 2, Variables: []string{"a", "b", "c", "d", "e"}}
	cohesion = analyzer.calculateFunctionLevelCohesion(fn)
	if cohesion < 0.0 || cohesion > 1.0 {
		t.Errorf("Expected cohesion to be between 0 and 1, got %f", cohesion)
	}
}

func TestCohesionAnalyzer_Integration(t *testing.T) {
	analyzer := &CohesionAnalyzer{}

	// Create a realistic UAST structure
	root := &node.Node{
		Type: "File",
		Children: []*node.Node{
			// Class/Struct
			{
				Type:  "Class",
				Roles: []node.Role{"Class", "Declaration"},
				Props: map[string]string{"name": "Calculator"},
				Children: []*node.Node{
					// Method 1
					{
						Type:  "Method",
						Roles: []node.Role{"Function", "Declaration", "Member"},
						Props: map[string]string{"name": "add"},
						Children: []*node.Node{
							{
								Type:  "Parameter",
								Roles: []node.Role{"Parameter"},
								Props: map[string]string{"name": "a"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "a",
										Roles: []node.Role{"Name"},
									},
								},
							},
							{
								Type:  "Parameter",
								Roles: []node.Role{"Parameter"},
								Props: map[string]string{"name": "b"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "b",
										Roles: []node.Role{"Name"},
									},
								},
							},
							{
								Type:  "Block",
								Roles: []node.Role{"Body"},
								Children: []*node.Node{
									{
										Type:  "Variable",
										Roles: []node.Role{"Variable", "Declaration"},
										Props: map[string]string{"name": "result"},
										Children: []*node.Node{
											{
												Type:  "Identifier",
												Token: "result",
												Roles: []node.Role{"Name"},
											},
										},
									},
								},
							},
						},
					},
					// Method 2
					{
						Type:  "Method",
						Roles: []node.Role{"Function", "Declaration", "Member"},
						Props: map[string]string{"name": "multiply"},
						Children: []*node.Node{
							{
								Type:  "Parameter",
								Roles: []node.Role{"Parameter"},
								Props: map[string]string{"name": "x"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "x",
										Roles: []node.Role{"Name"},
									},
								},
							},
							{
								Type:  "Parameter",
								Roles: []node.Role{"Parameter"},
								Props: map[string]string{"name": "y"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "y",
										Roles: []node.Role{"Name"},
									},
								},
							},
						},
					},
				},
			},
			// Standalone function
			{
				Type:  "Function",
				Roles: []node.Role{"Function", "Declaration"},
				Props: map[string]string{"name": "main"},
				Children: []*node.Node{
					{
						Type:  "Block",
						Roles: []node.Role{"Body"},
						Children: []*node.Node{
							{
								Type:  "Variable",
								Roles: []node.Role{"Variable", "Declaration"},
								Props: map[string]string{"name": "calc"},
								Children: []*node.Node{
									{
										Type:  "Identifier",
										Token: "calc",
										Roles: []node.Role{"Name"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	report, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the analysis found the expected functions
	if totalFunctions, ok := report["total_functions"].(int); !ok || totalFunctions != 3 {
		t.Errorf("Expected 3 functions (2 methods + 1 function), got %v", totalFunctions)
	}

	// Verify functions are present in the report
	if functions, ok := report["functions"].([]map[string]interface{}); ok {
		functionNames := make(map[string]bool)
		for _, fn := range functions {
			if name, ok := fn["name"].(string); ok {
				functionNames[name] = true
			}
		}

		expectedNames := []string{"add", "multiply", "main"}
		for _, name := range expectedNames {
			if !functionNames[name] {
				t.Errorf("Expected function '%s' to be found in analysis", name)
			}
		}
	}

	// Test aggregator with this result
	aggregator := analyzer.CreateAggregator()
	results := map[string]analyze.Report{"test": report}
	aggregator.Aggregate(results)

	finalResult := aggregator.GetResult()
	if finalResult == nil {
		t.Error("Expected GetResult to return a non-nil report")
	}
}

// Benchmark tests
func BenchmarkCohesionAnalyzer_Analyze(b *testing.B) {
	analyzer := &CohesionAnalyzer{}

	// Create a complex UAST for benchmarking
	root := createComplexUAST()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.Analyze(root)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkCohesionAggregator_Aggregate(b *testing.B) {
	aggregator := &CohesionAggregator{
		combinedFunctions: make(map[string]interface{}),
		totalLCOM:         0.0,
		totalCohesion:     0.0,
		functionCount:     0,
	}

	// Create test results for benchmarking
	results := createBenchmarkResults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aggregator.Aggregate(results)
	}
}

// Helper functions for benchmarks
func createComplexUAST() *node.Node {
	// Create a UAST with many functions for benchmarking
	children := make([]*node.Node, 0, 100)

	for i := 0; i < 50; i++ {
		children = append(children, &node.Node{
			Type:  "Function",
			Roles: []node.Role{"Function", "Declaration"},
			Props: map[string]string{"name": fmt.Sprintf("function%d", i)},
			Children: []*node.Node{
				{
					Type:  "Parameter",
					Roles: []node.Role{"Parameter"},
					Props: map[string]string{"name": "param"},
					Children: []*node.Node{
						{
							Type:  "Identifier",
							Token: "param",
							Roles: []node.Role{"Name"},
						},
					},
				},
				{
					Type:  "Block",
					Roles: []node.Role{"Body"},
					Children: []*node.Node{
						{
							Type:  "Variable",
							Roles: []node.Role{"Variable", "Declaration"},
							Props: map[string]string{"name": "localVar"},
							Children: []*node.Node{
								{
									Type:  "Identifier",
									Token: "localVar",
									Roles: []node.Role{"Name"},
								},
							},
						},
					},
				},
			},
		})
	}

	return &node.Node{
		Type:     "File",
		Children: children,
	}
}

func createBenchmarkResults() map[string]analyze.Report {
	results := make(map[string]analyze.Report)

	for i := 0; i < 10; i++ {
		results[fmt.Sprintf("file%d", i)] = analyze.Report{
			"total_functions":   5,
			"lcom":              float64(i),
			"cohesion_score":    0.5 + float64(i)*0.05,
			"function_cohesion": 0.6 + float64(i)*0.04,
			"functions": []map[string]interface{}{
				{
					"name":           fmt.Sprintf("func%d", i),
					"line_count":     10,
					"variable_count": 3,
					"cohesion":       0.7,
				},
			},
		}
	}

	return results
}
