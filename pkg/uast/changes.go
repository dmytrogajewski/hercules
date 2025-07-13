package uast

import (
	"fmt"

	"github.com/dmytrogajewski/hercules/pkg/uast/internal/node"
)

// DetectChanges detects structural changes between two UAST nodes.
// It returns a slice of Change objects describing added, removed, and modified nodes.
// Now uses the final optimized implementation with ultra-fast integer keys.
//
// Example:
//
//	changes := uast.DetectChanges(before, after)
//	for _, c := range changes {
//	    fmt.Println(c.Type)
//	}
func DetectChanges(before, after *node.Node) []Change {
	return finalOptimizedDetectChangesInt(before, after)
}

// FilterChangesByType filters the given changes by their ChangeType.
// Returns a slice of changes matching the specified type.
func FilterChangesByType(changes []Change, changeType ChangeType) []Change {
	var filtered []Change
	for _, change := range changes {
		if isChangeOfType(change, changeType) {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// FilterChangesByNodeType filters the given changes by the type of nodes involved.
// Returns a slice of changes where either Before or After node matches nodeType.
func FilterChangesByNodeType(changes []Change, nodeType string) []Change {
	var filtered []Change
	for _, change := range changes {
		if hasNodeType(change, nodeType) {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// CountChangesByType counts the number of changes for each ChangeType.
// Returns a map from ChangeType to count.
func CountChangesByType(changes []Change) map[ChangeType]int {
	counts := make(map[ChangeType]int)
	for _, change := range changes {
		counts[change.Type]++
	}
	return counts
}

// GetModifiedNodes returns all nodes that were modified in the given changes.
func GetModifiedNodes(changes []Change) []*node.Node {
	var nodes []*node.Node
	for _, change := range changes {
		if isModifiedChangeWithAfter(change) {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetAddedNodes returns all nodes that were added in the given changes.
func GetAddedNodes(changes []Change) []*node.Node {
	var nodes []*node.Node
	for _, change := range changes {
		if isAddedChangeWithAfter(change) {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetRemovedNodes returns all nodes that were removed in the given changes.
func GetRemovedNodes(changes []Change) []*node.Node {
	var nodes []*node.Node
	for _, change := range changes {
		if isRemovedChangeWithBefore(change) {
			nodes = append(nodes, change.Before)
		}
	}
	return nodes
}

// isNodeAdded checks if a node was added
func isNodeAdded(before, after *node.Node) bool {
	return before == nil && after != nil
}

// isNodeRemoved checks if a node was removed
func isNodeRemoved(before, after *node.Node) bool {
	return before != nil && after == nil
}

// isNoChange checks if there are no changes
func isNoChange(before, after *node.Node) bool {
	return before == nil && after == nil
}

// appendAddedChange appends an added change to the changes slice
func appendAddedChange(changes []Change, after *node.Node) []Change {
	return append(changes, Change{
		Before: nil,
		After:  after,
		Type:   ChangeAdded,
	})
}

// appendRemovedChange appends a removed change to the changes slice
func appendRemovedChange(changes []Change, before *node.Node) []Change {
	return append(changes, Change{
		Before: before,
		After:  nil,
		Type:   ChangeRemoved,
	})
}

// appendModifiedChange appends a modified change to the changes slice
func appendModifiedChange(changes []Change, before, after *node.Node) []Change {
	return append(changes, Change{
		Before: before,
		After:  after,
		Type:   ChangeModified,
	})
}

// appendChildrenChanges appends children changes to the changes slice
func appendChildrenChanges(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children
	childrenChanges := detectChildrenChanges(beforeChildren, afterChildren)
	return append(changes, childrenChanges...)
}

// isNodeModified checks if a node has been modified
func isNodeModified(before, after *node.Node) bool {
	if hasDifferentType(before, after) {
		return true
	}

	if hasDifferentToken(before, after) {
		return true
	}

	if hasSignificantPositionChange(before, after) {
		return true
	}

	if hasDifferentStringRepresentation(before, after) {
		return true
	}

	return false
}

// hasDifferentType checks if nodes have different types
func hasDifferentType(before, after *node.Node) bool {
	return before.Type != after.Type
}

// hasDifferentToken checks if nodes have different tokens
func hasDifferentToken(before, after *node.Node) bool {
	return before.Token != after.Token
}

// hasSignificantPositionChange checks if position change is significant
func hasSignificantPositionChange(before, after *node.Node) bool {
	beforePos := before.Pos
	afterPos := after.Pos

	if isOnePositionNil(beforePos, afterPos) {
		return beforePos != afterPos
	}

	return hasSignificantLineOrColumnChange(beforePos, afterPos)
}

// isOnePositionNil checks if one of the positions is nil
func isOnePositionNil(beforePos, afterPos *node.Positions) bool {
	return beforePos == nil || afterPos == nil
}

// hasSignificantLineOrColumnChange checks if line or column change is significant
func hasSignificantLineOrColumnChange(beforePos, afterPos *node.Positions) bool {
	lineDiff := abs(beforePos.StartLine - afterPos.StartLine)
	columnDiff := abs(beforePos.StartCol - afterPos.StartCol)
	return lineDiff > 1 || columnDiff > 10
}

// hasDifferentStringRepresentation checks if nodes have different string representations
func hasDifferentStringRepresentation(before, after *node.Node) bool {
	// Use optimized comparison instead of expensive JSON marshaling
	return hasDifferentKeyProperties(before, after)
}

// detectChildrenChanges detects changes in child nodes
func detectChildrenChanges(beforeChildren, afterChildren []*node.Node) []Change {
	var changes []Change

	beforeMap := buildNodeMap(beforeChildren)
	afterMap := buildNodeMap(afterChildren)

	changes = appendModifiedChildren(changes, beforeMap, afterMap)
	changes = appendRemovedChildren(changes, beforeMap, afterMap)
	changes = appendAddedChildren(changes, beforeMap, afterMap)

	return changes
}

// buildNodeMap builds a map of nodes by key
func buildNodeMap(children []*node.Node) map[string]*node.Node {
	nodeMap := make(map[string]*node.Node)
	for _, child := range children {
		key := getNodeKey(child)
		nodeMap[key] = child
	}
	return nodeMap
}

// appendModifiedChildren appends modified children changes
func appendModifiedChildren(changes []Change, beforeMap, afterMap map[string]*node.Node) []Change {
	for key, beforeChild := range beforeMap {
		if afterChild, exists := afterMap[key]; exists {
			if isNodeModified(beforeChild, afterChild) {
				changes = appendModifiedChange(changes, beforeChild, afterChild)
			}
		}
	}
	return changes
}

// appendRemovedChildren appends removed children changes
func appendRemovedChildren(changes []Change, beforeMap, afterMap map[string]*node.Node) []Change {
	for key, beforeChild := range beforeMap {
		if isNodeNotInAfterMap(key, afterMap) {
			changes = appendRemovedChange(changes, beforeChild)
		}
	}
	return changes
}

// appendAddedChildren appends added children changes
func appendAddedChildren(changes []Change, beforeMap, afterMap map[string]*node.Node) []Change {
	for key, afterChild := range afterMap {
		if isNodeNotInBeforeMap(key, beforeMap) {
			changes = appendAddedChange(changes, afterChild)
		}
	}
	return changes
}

// isNodeNotInAfterMap checks if a node key is not in the after map
func isNodeNotInAfterMap(key string, afterMap map[string]*node.Node) bool {
	_, exists := afterMap[key]
	return !exists
}

// isNodeNotInBeforeMap checks if a node key is not in the before map
func isNodeNotInBeforeMap(key string, beforeMap map[string]*node.Node) bool {
	_, exists := beforeMap[key]
	return !exists
}

// getNodeKey returns a unique key for a node
func getNodeKey(node *node.Node) string {
	// Use ultra-fast integer key generation for better performance
	return fmt.Sprintf("%d", ultraFastGetNodeKey(node))
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// isChangeOfType checks if a change is of the specified type
func isChangeOfType(change Change, changeType ChangeType) bool {
	return change.Type == changeType
}

// hasNodeType checks if a change involves a node of the specified type
func hasNodeType(change Change, nodeType string) bool {
	return (change.Before != nil && change.Before.Type == nodeType) ||
		(change.After != nil && change.After.Type == nodeType)
}

// isModifiedChangeWithAfter checks if a change is modified and has an after node
func isModifiedChangeWithAfter(change Change) bool {
	return change.Type == ChangeModified && change.After != nil
}

// isAddedChangeWithAfter checks if a change is added and has an after node
func isAddedChangeWithAfter(change Change) bool {
	return change.Type == ChangeAdded && change.After != nil
}

// isRemovedChangeWithBefore checks if a change is removed and has a before node
func isRemovedChangeWithBefore(change Change) bool {
	return change.Type == ChangeRemoved && change.Before != nil
}

// Final optimized change detection with minimal allocations
func finalOptimizedDetectChanges(before, after *node.Node) []Change {
	var changes []Change

	if isNoChange(before, after) {
		return changes
	}

	if isNodeAdded(before, after) {
		changes = appendAddedChange(changes, after)
		return changes
	}

	if isNodeRemoved(before, after) {
		changes = appendRemovedChange(changes, before)
		return changes
	}

	if isNodeModifiedFinalOptimized(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChangesFinalOptimized(changes, before, after)

	return changes
}

// Final optimized node modification check
func isNodeModifiedFinalOptimized(before, after *node.Node) bool {
	// Fast path checks first
	if before.Type != after.Type {
		return true
	}

	if before.Token != after.Token {
		return true
	}

	// Skip position check for performance
	// Skip string comparison entirely

	// Only check roles if they exist
	if len(before.Roles) != len(after.Roles) {
		return true
	}

	for i := range before.Roles {
		if before.Roles[i] != after.Roles[i] {
			return true
		}
	}

	return false
}

// Final optimized children changes detection
func appendChildrenChangesFinalOptimized(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children

	if len(beforeChildren) == 0 && len(afterChildren) == 0 {
		return changes
	}

	return appendChildrenChangesSimple(changes, beforeChildren, afterChildren)
}

// Ultra-fast node key generation using integer hash
func ultraFastGetNodeKey(node *node.Node) int64 {
	if node == nil {
		return 0
	}

	// Use a fast hash of the node's key properties
	var hash int64 = 5381
	hash = ((hash << 5) + hash) + int64(len(node.Type))
	hash = ((hash << 5) + hash) + int64(len(node.Token))
	hash = ((hash << 5) + hash) + int64(len(node.Children))
	hash = ((hash << 5) + hash) + int64(len(node.Roles))

	// Add hash of type string
	for _, c := range node.Type {
		hash = ((hash << 5) + hash) + int64(c)
	}

	// Add hash of token string
	for _, c := range node.Token {
		hash = ((hash << 5) + hash) + int64(c)
	}

	return hash
}

// Final optimized change detection using ultra-fast integer keys
func finalOptimizedDetectChangesInt(before, after *node.Node) []Change {
	var changes []Change

	if isNoChange(before, after) {
		return changes
	}

	if isNodeAdded(before, after) {
		changes = appendAddedChange(changes, after)
		return changes
	}

	if isNodeRemoved(before, after) {
		changes = appendRemovedChange(changes, before)
		return changes
	}

	if isNodeModifiedFinalOptimized(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChangesFinalOptimizedInt(changes, before, after)

	return changes
}

// Final optimized children changes with ultra-fast integer keys
func appendChildrenChangesFinalOptimizedInt(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children

	if len(beforeChildren) == 0 && len(afterChildren) == 0 {
		return changes
	}

	if len(beforeChildren) == 0 {
		return appendAddedChange(changes, after)
	}

	if len(afterChildren) == 0 {
		return appendRemovedChange(changes, before)
	}

	// Use ultra-fast integer key maps
	beforeMap := buildNodeMapUltraFast(beforeChildren)
	afterMap := buildNodeMapUltraFast(afterChildren)

	changes = appendModifiedChildrenInt(changes, beforeMap, afterMap)
	changes = appendRemovedChildrenInt(changes, beforeMap, afterMap)
	changes = appendAddedChildrenInt(changes, beforeMap, afterMap)

	return changes
}

// Build node map using ultra-fast integer keys
func buildNodeMapUltraFast(children []*node.Node) map[int64]*node.Node {
	if len(children) == 0 {
		return make(map[int64]*node.Node)
	}

	nodeMap := make(map[int64]*node.Node, len(children))
	for _, child := range children {
		key := ultraFastGetNodeKey(child)
		nodeMap[key] = child
	}
	return nodeMap
}

// Optimized node key generation without string concatenation
func optimizedGetNodeKey(node *node.Node) string {
	if node == nil {
		return ""
	}

	// Use a fast hash-based approach
	var hash uint32 = 5381
	hash = ((hash << 5) + hash) + uint32(len(node.Type))
	hash = ((hash << 5) + hash) + uint32(len(node.Token))
	hash = ((hash << 5) + hash) + uint32(len(node.Children))
	hash = ((hash << 5) + hash) + uint32(len(node.Roles))

	// Add hash of type string
	for _, c := range node.Type {
		hash = ((hash << 5) + hash) + uint32(c)
	}

	// Add hash of token string
	for _, c := range node.Token {
		hash = ((hash << 5) + hash) + uint32(c)
	}

	return fmt.Sprintf("%d", hash)
}

// Zero-allocation change detection with ultra-fast integer keys
func zeroAllocationDetectChangesInt(before, after *node.Node) []Change {
	var changes []Change

	if isNoChange(before, after) {
		return changes
	}

	if isNodeAdded(before, after) {
		changes = appendAddedChange(changes, after)
		return changes
	}

	if isNodeRemoved(before, after) {
		changes = appendRemovedChange(changes, before)
		return changes
	}

	if isNodeModifiedZeroAllocation(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChangesZeroAllocationInt(changes, before, after)

	return changes
}

// Zero-allocation node modification check
func isNodeModifiedZeroAllocation(before, after *node.Node) bool {
	// Fast path checks first
	if before.Type != after.Type {
		return true
	}

	if before.Token != after.Token {
		return true
	}

	// Skip position check for performance
	// Skip string comparison entirely

	// Only check roles if they exist
	if len(before.Roles) != len(after.Roles) {
		return true
	}

	for i := range before.Roles {
		if before.Roles[i] != after.Roles[i] {
			return true
		}
	}

	return false
}

// Zero-allocation children changes with ultra-fast integer keys
func appendChildrenChangesZeroAllocationInt(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children

	if len(beforeChildren) == 0 && len(afterChildren) == 0 {
		return changes
	}

	if len(beforeChildren) == 0 {
		return appendAddedChange(changes, after)
	}

	if len(afterChildren) == 0 {
		return appendRemovedChange(changes, before)
	}

	// Use ultra-fast integer key maps
	beforeMap := buildNodeMapUltraFast(beforeChildren)
	afterMap := buildNodeMapUltraFast(afterChildren)

	changes = appendModifiedChildrenInt(changes, beforeMap, afterMap)
	changes = appendRemovedChildrenInt(changes, beforeMap, afterMap)
	changes = appendAddedChildrenInt(changes, beforeMap, afterMap)

	return changes
}

// Optimized children changes detection with reduced allocations
func optimizedDetectChanges(before, after *node.Node) []Change {
	var changes []Change

	if isNoChange(before, after) {
		return changes
	}

	if isNodeAdded(before, after) {
		changes = appendAddedChange(changes, after)
		return changes
	}

	if isNodeRemoved(before, after) {
		changes = appendRemovedChange(changes, before)
		return changes
	}

	if isNodeModifiedOptimized(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChangesOptimized(changes, before, after)

	return changes
}

// Optimized node modification check without string comparison
func isNodeModifiedOptimized(before, after *node.Node) bool {
	if hasDifferentType(before, after) {
		return true
	}

	if hasDifferentToken(before, after) {
		return true
	}

	if hasSignificantPositionChange(before, after) {
		return true
	}

	if hasDifferentKeyProperties(before, after) {
		return true
	}

	return false
}

// Check for different key properties efficiently
func hasDifferentKeyProperties(before, after *node.Node) bool {
	if len(before.Props) != len(after.Props) {
		return true
	}

	for k, v := range before.Props {
		if after.Props[k] != v {
			return true
		}
	}

	return false
}

// Optimized children changes detection with reduced allocations
func appendChildrenChangesOptimized(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children

	if len(beforeChildren) == 0 && len(afterChildren) == 0 {
		return changes
	}

	if len(beforeChildren) == 0 {
		return appendAddedChange(changes, after)
	}

	if len(afterChildren) == 0 {
		return appendRemovedChange(changes, before)
	}

	// Use optimized map building for larger lists
	if len(beforeChildren) > 10 || len(afterChildren) > 10 {
		beforeMap := buildNodeMapOptimized(beforeChildren)
		afterMap := buildNodeMapOptimized(afterChildren)

		changes = appendModifiedChildren(changes, beforeMap, afterMap)
		changes = appendRemovedChildren(changes, beforeMap, afterMap)
		changes = appendAddedChildren(changes, beforeMap, afterMap)
		return changes
	}

	// For small lists, use simple comparison
	return appendChildrenChangesSimple(changes, beforeChildren, afterChildren)
}

// Simple children changes for small lists
func appendChildrenChangesSimple(changes []Change, beforeChildren, afterChildren []*node.Node) []Change {
	maxLen := len(beforeChildren)
	if len(afterChildren) > maxLen {
		maxLen = len(afterChildren)
	}

	for i := 0; i < maxLen; i++ {
		var beforeChild, afterChild *node.Node
		if i < len(beforeChildren) {
			beforeChild = beforeChildren[i]
		}
		if i < len(afterChildren) {
			afterChild = afterChildren[i]
		}

		if isNodeModifiedOptimized(beforeChild, afterChild) {
			changes = append(changes, Change{
				Before: beforeChild,
				After:  afterChild,
				Type:   ChangeModified,
			})
		}
	}

	return changes
}

// Optimized node map building with better memory management
func buildNodeMapOptimized(children []*node.Node) map[string]*node.Node {
	if len(children) == 0 {
		return make(map[string]*node.Node)
	}

	nodeMap := make(map[string]*node.Node, len(children))
	for _, child := range children {
		key := optimizedGetNodeKey(child)
		nodeMap[key] = child
	}
	return nodeMap
}

// Optimized node key with integer-based approach for better performance
func optimizedGetNodeKeyInt(node *node.Node) int64 {
	if node == nil {
		return 0
	}

	// Use a fast hash of the node's key properties
	var hash int64 = 5381
	hash = ((hash << 5) + hash) + int64(len(node.Type))
	hash = ((hash << 5) + hash) + int64(len(node.Token))
	hash = ((hash << 5) + hash) + int64(len(node.Children))
	hash = ((hash << 5) + hash) + int64(len(node.Roles))

	// Add hash of type string
	for _, c := range node.Type {
		hash = ((hash << 5) + hash) + int64(c)
	}

	// Add hash of token string
	for _, c := range node.Token {
		hash = ((hash << 5) + hash) + int64(c)
	}

	return hash
}

// Optimized change detection using integer keys
func optimizedDetectChangesInt(before, after *node.Node) []Change {
	var changes []Change

	if isNoChange(before, after) {
		return changes
	}

	if isNodeAdded(before, after) {
		changes = appendAddedChange(changes, after)
		return changes
	}

	if isNodeRemoved(before, after) {
		changes = appendRemovedChange(changes, before)
		return changes
	}

	if isNodeModifiedOptimized(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChangesOptimizedInt(changes, before, after)

	return changes
}

// Optimized children changes with integer keys
func appendChildrenChangesOptimizedInt(changes []Change, before, after *node.Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children

	if len(beforeChildren) == 0 && len(afterChildren) == 0 {
		return changes
	}

	if len(beforeChildren) == 0 {
		return appendAddedChange(changes, after)
	}

	if len(afterChildren) == 0 {
		return appendRemovedChange(changes, before)
	}

	// Use integer key maps
	beforeMap := buildNodeMapInt(beforeChildren)
	afterMap := buildNodeMapInt(afterChildren)

	changes = appendModifiedChildrenInt(changes, beforeMap, afterMap)
	changes = appendRemovedChildrenInt(changes, beforeMap, afterMap)
	changes = appendAddedChildrenInt(changes, beforeMap, afterMap)

	return changes
}

// Build node map using integer keys
func buildNodeMapInt(children []*node.Node) map[int64]*node.Node {
	if len(children) == 0 {
		return make(map[int64]*node.Node)
	}

	nodeMap := make(map[int64]*node.Node, len(children))
	for _, child := range children {
		key := optimizedGetNodeKeyInt(child)
		nodeMap[key] = child
	}
	return nodeMap
}

// Append modified children using integer keys
func appendModifiedChildrenInt(changes []Change, beforeMap, afterMap map[int64]*node.Node) []Change {
	for key, beforeChild := range beforeMap {
		if afterChild, exists := afterMap[key]; exists {
			if isNodeModifiedOptimized(beforeChild, afterChild) {
				changes = append(changes, Change{
					Before: beforeChild,
					After:  afterChild,
					Type:   ChangeModified,
				})
			}
		}
	}
	return changes
}

// Append removed children using integer keys
func appendRemovedChildrenInt(changes []Change, beforeMap, afterMap map[int64]*node.Node) []Change {
	for key, beforeChild := range beforeMap {
		if _, exists := afterMap[key]; !exists {
			changes = append(changes, Change{
				Before: beforeChild,
				After:  nil,
				Type:   ChangeRemoved,
			})
		}
	}
	return changes
}

// Append added children using integer keys
func appendAddedChildrenInt(changes []Change, beforeMap, afterMap map[int64]*node.Node) []Change {
	for key, afterChild := range afterMap {
		if _, exists := beforeMap[key]; !exists {
			changes = append(changes, Change{
				Before: nil,
				After:  afterChild,
				Type:   ChangeAdded,
			})
		}
	}
	return changes
}
