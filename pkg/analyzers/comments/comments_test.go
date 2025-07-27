package comments

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentsAnalyzer_Name(t *testing.T) {
	analyzer := NewCommentsAnalyzer()
	assert.Equal(t, "comments", analyzer.Name())
}

func TestCommentsAnalyzer_Thresholds(t *testing.T) {
	analyzer := NewCommentsAnalyzer()
	thresholds := analyzer.Thresholds()

	assert.NotNil(t, thresholds)
	assert.Contains(t, thresholds, "overall_score")
	assert.Contains(t, thresholds, "good_comments_ratio")
	assert.Contains(t, thresholds, "documentation_coverage")

	overallScore := thresholds["overall_score"]
	assert.Equal(t, 0.8, overallScore["green"])
	assert.Equal(t, 0.6, overallScore["yellow"])
	assert.Equal(t, 0.4, overallScore["red"])
}

func TestCommentsAnalyzer_CreateAggregator(t *testing.T) {
	analyzer := NewCommentsAnalyzer()
	aggregator := analyzer.CreateAggregator()

	assert.NotNil(t, aggregator)
	assert.IsType(t, &CommentsAggregator{}, aggregator)
}

func TestCommentsAnalyzer_DefaultConfig(t *testing.T) {
	analyzer := NewCommentsAnalyzer()
	config := analyzer.DefaultConfig()

	assert.Equal(t, 1.0, config.RewardScore)
	assert.Equal(t, 500, config.MaxCommentLength)
	assert.NotNil(t, config.PenaltyScores)

	// Check penalty scores for different node types
	assert.Equal(t, -0.5, config.PenaltyScores[node.UASTFunction])
	assert.Equal(t, -0.5, config.PenaltyScores[node.UASTMethod])
	assert.Equal(t, -0.3, config.PenaltyScores[node.UASTClass])
	assert.Equal(t, -0.1, config.PenaltyScores[node.UASTVariable])
}

func TestCommentsAnalyzer_Analyze_EmptyTree(t *testing.T) {
	analyzer := NewCommentsAnalyzer()
	root := node.NewWithType(node.UASTFile)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 0, result["total_comments"])
	assert.Equal(t, 0, result["good_comments"])
	assert.Equal(t, 0, result["bad_comments"])
	assert.Equal(t, 0.0, result["overall_score"])
	assert.Equal(t, 0, result["total_functions"])
	assert.Equal(t, 0, result["documented_functions"])
}

func TestCommentsAnalyzer_Analyze_GoodCommentPlacement(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with a good comment above a function
	root := node.NewWithType(node.UASTFile)

	// Add a comment
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This is a good comment"
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(comment)

	// Add a function
	function := node.NewWithType(node.UASTFunction)
	function.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	// Add function name
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "testFunction"
	name.Roles = []node.Role{node.RoleName}
	function.AddChild(name)

	root.AddChild(function)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 1, result["total_comments"])
	assert.Equal(t, 1, result["good_comments"])
	assert.Equal(t, 0, result["bad_comments"])
	assert.Equal(t, 1.0, result["overall_score"])
	assert.Equal(t, 1, result["total_functions"])
	assert.Equal(t, 1, result["documented_functions"])

	// Check if line numbers are included in comment details
	commentDetails, ok := result["comment_details"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, commentDetails, 1)

	detail := commentDetails[0]
	lineNumber, exists := detail["line_number"]
	assert.True(t, exists, "line_number field should exist")
	assert.Equal(t, 1, lineNumber, "line_number should be 1")
}

func TestCommentsAnalyzer_Analyze_BadCommentPlacement(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with a bad comment (inside function body)
	root := node.NewWithType(node.UASTFile)

	// Add a function
	function := node.NewWithType(node.UASTFunction)
	function.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   5,
	}

	// Add function name
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "testFunction"
	name.Roles = []node.Role{node.RoleName}
	function.AddChild(name)

	// Add function body
	body := node.NewWithType(node.UASTBlock)
	body.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	// Add a comment inside the function body (bad placement)
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This is a bad comment"
	comment.Pos = &node.Positions{
		StartLine: 3,
		EndLine:   3,
	}
	body.AddChild(comment)

	function.AddChild(body)
	root.AddChild(function)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 1, result["total_comments"])
	assert.Equal(t, 0, result["good_comments"])
	assert.Equal(t, 1, result["bad_comments"])
	assert.Equal(t, 0.0, result["overall_score"])
	assert.Equal(t, 1, result["total_functions"])
	assert.Equal(t, 0, result["documented_functions"])
}

func TestCommentsAnalyzer_Analyze_MixedCommentPlacement(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with both good and bad comments
	root := node.NewWithType(node.UASTFile)

	// Add a good comment above first function
	goodComment := node.NewWithType(node.UASTComment)
	goodComment.Token = "// Good comment above function"
	goodComment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(goodComment)

	// Add first function
	func1 := node.NewWithType(node.UASTFunction)
	func1.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	name1 := node.NewWithType(node.UASTIdentifier)
	name1.Token = "function1"
	name1.Roles = []node.Role{node.RoleName}
	func1.AddChild(name1)
	root.AddChild(func1)

	// Add second function without comment
	func2 := node.NewWithType(node.UASTFunction)
	func2.Pos = &node.Positions{
		StartLine: 6,
		EndLine:   8,
	}

	name2 := node.NewWithType(node.UASTIdentifier)
	name2.Token = "function2"
	name2.Roles = []node.Role{node.RoleName}
	func2.AddChild(name2)
	root.AddChild(func2)

	// Add a bad comment after second function
	badComment := node.NewWithType(node.UASTComment)
	badComment.Token = "// Bad comment after function"
	badComment.Pos = &node.Positions{
		StartLine: 9,
		EndLine:   9,
	}
	root.AddChild(badComment)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 2, result["total_comments"])
	assert.Equal(t, 1, result["good_comments"])
	assert.Equal(t, 1, result["bad_comments"])
	assert.Equal(t, 0.5, result["overall_score"])
	assert.Equal(t, 2, result["total_functions"])
	assert.Equal(t, 1, result["documented_functions"])
}

func TestCommentsAnalyzer_Analyze_ClassWithMethod(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with a class and method
	root := node.NewWithType(node.UASTFile)

	// Add a good comment above class
	classComment := node.NewWithType(node.UASTComment)
	classComment.Token = "// This is a class"
	classComment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(classComment)

	// Add a class
	class := node.NewWithType(node.UASTClass)
	class.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   8,
	}

	className := node.NewWithType(node.UASTIdentifier)
	className.Token = "TestClass"
	className.Roles = []node.Role{node.RoleName}
	class.AddChild(className)

	// Add a good comment above method
	methodComment := node.NewWithType(node.UASTComment)
	methodComment.Token = "// This is a method"
	methodComment.Pos = &node.Positions{
		StartLine: 4,
		EndLine:   4,
	}
	class.AddChild(methodComment)

	// Add a method
	method := node.NewWithType(node.UASTMethod)
	method.Pos = &node.Positions{
		StartLine: 5,
		EndLine:   7,
	}

	methodName := node.NewWithType(node.UASTIdentifier)
	methodName.Token = "testMethod"
	methodName.Roles = []node.Role{node.RoleName}
	method.AddChild(methodName)

	class.AddChild(method)
	root.AddChild(class)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 2, result["total_comments"])
	assert.Equal(t, 2, result["good_comments"])
	assert.Equal(t, 0, result["bad_comments"])
	assert.Equal(t, 1.0, result["overall_score"])
	assert.Equal(t, 2, result["total_functions"]) // class + method
	assert.Equal(t, 2, result["documented_functions"])
}

func TestCommentsAnalyzer_Analyze_UnassociatedComment(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with an unassociated comment
	root := node.NewWithType(node.UASTFile)

	// Add a comment without any function/class
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This comment is not associated with anything"
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(comment)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	assert.Equal(t, 1, result["total_comments"])
	assert.Equal(t, 0, result["good_comments"])
	assert.Equal(t, 1, result["bad_comments"])
	assert.Equal(t, 0.0, result["overall_score"])
	assert.Equal(t, 0, result["total_functions"])
	assert.Equal(t, 0, result["documented_functions"])
}

func TestCommentsAnalyzer_FindComments(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with comments at different levels
	root := node.NewWithType(node.UASTFile)

	// Add comment at root level
	comment1 := node.NewWithType(node.UASTComment)
	comment1.Token = "// Root comment"
	root.AddChild(comment1)

	// Add a function with comment
	function := node.NewWithType(node.UASTFunction)
	comment2 := node.NewWithType(node.UASTComment)
	comment2.Token = "// Function comment"
	function.AddChild(comment2)
	root.AddChild(function)

	comments := analyzer.findComments(root)
	assert.Len(t, comments, 2)
	assert.Equal(t, "// Root comment", comments[0].Token)
	assert.Equal(t, "// Function comment", comments[1].Token)
}

func TestCommentsAnalyzer_FindFunctions(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with different function types
	root := node.NewWithType(node.UASTFile)

	// Add a function
	function := node.NewWithType(node.UASTFunction)
	root.AddChild(function)

	// Add a method
	method := node.NewWithType(node.UASTMethod)
	root.AddChild(method)

	// Add a class
	class := node.NewWithType(node.UASTClass)
	root.AddChild(class)

	// Add a variable (should not be included)
	variable := node.NewWithType(node.UASTVariable)
	root.AddChild(variable)

	functions := analyzer.findFunctions(root)
	assert.Len(t, functions, 3) // function, method, class
	assert.Equal(t, string(node.UASTFunction), string(functions[0].Type))
	assert.Equal(t, string(node.UASTMethod), string(functions[1].Type))
	assert.Equal(t, string(node.UASTClass), string(functions[2].Type))
}

func TestCommentsAnalyzer_ExtractTargetName(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Test with Name role
	function := node.NewWithType(node.UASTFunction)
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "testFunction"
	name.Roles = []node.Role{node.RoleName}
	function.AddChild(name)

	result := analyzer.extractTargetName(function)
	assert.Equal(t, "testFunction", result)

	// Test with props
	function2 := node.NewWithType(node.UASTFunction)
	function2.Props = map[string]string{"name": "functionFromProps"}

	result2 := analyzer.extractTargetName(function2)
	assert.Equal(t, "functionFromProps", result2)

	// Test fallback to first identifier
	function3 := node.NewWithType(node.UASTFunction)
	identifier := node.NewWithType(node.UASTIdentifier)
	identifier.Token = "fallbackName"
	function3.AddChild(identifier)

	result3 := analyzer.extractTargetName(function3)
	assert.Equal(t, "fallbackName", result3)

	// Test unknown case
	function4 := node.NewWithType(node.UASTFunction)
	result4 := analyzer.extractTargetName(function4)
	assert.Equal(t, "unknown", result4)
}

func TestCommentsAnalyzer_IsCommentProperlyPlaced(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Test properly placed comment (directly above)
	comment := node.NewWithType(node.UASTComment)
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}

	target := node.NewWithType(node.UASTFunction)
	target.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	assert.True(t, analyzer.isCommentProperlyPlaced(comment, target))

	// Test comment with gap
	comment2 := node.NewWithType(node.UASTComment)
	comment2.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}

	target2 := node.NewWithType(node.UASTFunction)
	target2.Pos = &node.Positions{
		StartLine: 4,
		EndLine:   6,
	}

	assert.False(t, analyzer.isCommentProperlyPlaced(comment2, target2))

	// Test comment below target
	comment3 := node.NewWithType(node.UASTComment)
	comment3.Pos = &node.Positions{
		StartLine: 3,
		EndLine:   3,
	}

	target3 := node.NewWithType(node.UASTFunction)
	target3.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   2,
	}

	assert.False(t, analyzer.isCommentProperlyPlaced(comment3, target3))

	// Test with missing position info
	comment4 := node.NewWithType(node.UASTComment)
	target4 := node.NewWithType(node.UASTFunction)

	assert.False(t, analyzer.isCommentProperlyPlaced(comment4, target4))
}

func TestCommentsAggregator_Aggregate(t *testing.T) {
	aggregator := NewCommentsAggregator()

	// Test aggregation
	results := map[string]map[string]any{
		"file1": {
			"total_comments":       2,
			"good_comments":        1,
			"bad_comments":         1,
			"total_functions":      3,
			"documented_functions": 1,
			"overall_score":        0.5, // 1 good out of 2 total
		},
		"file2": {
			"total_comments":       1,
			"good_comments":        1,
			"bad_comments":         0,
			"total_functions":      2,
			"documented_functions": 1,
			"overall_score":        1.0, // 1 good out of 1 total
		},
	}

	aggregator.Aggregate(results)

	result := aggregator.GetResult()
	assert.Equal(t, 3, result["total_comments"])
	assert.Equal(t, 2, result["good_comments"])
	assert.Equal(t, 1, result["bad_comments"])
	assert.Equal(t, 5, result["total_functions"])
	assert.Equal(t, 2, result["documented_functions"])
	assert.Equal(t, 0.75, result["overall_score"]) // average of 0.5 and 1.0
}

func TestCommentsAggregator_GetResult_Empty(t *testing.T) {
	aggregator := NewCommentsAggregator()

	result := aggregator.GetResult()
	assert.Equal(t, 0, result["total_comments"])
	assert.Equal(t, 0, result["good_comments"])
	assert.Equal(t, 0, result["bad_comments"])
	assert.Equal(t, 0.0, result["overall_score"])
	assert.Equal(t, 0, result["total_functions"])
	assert.Equal(t, 0, result["documented_functions"])
}

func TestCommentsAnalyzer_DebugOutput(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with a good comment above a function
	root := node.NewWithType(node.UASTFile)

	// Add a comment
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This is a good comment"
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(comment)

	// Add a function
	function := node.NewWithType(node.UASTFunction)
	function.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	// Add function name
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "testFunction"
	name.Roles = []node.Role{node.RoleName}
	function.AddChild(name)

	root.AddChild(function)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	// Print the result structure
	t.Logf("Result keys: %v", getKeys(result))

	if commentDetails, ok := result["comment_details"]; ok {
		t.Logf("Comment details: %+v", commentDetails)
	}

	if functionSummary, ok := result["function_summary"]; ok {
		t.Logf("Function summary: %+v", functionSummary)
	}

	// Print full result for debugging
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("Full result: %s", string(resultJSON))
}

func TestCommentsAnalyzer_FormatReport(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with a good comment above a function
	root := node.NewWithType(node.UASTFile)

	// Add a comment
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This is a good comment"
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(comment)

	// Add a function
	function := node.NewWithType(node.UASTFunction)
	function.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   4,
	}

	// Add function name
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "testFunction"
	name.Roles = []node.Role{node.RoleName}
	function.AddChild(name)

	root.AddChild(function)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	// Test the formatted output
	var buf strings.Builder
	err = analyzer.FormatReport(result, &buf)
	require.NoError(t, err)

	formatted := buf.String()
	t.Logf("Formatted Report:\n%s", formatted)

	// Verify the output contains expected sections
	assert.Contains(t, formatted, "Excellent comment quality and placement")
	assert.Contains(t, formatted, "overall_score: 1.00")
	assert.Contains(t, formatted, "good_comments_ratio: 1.00")
	assert.Contains(t, formatted, "documentation_coverage: 1.00")
}

func TestCommentsAnalyzer_FormatReport_Complex(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Create a tree with multiple functions and comments
	root := node.NewWithType(node.UASTFile)

	// Add a good comment above first function
	goodComment := node.NewWithType(node.UASTComment)
	goodComment.Token = "// This is a well-documented function"
	goodComment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(goodComment)

	// Add first function (well documented)
	func1 := node.NewWithType(node.UASTFunction)
	func1.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   5,
	}
	name1 := node.NewWithType(node.UASTIdentifier)
	name1.Token = "wellDocumentedFunction"
	name1.Roles = []node.Role{node.RoleName}
	func1.AddChild(name1)
	root.AddChild(func1)

	// Add second function without comment (missing documentation)
	func2 := node.NewWithType(node.UASTFunction)
	func2.Pos = &node.Positions{
		StartLine: 7,
		EndLine:   10,
	}
	name2 := node.NewWithType(node.UASTIdentifier)
	name2.Token = "undocumentedFunction"
	name2.Roles = []node.Role{node.RoleName}
	func2.AddChild(name2)
	root.AddChild(func2)

	// Add a bad comment (inside function body)
	func3 := node.NewWithType(node.UASTFunction)
	func3.Pos = &node.Positions{
		StartLine: 12,
		EndLine:   16,
	}
	name3 := node.NewWithType(node.UASTIdentifier)
	name3.Token = "functionWithBadComment"
	name3.Roles = []node.Role{node.RoleName}
	func3.AddChild(name3)

	// Add function body
	body := node.NewWithType(node.UASTBlock)
	body.Pos = &node.Positions{
		StartLine: 13,
		EndLine:   15,
	}

	// Add a comment inside the function body (bad placement)
	badComment := node.NewWithType(node.UASTComment)
	badComment.Token = "// This comment is in the wrong place"
	badComment.Pos = &node.Positions{
		StartLine: 14,
		EndLine:   14,
	}
	body.AddChild(badComment)
	func3.AddChild(body)
	root.AddChild(func3)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	// Test the formatted output
	var buf strings.Builder
	err = analyzer.FormatReport(result, &buf)
	require.NoError(t, err)

	formatted := buf.String()
	t.Logf("Complex Formatted Report:\n%s", formatted)

	// Verify the output contains expected sections
	assert.Contains(t, formatted, "Fair comment quality - consider improving placement")
	assert.Contains(t, formatted, "overall_score: 0.50")
	assert.Contains(t, formatted, "good_comments_ratio: 0.50")
	assert.Contains(t, formatted, "documentation_coverage: 0.33")
}

func TestCommentsAnalyzer_RealFile(t *testing.T) {
	analyzer := NewCommentsAnalyzer()

	// Parse a real Go file
	root := node.NewWithType(node.UASTFile)

	// Add a real comment
	comment := node.NewWithType(node.UASTComment)
	comment.Token = "// This is a real comment"
	comment.Pos = &node.Positions{
		StartLine: 1,
		EndLine:   1,
	}
	root.AddChild(comment)

	// Add a real function
	funcNode := node.NewWithType(node.UASTFunction)
	funcNode.Pos = &node.Positions{
		StartLine: 2,
		EndLine:   5,
	}
	name := node.NewWithType(node.UASTIdentifier)
	name.Token = "realFunction"
	name.Roles = []node.Role{node.RoleName}
	funcNode.AddChild(name)
	root.AddChild(funcNode)

	result, err := analyzer.Analyze(root)
	require.NoError(t, err)

	t.Logf("Result keys: %v", getKeys(result))

	if commentDetails, ok := result["comment_details"]; ok {
		t.Logf("Comment details type: %T", commentDetails)
		t.Logf("Comment details: %+v", commentDetails)
		if details, ok := commentDetails.([]map[string]interface{}); ok {
			t.Logf("Comment details length: %d", len(details))
		}
	}

	if functions, ok := result["functions"]; ok {
		t.Logf("Functions type: %T", functions)
		t.Logf("Functions: %+v", functions)
		if funcs, ok := functions.([]map[string]interface{}); ok {
			t.Logf("Functions length: %d", len(funcs))
		}
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("Full result: %s", string(resultJSON))
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
