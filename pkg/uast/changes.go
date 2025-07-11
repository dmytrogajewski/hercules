package uast

// DetectChanges detects structural changes between two UAST nodes.
// It returns a slice of Change objects describing added, removed, and modified nodes.
//
// Example:
//
//	changes := uast.DetectChanges(before, after)
//	for _, c := range changes {
//	    fmt.Println(c.Type)
//	}
func DetectChanges(before, after *Node) []Change {
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

	if isNodeModified(before, after) {
		changes = appendModifiedChange(changes, before, after)
	}

	changes = appendChildrenChanges(changes, before, after)

	return changes
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
func GetModifiedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if isModifiedChangeWithAfter(change) {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetAddedNodes returns all nodes that were added in the given changes.
func GetAddedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if isAddedChangeWithAfter(change) {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetRemovedNodes returns all nodes that were removed in the given changes.
func GetRemovedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if isRemovedChangeWithBefore(change) {
			nodes = append(nodes, change.Before)
		}
	}
	return nodes
}

// isNodeAdded checks if a node was added
func isNodeAdded(before, after *Node) bool {
	return before == nil && after != nil
}

// isNodeRemoved checks if a node was removed
func isNodeRemoved(before, after *Node) bool {
	return before != nil && after == nil
}

// isNoChange checks if there are no changes
func isNoChange(before, after *Node) bool {
	return before == nil && after == nil
}

// appendAddedChange appends an added change to the changes slice
func appendAddedChange(changes []Change, after *Node) []Change {
	return append(changes, Change{
		Before: nil,
		After:  after,
		Type:   ChangeAdded,
	})
}

// appendRemovedChange appends a removed change to the changes slice
func appendRemovedChange(changes []Change, before *Node) []Change {
	return append(changes, Change{
		Before: before,
		After:  nil,
		Type:   ChangeRemoved,
	})
}

// appendModifiedChange appends a modified change to the changes slice
func appendModifiedChange(changes []Change, before, after *Node) []Change {
	return append(changes, Change{
		Before: before,
		After:  after,
		Type:   ChangeModified,
	})
}

// appendChildrenChanges appends children changes to the changes slice
func appendChildrenChanges(changes []Change, before, after *Node) []Change {
	beforeChildren := before.Children
	afterChildren := after.Children
	childrenChanges := detectChildrenChanges(beforeChildren, afterChildren)
	return append(changes, childrenChanges...)
}

// isNodeModified checks if a node has been modified
func isNodeModified(before, after *Node) bool {
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
func hasDifferentType(before, after *Node) bool {
	return before.Type != after.Type
}

// hasDifferentToken checks if nodes have different tokens
func hasDifferentToken(before, after *Node) bool {
	return before.Token != after.Token
}

// hasSignificantPositionChange checks if position change is significant
func hasSignificantPositionChange(before, after *Node) bool {
	beforePos := before.Pos
	afterPos := after.Pos

	if isOnePositionNil(beforePos, afterPos) {
		return beforePos != afterPos
	}

	return hasSignificantLineOrColumnChange(beforePos, afterPos)
}

// isOnePositionNil checks if one of the positions is nil
func isOnePositionNil(beforePos, afterPos *Positions) bool {
	return beforePos == nil || afterPos == nil
}

// hasSignificantLineOrColumnChange checks if line or column change is significant
func hasSignificantLineOrColumnChange(beforePos, afterPos *Positions) bool {
	lineDiff := abs(beforePos.StartLine - afterPos.StartLine)
	columnDiff := abs(beforePos.StartCol - afterPos.StartCol)
	return lineDiff > 1 || columnDiff > 10
}

// hasDifferentStringRepresentation checks if nodes have different string representations
func hasDifferentStringRepresentation(before, after *Node) bool {
	return before.String() != after.String()
}

// detectChildrenChanges detects changes in child nodes
func detectChildrenChanges(beforeChildren, afterChildren []*Node) []Change {
	var changes []Change

	beforeMap := buildNodeMap(beforeChildren)
	afterMap := buildNodeMap(afterChildren)

	changes = appendModifiedChildren(changes, beforeMap, afterMap)
	changes = appendRemovedChildren(changes, beforeMap, afterMap)
	changes = appendAddedChildren(changes, beforeMap, afterMap)

	return changes
}

// buildNodeMap builds a map of nodes by key
func buildNodeMap(children []*Node) map[string]*Node {
	nodeMap := make(map[string]*Node)
	for _, child := range children {
		key := getNodeKey(child)
		nodeMap[key] = child
	}
	return nodeMap
}

// appendModifiedChildren appends modified children changes
func appendModifiedChildren(changes []Change, beforeMap, afterMap map[string]*Node) []Change {
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
func appendRemovedChildren(changes []Change, beforeMap, afterMap map[string]*Node) []Change {
	for key, beforeChild := range beforeMap {
		if isNodeNotInAfterMap(key, afterMap) {
			changes = appendRemovedChange(changes, beforeChild)
		}
	}
	return changes
}

// appendAddedChildren appends added children changes
func appendAddedChildren(changes []Change, beforeMap, afterMap map[string]*Node) []Change {
	for key, afterChild := range afterMap {
		if isNodeNotInBeforeMap(key, beforeMap) {
			changes = appendAddedChange(changes, afterChild)
		}
	}
	return changes
}

// isNodeNotInAfterMap checks if a node key is not in the after map
func isNodeNotInAfterMap(key string, afterMap map[string]*Node) bool {
	_, exists := afterMap[key]
	return !exists
}

// isNodeNotInBeforeMap checks if a node key is not in the before map
func isNodeNotInBeforeMap(key string, beforeMap map[string]*Node) bool {
	_, exists := beforeMap[key]
	return !exists
}

// getNodeKey creates a unique key for a node
func getNodeKey(node *Node) string {
	return node.Type + ":" + node.Token
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
