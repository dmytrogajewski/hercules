package complexity

import (
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// NestingContext tracks the nesting level and type for cognitive complexity
type NestingContext struct {
	Level int
	Type  string
}

// CognitiveComplexityCalculator implements the SonarSource cognitive complexity algorithm
type CognitiveComplexityCalculator struct {
	complexity     int
	nestingStack   []NestingContext
	currentNesting int
}

// NewCognitiveComplexityCalculator creates a new cognitive complexity calculator
func NewCognitiveComplexityCalculator() *CognitiveComplexityCalculator {
	return &CognitiveComplexityCalculator{
		nestingStack:   make([]NestingContext, 0),
		currentNesting: 0,
	}
}

// CalculateCognitiveComplexity calculates cognitive complexity according to SonarSource specification
func (c *CognitiveComplexityCalculator) CalculateCognitiveComplexity(fn *node.Node) int {
	calculator := NewCognitiveComplexityCalculator()

	fn.VisitPreOrder(func(n *node.Node) {
		calculator.processNode(n)
	})

	return calculator.complexity
}

// processNode processes a single node and updates cognitive complexity
func (c *CognitiveComplexityCalculator) processNode(n *node.Node) {
	if c.isComplexityIncreasingElement(n) {
		c.addComplexityPoint(n)
	}

	if c.isNestingStart(n) {
		c.startNesting(n)
	}

	if c.isNestingEnd(n) {
		c.endNesting()
	}
}

// isComplexityIncreasingElement checks if a node increases cognitive complexity
func (c *CognitiveComplexityCalculator) isComplexityIncreasingElement(n *node.Node) bool {
	return c.isConditional(n) ||
		c.isLoop(n) ||
		c.isSwitchCase(n) ||
		c.isExceptionHandling(n) ||
		c.isLogicalOperator(n)
}

// isConditional checks if a node is a conditional statement
func (c *CognitiveComplexityCalculator) isConditional(n *node.Node) bool {
	switch n.Type {
	case node.UASTIf:
		return true
	}

	if n.HasAnyRole(node.RoleCondition, node.RoleBranch) {
		return true
	}

	return false
}

// isLoop checks if a node is a loop construct
func (c *CognitiveComplexityCalculator) isLoop(n *node.Node) bool {
	switch n.Type {
	case node.UASTLoop:
		return true
	}

	if n.HasAnyRole(node.RoleLoop) {
		return true
	}

	return false
}

// isSwitchCase checks if a node is a switch case
func (c *CognitiveComplexityCalculator) isSwitchCase(n *node.Node) bool {
	switch n.Type {
	case node.UASTCase:
		return true
	}

	if n.HasAnyRole(node.RoleBranch) {
		return true
	}

	return false
}

// isExceptionHandling checks if a node is exception handling
func (c *CognitiveComplexityCalculator) isExceptionHandling(n *node.Node) bool {
	switch n.Type {
	case node.UASTCatch, node.UASTFinally:
		return true
	}

	if n.HasAnyRole(node.RoleCatch, node.RoleFinally) {
		return true
	}

	return false
}

// isLogicalOperator checks if a node is a logical operator
func (c *CognitiveComplexityCalculator) isLogicalOperator(n *node.Node) bool {
	if n.Type != node.UASTBinaryOp {
		return false
	}

	operator := c.getOperator(n)
	return c.isLogicalOperatorToken(operator)
}

// isLogicalOperatorToken checks if a token represents a logical operator
func (c *CognitiveComplexityCalculator) isLogicalOperatorToken(operator string) bool {
	logicalOps := map[string]bool{
		"&&": true, "||": true,
		"and": true, "or": true,
		"AND": true, "OR": true,
	}
	return logicalOps[operator]
}

// getOperator extracts the operator from a binary operation node
func (c *CognitiveComplexityCalculator) getOperator(n *node.Node) string {
	if n.Token != "" {
		return n.Token
	}

	if op, ok := n.Props["operator"]; ok {
		return op
	}

	return c.getOperatorFromChildren(n)
}

// getOperatorFromChildren extracts operator from child nodes
func (c *CognitiveComplexityCalculator) getOperatorFromChildren(n *node.Node) string {
	for _, child := range n.Children {
		if child.HasAnyRole(node.RoleOperator) {
			return child.Token
		}
	}
	return ""
}

// isNestingStart checks if a node starts a nesting level
func (c *CognitiveComplexityCalculator) isNestingStart(n *node.Node) bool {
	switch n.Type {
	case node.UASTIf, node.UASTLoop, node.UASTSwitch, node.UASTTry, node.UASTCatch, node.UASTFinally:
		return true
	}

	if n.HasAnyRole(node.RoleCondition, node.RoleLoop, node.RoleTry, node.RoleCatch) {
		return true
	}

	return false
}

// isNestingEnd checks if a node ends a nesting level
func (c *CognitiveComplexityCalculator) isNestingEnd(n *node.Node) bool {
	switch n.Type {
	case node.UASTBlock, node.UASTFunction:
		return true
	}

	return false
}

// addComplexityPoint adds a complexity point with appropriate nesting penalty
func (c *CognitiveComplexityCalculator) addComplexityPoint(n *node.Node) {
	c.complexity++

	if c.isLogicalOperator(n) {
		return
	}

	if c.currentNesting > 0 {
		c.complexity += c.currentNesting
	}
}

// startNesting starts a new nesting level
func (c *CognitiveComplexityCalculator) startNesting(n *node.Node) {
	nestingType := c.getNestingType(n)

	context := NestingContext{
		Level: c.currentNesting,
		Type:  nestingType,
	}

	c.nestingStack = append(c.nestingStack, context)
	c.currentNesting++
}

// endNesting ends the current nesting level
func (c *CognitiveComplexityCalculator) endNesting() {
	if len(c.nestingStack) > 0 {
		c.nestingStack = c.nestingStack[:len(c.nestingStack)-1]
		if c.currentNesting > 0 {
			c.currentNesting--
		}
	}
}

// getNestingType determines the type of nesting for a node
func (c *CognitiveComplexityCalculator) getNestingType(n *node.Node) string {
	nestingType := c.getNestingTypeByNodeType(n)
	if nestingType != "unknown" {
		return nestingType
	}

	return c.getNestingTypeByRole(n)
}

// getNestingTypeByNodeType determines nesting type based on node type
func (c *CognitiveComplexityCalculator) getNestingTypeByNodeType(n *node.Node) string {
	nestingTypes := map[node.Type]string{
		node.UASTIf:      "if",
		node.UASTLoop:    "loop",
		node.UASTSwitch:  "switch",
		node.UASTTry:     "try",
		node.UASTCatch:   "catch",
		node.UASTFinally: "finally",
	}

	if nestingType, exists := nestingTypes[n.Type]; exists {
		return nestingType
	}
	return "unknown"
}

// getNestingTypeByRole determines nesting type based on node roles
func (c *CognitiveComplexityCalculator) getNestingTypeByRole(n *node.Node) string {
	roleToType := map[node.Role]string{
		node.RoleCondition: "if",
		node.RoleLoop:      "loop",
		node.RoleTry:       "try",
		node.RoleCatch:     "catch",
	}

	for role, nestingType := range roleToType {
		if n.HasAnyRole(role) {
			return nestingType
		}
	}

	return "unknown"
}

// GetComplexity returns the current cognitive complexity
func (c *CognitiveComplexityCalculator) GetComplexity() int {
	return c.complexity
}

// GetNestingLevel returns the current nesting level
func (c *CognitiveComplexityCalculator) GetNestingLevel() int {
	return c.currentNesting
}
