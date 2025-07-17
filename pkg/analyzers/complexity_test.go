package analyzers

import (
	"testing"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestCyclomaticComplexityAnalyzer_Name(t *testing.T) {
	analyzer := &CyclomaticComplexityAnalyzer{}
	assert.Equal(t, "cyclomatic_complexity", analyzer.Name())
}

func TestCyclomaticComplexityAnalyzer_Thresholds(t *testing.T) {
	analyzer := &CyclomaticComplexityAnalyzer{}
	thresholds := analyzer.Thresholds()

	assert.NotNil(t, thresholds)
	assert.Contains(t, thresholds, "complexity")

	complexityThresholds := thresholds["complexity"]
	assert.Equal(t, 1, complexityThresholds["green"])
	assert.Equal(t, 5, complexityThresholds["yellow"])
	assert.Equal(t, 10, complexityThresholds["red"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_NilRoot(t *testing.T) {
	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(nil)

	assert.NoError(t, err)
	assert.Equal(t, 0, result["total_complexity"])
	assert.Equal(t, 0, result["function_count"])
	assert.Equal(t, map[string]int{}, result["functions"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_SimpleFunction(t *testing.T) {
	// Create a simple function with no decision points
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}

	nameNode := node.NewWithType("Identifier")
	nameNode.Token = "simpleFunction"
	nameNode.Roles = []node.Role{"Name"}

	body := node.NewWithType("Block")
	body.Roles = []node.Role{"Body"}

	function.AddChild(nameNode)
	function.AddChild(body)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 1, result["total_complexity"]) // Base complexity of 1
	assert.Equal(t, 1, result["function_count"])

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 1, functions["simpleFunction"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_FunctionWithIf(t *testing.T) {
	// Create a function with an if statement
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}

	nameNode := node.NewWithType("Identifier")
	nameNode.Token = "functionWithIf"
	nameNode.Roles = []node.Role{"Name"}

	body := node.NewWithType("Block")
	body.Roles = []node.Role{"Body"}

	// Create if statement using canonical type
	ifStmt := node.NewWithType("If")
	ifStmt.Roles = []node.Role{"Condition"}

	body.AddChild(ifStmt)
	function.AddChild(nameNode)
	function.AddChild(body)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 2, result["total_complexity"]) // Base 1 + 1 if

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 2, functions["functionWithIf"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_FunctionWithMultipleDecisionPoints(t *testing.T) {
	// Create a function with multiple decision points
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}

	nameNode := node.NewWithType("Identifier")
	nameNode.Token = "complexFunction"
	nameNode.Roles = []node.Role{"Name"}

	// Create body node
	body := node.NewWithType("Block")
	body.Roles = []node.Role{"Body"}

	// Create if statement using canonical type
	ifStmt := node.NewWithType("If")
	ifStmt.Roles = []node.Role{"Condition"}

	// Create while loop using canonical type
	whileStmt := node.NewWithType("Loop")
	whileStmt.Props = map[string]string{"kind": "while"}

	// Create switch statement using canonical type
	switchStmt := node.NewWithType("Switch")

	// Create case statement using canonical type
	caseStmt := node.NewWithType("Case")
	caseStmt.Roles = []node.Role{"Branch"}

	// Create logical operator using canonical type
	logicalOp := node.NewWithType("BinaryOp")
	logicalOp.Props = map[string]string{"operator": "&&"}

	// Assemble the tree
	function.AddChild(nameNode)
	function.AddChild(body)
	body.AddChild(ifStmt)
	body.AddChild(whileStmt)
	body.AddChild(switchStmt)
	switchStmt.AddChild(caseStmt)
	body.AddChild(logicalOp)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 6, result["total_complexity"]) // Base 1 + 5 decision points

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 6, functions["complexFunction"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_MultipleFunctions(t *testing.T) {
	// Create a module with multiple functions
	root := node.NewWithType("Module")

	// Function 1: simple
	function1 := node.NewWithType("Function")
	function1.Roles = []node.Role{"Function"}
	name1 := node.NewWithType("Identifier")
	name1.Token = "function1"
	name1.Roles = []node.Role{"Name"}
	function1.AddChild(name1)

	// Function 2: with if
	function2 := node.NewWithType("Function")
	function2.Roles = []node.Role{"Function"}
	name2 := node.NewWithType("Identifier")
	name2.Token = "function2"
	name2.Roles = []node.Role{"Name"}
	ifStmt := node.NewWithType("If")
	ifStmt.Roles = []node.Role{"Condition"}
	function2.AddChild(name2)
	function2.AddChild(ifStmt)

	// Function 3: with multiple decision points
	function3 := node.NewWithType("Function")
	function3.Roles = []node.Role{"Function"}
	name3 := node.NewWithType("Identifier")
	name3.Token = "function3"
	name3.Roles = []node.Role{"Name"}
	whileStmt := node.NewWithType("Loop")
	whileStmt.Props = map[string]string{"kind": "while"}
	switchStmt := node.NewWithType("Switch")
	function3.AddChild(name3)
	function3.AddChild(whileStmt)
	function3.AddChild(switchStmt)

	root.AddChild(function1)
	root.AddChild(function2)
	root.AddChild(function3)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 6, result["total_complexity"]) // 1 + 2 + 3
	assert.Equal(t, 3, result["function_count"])

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 1, functions["function1"])
	assert.Equal(t, 2, functions["function2"])
	assert.Equal(t, 3, functions["function3"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_Operators(t *testing.T) {
	// Test various operators that create decision points
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}
	name := node.NewWithType("Identifier")
	name.Token = "operatorTest"
	name.Roles = []node.Role{"Name"}
	function.AddChild(name)

	// Test logical operators using canonical types
	logicalAnd := node.NewWithType("BinaryOp")
	logicalAnd.Props = map[string]string{"operator": "&&"}

	logicalOr := node.NewWithType("BinaryOp")
	logicalOr.Props = map[string]string{"operator": "||"}

	// Test comparison operators
	comparison := node.NewWithType("BinaryOp")
	comparison.Props = map[string]string{"operator": "=="}

	// Test unary operator
	unaryOp := node.NewWithType("UnaryOp")
	unaryOp.Props = map[string]string{"operator": "!"}

	// Test ternary operator (not canonical, should be ignored)
	ternary := node.NewWithType("Ternary")

	function.AddChild(logicalAnd)
	function.AddChild(logicalOr)
	function.AddChild(comparison)
	function.AddChild(unaryOp)
	function.AddChild(ternary)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 4, result["total_complexity"]) // Base 1 + 3 logical operators (&&, ||, !)

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 4, functions["operatorTest"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_ExceptionHandling(t *testing.T) {
	// Test exception handling constructs
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}
	name := node.NewWithType("Identifier")
	name.Token = "exceptionTest"
	name.Roles = []node.Role{"Name"}
	function.AddChild(name)

	// Test try-catch using canonical types
	tryStmt := node.NewWithType("Try")

	catchStmt := node.NewWithType("Catch")

	throwStmt := node.NewWithType("Throw")

	function.AddChild(tryStmt)
	function.AddChild(catchStmt)
	function.AddChild(throwStmt)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 4, result["total_complexity"]) // Base 1 + 3 exception constructs

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 4, functions["exceptionTest"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_AnonymousFunction(t *testing.T) {
	// Test function without a name
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}

	// No name node, should default to "anonymous"

	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 1, result["total_complexity"]) // Base complexity
	assert.Equal(t, 1, result["function_count"])

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 1, functions["anonymous"])
}

func TestCyclomaticComplexityAnalyzer_Analyze_FunctionWithNameInProps(t *testing.T) {
	// Test function with name in props
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}
	function.Props = map[string]string{"name": "functionInProps"}

	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	assert.Equal(t, 1, result["total_complexity"])
	assert.Equal(t, 1, result["function_count"])

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 1, functions["functionInProps"])
}

func TestCyclomaticComplexityAnalyzer_DSLQueryWorking(t *testing.T) {
	// Test that the DSL query is actually working by creating a complex nested structure
	// that would be difficult to traverse manually
	root := node.NewWithType("Module")

	function := node.NewWithType("Function")
	function.Roles = []node.Role{"Function"}
	name := node.NewWithType("Identifier")
	name.Token = "complexFunction"
	name.Roles = []node.Role{"Name"}
	function.AddChild(name)

	// Create a deeply nested structure with decision points
	body := node.NewWithType("Block")

	// Level 1: If statement
	ifStmt := node.NewWithType("If")
	ifStmt.Roles = []node.Role{"Condition"}

	// Level 2: Nested loop inside if
	nestedLoop := node.NewWithType("Loop")
	nestedLoop.Props = map[string]string{"kind": "while"}

	// Level 3: Switch inside nested loop
	nestedSwitch := node.NewWithType("Switch")

	// Level 4: Case inside switch
	caseStmt := node.NewWithType("Case")
	caseStmt.Roles = []node.Role{"Branch"}

	// Level 5: Try-catch inside case
	tryStmt := node.NewWithType("Try")
	catchStmt := node.NewWithType("Catch")
	throwStmt := node.NewWithType("Throw")

	// Level 6: Logical operators inside catch
	logicalAnd := node.NewWithType("BinaryOp")
	logicalAnd.Props = map[string]string{"operator": "&&"}

	logicalOr := node.NewWithType("BinaryOp")
	logicalOr.Props = map[string]string{"operator": "||"}

	unaryOp := node.NewWithType("UnaryOp")
	unaryOp.Props = map[string]string{"operator": "!"}

	// Assemble the deeply nested structure
	function.AddChild(body)
	body.AddChild(ifStmt)
	ifStmt.AddChild(nestedLoop)
	nestedLoop.AddChild(nestedSwitch)
	nestedSwitch.AddChild(caseStmt)
	caseStmt.AddChild(tryStmt)
	caseStmt.AddChild(catchStmt)
	caseStmt.AddChild(throwStmt)
	catchStmt.AddChild(logicalAnd)
	catchStmt.AddChild(logicalOr)
	catchStmt.AddChild(unaryOp)
	root.AddChild(function)

	analyzer := &CyclomaticComplexityAnalyzer{}
	result, err := analyzer.Analyze(root)

	assert.NoError(t, err)
	// Should find: If, Loop, Switch, Case, Try, Catch, Throw, BinaryOp(&&), BinaryOp(||), UnaryOp(!)
	// Base complexity 1 + 10 decision points = 11
	assert.Equal(t, 11, result["total_complexity"])

	functions := result["functions"].(map[string]int)
	assert.Equal(t, 11, functions["complexFunction"])
}

func TestCyclomaticComplexityAnalyzer_CppComplexityTest(t *testing.T) {
	// Test the C++ complexity analyzer with the test cases from cpp-complexity-test.yaml
	analyzer := &CyclomaticComplexityAnalyzer{}

	// Test case 1: Simple function
	simpleFunction := node.NewWithType("Function")
	simpleFunction.Roles = []node.Role{"Function"}
	name1 := node.NewWithType("Identifier")
	name1.Token = "simple_function"
	name1.Roles = []node.Role{"Name"}
	simpleFunction.AddChild(name1)

	result1, err := analyzer.Analyze(simpleFunction)
	assert.NoError(t, err)
	assert.Equal(t, 1, result1["total_complexity"]) // Base complexity

	// Test case 2: Function with if statement
	functionWithIf := node.NewWithType("Function")
	functionWithIf.Roles = []node.Role{"Function"}
	name2 := node.NewWithType("Identifier")
	name2.Token = "function_with_if"
	name2.Roles = []node.Role{"Name"}
	ifStmt := node.NewWithType("If")
	ifStmt.Roles = []node.Role{"Condition"}
	functionWithIf.AddChild(name2)
	functionWithIf.AddChild(ifStmt)

	result2, err := analyzer.Analyze(functionWithIf)
	assert.NoError(t, err)
	assert.Equal(t, 2, result2["total_complexity"]) // Base 1 + If 1

	// Test case 3: Function with while loop
	functionWithWhile := node.NewWithType("Function")
	functionWithWhile.Roles = []node.Role{"Function"}
	name3 := node.NewWithType("Identifier")
	name3.Token = "function_with_while"
	name3.Roles = []node.Role{"Name"}
	whileStmt := node.NewWithType("Loop")
	whileStmt.Props = map[string]string{"kind": "while"}
	functionWithWhile.AddChild(name3)
	functionWithWhile.AddChild(whileStmt)

	result3, err := analyzer.Analyze(functionWithWhile)
	assert.NoError(t, err)
	assert.Equal(t, 2, result3["total_complexity"]) // Base 1 + Loop 1

	// Test case 4: Function with switch statement
	functionWithSwitch := node.NewWithType("Function")
	functionWithSwitch.Roles = []node.Role{"Function"}
	name4 := node.NewWithType("Identifier")
	name4.Token = "function_with_switch"
	name4.Roles = []node.Role{"Name"}
	switchStmt := node.NewWithType("Switch")
	case1 := node.NewWithType("Case")
	case2 := node.NewWithType("Case")
	case3 := node.NewWithType("Case")
	switchStmt.AddChild(case1)
	switchStmt.AddChild(case2)
	switchStmt.AddChild(case3)
	functionWithSwitch.AddChild(name4)
	functionWithSwitch.AddChild(switchStmt)

	result4, err := analyzer.Analyze(functionWithSwitch)
	assert.NoError(t, err)
	assert.Equal(t, 5, result4["total_complexity"]) // Base 1 + Switch 1 + 3 Cases

	// Test case 5: Function with try-catch
	functionWithTryCatch := node.NewWithType("Function")
	functionWithTryCatch.Roles = []node.Role{"Function"}
	name5 := node.NewWithType("Identifier")
	name5.Token = "function_with_try_catch"
	name5.Roles = []node.Role{"Name"}
	tryStmt := node.NewWithType("Try")
	catchStmt := node.NewWithType("Catch")
	throwStmt := node.NewWithType("Throw")
	functionWithTryCatch.AddChild(name5)
	functionWithTryCatch.AddChild(tryStmt)
	functionWithTryCatch.AddChild(catchStmt)
	functionWithTryCatch.AddChild(throwStmt)

	result5, err := analyzer.Analyze(functionWithTryCatch)
	assert.NoError(t, err)
	assert.Equal(t, 4, result5["total_complexity"]) // Base 1 + Try 1 + Catch 1 + Throw 1

	// Test case 6: Function with logical operators
	functionWithLogicalOps := node.NewWithType("Function")
	functionWithLogicalOps.Roles = []node.Role{"Function"}
	name6 := node.NewWithType("Identifier")
	name6.Token = "function_with_logical_ops"
	name6.Roles = []node.Role{"Name"}

	// Add if statements with logical operators
	ifStmt1 := node.NewWithType("If")
	ifStmt1.Roles = []node.Role{"Condition"}
	logicalAnd := node.NewWithType("BinaryOp")
	logicalAnd.Props = map[string]string{"operator": "&&"}
	ifStmt1.AddChild(logicalAnd)

	ifStmt2 := node.NewWithType("If")
	ifStmt2.Roles = []node.Role{"Condition"}
	logicalOr := node.NewWithType("BinaryOp")
	logicalOr.Props = map[string]string{"operator": "||"}
	ifStmt2.AddChild(logicalOr)

	unaryOp := node.NewWithType("UnaryOp")
	unaryOp.Props = map[string]string{"operator": "!"}

	functionWithLogicalOps.AddChild(name6)
	functionWithLogicalOps.AddChild(ifStmt1)
	functionWithLogicalOps.AddChild(ifStmt2)
	functionWithLogicalOps.AddChild(unaryOp)

	result6, err := analyzer.Analyze(functionWithLogicalOps)
	assert.NoError(t, err)
	assert.Equal(t, 6, result6["total_complexity"]) // Base 1 + 2 If + 2 logical operators + 1 unary operator
}
