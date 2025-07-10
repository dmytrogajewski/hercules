package uast

// DetectChanges detects structural changes between two UAST nodes
func DetectChanges(before, after *Node) []Change {
	var changes []Change

	if before == nil && after != nil {
		// Node was added
		changes = append(changes, Change{
			Before: nil,
			After:  after,
			Type:   ChangeAdded,
		})
		return changes
	}

	if before != nil && after == nil {
		// Node was removed
		changes = append(changes, Change{
			Before: before,
			After:  nil,
			Type:   ChangeRemoved,
		})
		return changes
	}

	if before == nil && after == nil {
		// No changes
		return changes
	}

	// Check if node was modified
	if isNodeModified(before, after) {
		changes = append(changes, Change{
			Before: before,
			After:  after,
			Type:   ChangeModified,
		})
	}

	// Detect changes in children
	beforeChildren := before.Children
	afterChildren := after.Children

	changes = append(changes, detectChildrenChanges(beforeChildren, afterChildren)...)

	return changes
}

// isNodeModified checks if a node has been modified
func isNodeModified(before, after *Node) bool {
	if before.Type != after.Type {
		return true
	}

	if before.Token != after.Token {
		return true
	}

	// Compare positions (allowing for minor position changes)
	beforePos := before.Pos
	afterPos := after.Pos

	if beforePos == nil || afterPos == nil {
		if beforePos != afterPos {
			return true
		}
	} else {
		if abs(beforePos.StartLine-afterPos.StartLine) > 1 || abs(beforePos.StartCol-afterPos.StartCol) > 10 {
			return true
		}
	}

	// For Go functions, check if the function body has changed
	if before.Type == "go:function" && after.Type == "go:function" {
		beforeStr := before.String()
		afterStr := after.String()
		if beforeStr != afterStr {
			return true
		}
	}

	// For Java classes, check if the class content has changed
	if before.Type == "java:class" && after.Type == "java:class" {
		beforeStr := before.String()
		afterStr := after.String()
		if beforeStr != afterStr {
			return true
		}
	}

	// For Java methods, check if the method body has changed
	if before.Type == "java:method" && after.Type == "java:method" {
		beforeStr := before.String()
		afterStr := after.String()
		if beforeStr != afterStr {
			return true
		}
	}

	// For Java constructors, check if the constructor body has changed
	if before.Type == "java:constructor" && after.Type == "java:constructor" {
		beforeStr := before.String()
		afterStr := after.String()
		if beforeStr != afterStr {
			return true
		}
	}

	return false
}

// detectChildrenChanges detects changes in child nodes
func detectChildrenChanges(beforeChildren, afterChildren []*Node) []Change {
	var changes []Change

	beforeMap := make(map[string]*Node)
	afterMap := make(map[string]*Node)

	for _, child := range beforeChildren {
		key := getNodeKey(child)
		beforeMap[key] = child
	}

	for _, child := range afterChildren {
		key := getNodeKey(child)
		afterMap[key] = child
	}

	for key, beforeChild := range beforeMap {
		if _, exists := afterMap[key]; !exists {
			changes = append(changes, Change{
				Before: beforeChild,
				After:  nil,
				Type:   ChangeRemoved,
			})
		}
	}

	for key, afterChild := range afterMap {
		if _, exists := beforeMap[key]; !exists {
			changes = append(changes, Change{
				Before: nil,
				After:  afterChild,
				Type:   ChangeAdded,
			})
		}
	}

	for key, beforeChild := range beforeMap {
		if afterChild, exists := afterMap[key]; exists {
			if isNodeModified(beforeChild, afterChild) {
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

// FilterChangesByType filters changes by their type
func FilterChangesByType(changes []Change, changeType ChangeType) []Change {
	var filtered []Change
	for _, change := range changes {
		if change.Type == changeType {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// FilterChangesByNodeType filters changes by the type of nodes involved
func FilterChangesByNodeType(changes []Change, nodeType string) []Change {
	var filtered []Change
	for _, change := range changes {
		if (change.Before != nil && change.Before.Type == nodeType) ||
			(change.After != nil && change.After.Type == nodeType) {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// CountChangesByType counts changes by their type
func CountChangesByType(changes []Change) map[ChangeType]int {
	counts := make(map[ChangeType]int)
	for _, change := range changes {
		counts[change.Type]++
	}
	return counts
}

// GetModifiedNodes returns all nodes that were modified
func GetModifiedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if change.Type == ChangeModified && change.After != nil {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetAddedNodes returns all nodes that were added
func GetAddedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if change.Type == ChangeAdded && change.After != nil {
			nodes = append(nodes, change.After)
		}
	}
	return nodes
}

// GetRemovedNodes returns all nodes that were removed
func GetRemovedNodes(changes []Change) []*Node {
	var nodes []*Node
	for _, change := range changes {
		if change.Type == ChangeRemoved && change.Before != nil {
			nodes = append(nodes, change.Before)
		}
	}
	return nodes
}
