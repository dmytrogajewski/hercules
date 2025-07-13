package uast

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/dmytrogajewski/hercules/pkg/uast/internal/node"
)

// TreeSitterProvider implements the UAST provider interface using Tree-sitter.
type TreeSitterProvider struct {
	language        *sitter.Language
	langName        string
	mapping         map[string]Mapping // kind -> Mapping
	IncludeUnmapped bool
}

// Parse parses the given file content and returns the root UAST node.
// Returns an error if parsing fails.
func (p *TreeSitterProvider) Parse(filename string, content []byte) (*node.Node, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("tree-sitter: failed to parse: %w", err)
	}
	root := tree.RootNode()
	if root.IsNull() {
		return nil, errors.New("tree-sitter: no root node")
	}

	tsNode := p.createTreeSitterNode(root, tree, content)
	canonical := tsNode.ToCanonicalNode()
	if canonical == nil {
		return nil, nil
	}
	return canonical, nil
}

// Language returns the language name for this provider.
func (p *TreeSitterProvider) Language() string {
	return p.langName
}

// TreeSitterNode wraps a Tree-sitter node for conversion to UAST.
type TreeSitterNode struct {
	Root            sitter.Node
	Tree            *sitter.Tree
	Language        string
	Source          []byte
	Mapping         map[string]Mapping // kind -> Mapping
	IncludeUnmapped bool
	ParentContext   string // Track parent context for conditional filtering
}

// ToCanonicalNode converts the TreeSitterNode to a canonical UAST Node.
func (tn *TreeSitterNode) ToCanonicalNode() *node.Node {
	kind := tn.Root.Type()
	mapping, hasMapping := tn.Mapping[kind]

	children := tn.processChildren(mapping)
	if tn.shouldSkipNode(mapping) {
		return nil
	}

	if tn.shouldSkipEmptyFile(kind, mapping, children) {
		return nil
	}

	props := map[string]string{}

	var roles []node.Role

	if hasMapping {
		return tn.createMappedNode(mapping, children, props, roles)
	}

	return tn.createUnmappedNode(kind, props, roles)
}

// createTreeSitterNode creates a new TreeSitterNode with the given parameters
func (p *TreeSitterProvider) createTreeSitterNode(root sitter.Node, tree *sitter.Tree, content []byte) *TreeSitterNode {
	return &TreeSitterNode{
		Root:            root,
		Tree:            tree,
		Language:        p.langName,
		Source:          content,
		Mapping:         p.mapping,
		IncludeUnmapped: p.IncludeUnmapped,
	}
}

// processChildren processes all children of the node, applying conditional filtering
func (tn *TreeSitterNode) processChildren(mapping Mapping) []*node.Node {
	childCount := tn.Root.NamedChildCount()
	children := make([]*node.Node, 0, childCount)

	for i := range childCount {
		child := tn.Root.NamedChild(i)
		childNode := tn.createChildNode(child, mapping)

		if tn.shouldExcludeChild(childNode, mapping) {
			continue
		}

		c := childNode.ToCanonicalNode()
		if c != nil && tn.shouldIncludeChild(child) {
			children = append(children, c)
		}
	}

	return children
}

// createChildNode creates a child TreeSitterNode with proper parent context
func (tn *TreeSitterNode) createChildNode(child sitter.Node, mapping Mapping) *TreeSitterNode {
	parentContext := mapping.Type
	if parentContext == "" {
		parentContext = tn.Root.Type()
	}
	return &TreeSitterNode{
		Root:            child,
		Tree:            tn.Tree,
		Language:        tn.Language,
		Source:          tn.Source,
		Mapping:         tn.Mapping,
		IncludeUnmapped: tn.IncludeUnmapped,
		ParentContext:   parentContext,
	}
}

// determineParentContext determines the parent context for child nodes
func (tn *TreeSitterNode) determineParentContext(mapping Mapping) string {
	kind := tn.Root.Type()
	if mapping.Type != "" {
		return mapping.Type
	}
	return kind
}

// shouldExcludeChild checks if a child should be excluded based on conditional filters
func (tn *TreeSitterNode) shouldExcludeChild(childNode *TreeSitterNode, parentMapping Mapping) bool {
	childKind := childNode.Root.Type()
	includeOnlyPresent := false
	for _, childMapping := range parentMapping.Children {
		if childMapping.Type != childKind {
			continue
		}
		if childMapping.IncludeOnly != nil {
			includeOnlyPresent = true
			if matchesCondition(childNode, *childMapping.IncludeOnly) {
				return false
			}
		}
	}
	if includeOnlyPresent {
		return true
	}
	for _, childMapping := range parentMapping.Children {
		if childMapping.Type != childKind {
			continue
		}
		if childMapping.ExcludeIf != nil && matchesCondition(childNode, *childMapping.ExcludeIf) {
			return true
		}
	}
	return false
}

// shouldSkipNode checks if the current node should be skipped
func (tn *TreeSitterNode) shouldSkipNode(mapping Mapping) bool {
	return mapping.Skip
}

// shouldSkipEmptyFile checks if an empty file should be skipped
func (tn *TreeSitterNode) shouldSkipEmptyFile(kind string, mapping Mapping, children []*node.Node) bool {
	return kind == "source_file" && mapping.SkipIfEmpty && len(children) == 0 && len(tn.Source) == 0
}

// determineNodeType determines the type string for the node
func (tn *TreeSitterNode) determineNodeType(kind string) string {
	if kind == "source_file" {
		return tn.Language + ":file"
	}
	return tn.Language + ":" + kind
}

// shouldIncludeChild checks if a child should be included in the result
func (tn *TreeSitterNode) shouldIncludeChild(child sitter.Node) bool {
	childMapping, hasChildMapping := tn.Mapping[child.Type()]
	return !hasChildMapping || !childMapping.Skip
}

// createMappedNode creates a UAST node from a mapped Tree-sitter node
func (tn *TreeSitterNode) createMappedNode(mapping Mapping, children []*node.Node, props map[string]string, roles []node.Role) *node.Node {
	tn.extractRoles(mapping, &roles)
	tn.extractName(mapping, props)
	tn.extractProperties(mapping, props)

	n := node.New(0, mapping.Type, tn.Token(), roles, tn.Positions(), props)
	n.Children = children

	tn.extractToken(mapping, n)
	return n
}

// extractRoles extracts roles from the mapping
func (tn *TreeSitterNode) extractRoles(mapping Mapping, roles *[]node.Role) {
	for _, r := range mapping.Roles {
		*roles = append(*roles, node.Role(r))
	}
}

// extractName extracts name from the node if specified in mapping
func (tn *TreeSitterNode) extractName(mapping Mapping, props map[string]string) {
	if mapping.Name != nil {
		name := extractNameFromNode(tn, mapping.Name.Source)
		if name != "" {
			props["name"] = name
		}
	}
}

// extractProperties extracts properties from the node
func (tn *TreeSitterNode) extractProperties(mapping Mapping, props map[string]string) {
	for propKey, propVal := range mapping.Props {
		if propStr, ok := propVal.(string); ok {
			value := tn.extractPropertyValue(propStr)
			if value != "" {
				props[propKey] = value
			}
		}
	}
}

// extractPropertyValue extracts a property value from the node
func (tn *TreeSitterNode) extractPropertyValue(propStr string) string {
	if tn.isDescendantProperty(propStr) {
		return tn.extractDescendantProperty(propStr)
	}
	return tn.extractDirectChildProperty(propStr)
}

// isDescendantProperty checks if the property is a descendant property
func (tn *TreeSitterNode) isDescendantProperty(propStr string) bool {
	_, ok := strings.CutPrefix(propStr, "descendant:")
	return ok
}

// extractDescendantProperty extracts a descendant property
func (tn *TreeSitterNode) extractDescendantProperty(propStr string) string {
	after, _ := strings.CutPrefix(propStr, "descendant:")
	return extractTokenFromDescendant(tn, after)
}

// extractDirectChildProperty extracts a direct child property
func (tn *TreeSitterNode) extractDirectChildProperty(propStr string) string {
	for i := uint32(0); i < tn.Root.NamedChildCount(); i++ {
		c := tn.Root.NamedChild(i)
		childKind := c.Type()
		if childKind == propStr {
			return tn.extractChildText(c)
		}
	}
	return ""
}

// extractChildText extracts text from a child node
func (tn *TreeSitterNode) extractChildText(child sitter.Node) string {
	start := child.StartByte()
	end := child.EndByte()
	if int(end) <= len(tn.Source) {
		return string(tn.Source[start:end])
	}
	return ""
}

// extractToken extracts token from the node if specified in mapping
func (tn *TreeSitterNode) extractToken(mapping Mapping, node *node.Node) {
	if mapping.Token != "" {
		token := extractTokenFromNode(tn, mapping.Token)
		if token != "" {
			node.Token = token
		}
	}
}

// createUnmappedNode creates a UAST node for unmapped Tree-sitter nodes
func (tn *TreeSitterNode) createUnmappedNode(kind string, props map[string]string, roles []node.Role) *node.Node {
	mappedChildren := tn.processUnmappedChildren()

	if tn.IncludeUnmapped {
		return tn.createIncludeUnmappedNode(kind, mappedChildren, props, roles)
	}

	return tn.createSyntheticNode(mappedChildren)
}

// processUnmappedChildren processes children for unmapped nodes
func (tn *TreeSitterNode) processUnmappedChildren() []*node.Node {
	var mappedChildren []*node.Node
	for i := uint32(0); i < tn.Root.NamedChildCount(); i++ {
		child := tn.Root.NamedChild(i)
		childNode := tn.createUnmappedChildNode(child)
		c := childNode.ToCanonicalNode()
		if c != nil {
			mappedChildren = append(mappedChildren, c)
		}
	}
	return mappedChildren
}

// createUnmappedChildNode creates a child node for unmapped nodes
func (tn *TreeSitterNode) createUnmappedChildNode(child sitter.Node) *TreeSitterNode {
	return &TreeSitterNode{
		Root:            child,
		Tree:            tn.Tree,
		Language:        tn.Language,
		Source:          tn.Source,
		Mapping:         tn.Mapping,
		IncludeUnmapped: tn.IncludeUnmapped,
		ParentContext:   tn.ParentContext,
	}
}

// createIncludeUnmappedNode creates a node when IncludeUnmapped is true
func (tn *TreeSitterNode) createIncludeUnmappedNode(kind string, mappedChildren []*node.Node, props map[string]string, roles []node.Role) *node.Node {
	node := node.New(0, tn.Language+":"+kind, tn.Token(), roles, tn.Positions(), props)
	node.Children = mappedChildren
	return node
}

// createSyntheticNode creates a synthetic node for multiple children
func (tn *TreeSitterNode) createSyntheticNode(mappedChildren []*node.Node) *node.Node {
	if len(mappedChildren) == 1 {
		return mappedChildren[0]
	}
	if len(mappedChildren) > 1 {
		synth := node.New(0, "Synthetic", "", nil, nil, nil)
		synth.Children = mappedChildren
		return synth
	}
	return nil
}

// matchesCondition checks if a node matches a conditional filter
func matchesCondition(node *TreeSitterNode, filter ConditionalFilter) bool {
	if !matchesTypeCondition(node, filter) {
		return false
	}

	if !matchesParentContextCondition(node, filter) {
		return false
	}

	if !matchesFieldCondition(node, filter) {
		return false
	}

	if !matchesPropsCondition(node, filter) {
		return false
	}

	return true
}

// matchesTypeCondition checks if the node matches the type condition
func matchesTypeCondition(node *TreeSitterNode, filter ConditionalFilter) bool {
	if filter.Type == "" {
		return true
	}
	return node.Root.Type() == filter.Type
}

// matchesParentContextCondition checks if the node matches the parent context condition
func matchesParentContextCondition(node *TreeSitterNode, filter ConditionalFilter) bool {
	if filter.ParentContext == "" {
		return true
	}
	return node.ParentContext == filter.ParentContext
}

// matchesFieldCondition checks if the node matches the field condition
func matchesFieldCondition(node *TreeSitterNode, filter ConditionalFilter) bool {
	if filter.HasField == "" {
		return true
	}
	return node.hasField(filter.HasField)
}

// matchesPropsCondition checks if the node matches the props condition
func matchesPropsCondition(node *TreeSitterNode, filter ConditionalFilter) bool {
	if len(filter.Props) == 0 {
		return true
	}

	for key, value := range filter.Props {
		if !node.hasProperty(key, value) {
			return false
		}
	}
	return true
}

// hasField checks if the node has a specific field
func (tn *TreeSitterNode) hasField(fieldName string) bool {
	if tn.hasFieldByName(fieldName) {
		return true
	}
	return tn.hasFieldByType(fieldName)
}

// hasFieldByName checks if the node has a field by name using Tree-sitter's field API
func (tn *TreeSitterNode) hasFieldByName(fieldName string) bool {
	fieldNode := tn.Root.ChildByFieldName(fieldName)
	return !fieldNode.IsNull()
}

// hasFieldByType checks if any named child has the field name as its type
func (tn *TreeSitterNode) hasFieldByType(fieldName string) bool {
	for i := uint32(0); i < tn.Root.NamedChildCount(); i++ {
		child := tn.Root.NamedChild(i)
		if child.Type() == fieldName {
			return true
		}
	}
	return false
}

// hasProperty checks if the node has a specific property with a value
func (tn *TreeSitterNode) hasProperty(key, value string) bool {
	for i := uint32(0); i < tn.Root.NamedChildCount(); i++ {
		child := tn.Root.NamedChild(i)
		if tn.matchesPropertyChild(child, key, value) {
			return true
		}
	}
	return false
}

// matchesPropertyChild checks if a child matches the property criteria
func (tn *TreeSitterNode) matchesPropertyChild(child sitter.Node, key, value string) bool {
	if child.Type() != key {
		return false
	}

	if value == "" {
		return true
	}

	childText := tn.extractChildText(child)
	return childText == value
}

// extractNameFromNode extracts a name from a node using the specified source
func extractNameFromNode(node *TreeSitterNode, source string) string {
	switch source {
	case "fields.name":
		return extractNameFromField(node, "name")
	case "props.name":
		return extractNameFromProps(node)
	case "text":
		return extractNameFromText(node)
	default:
		return ""
	}
}

// extractNameFromField extracts a name from a specific field using Tree-sitter's field API
func extractNameFromField(node *TreeSitterNode, fieldName string) string {
	fieldNode := node.Root.ChildByFieldName(fieldName)
	if !fieldNode.IsNull() {
		return node.extractNodeText(fieldNode)
	}
	return node.extractNameFromChildType(fieldName)
}

// extractNameFromChildType extracts name from a child with the field name as its type
func (tn *TreeSitterNode) extractNameFromChildType(fieldName string) string {
	for i := uint32(0); i < tn.Root.NamedChildCount(); i++ {
		child := tn.Root.NamedChild(i)
		if child.Type() == fieldName {
			return tn.extractNodeText(child)
		}
	}
	return ""
}

// extractNodeText extracts text from a Tree-sitter node
func (tn *TreeSitterNode) extractNodeText(node sitter.Node) string {
	start := node.StartByte()
	end := node.EndByte()
	if int(end) <= len(tn.Source) {
		return string(tn.Source[start:end])
	}
	return ""
}

// extractNameFromText extracts name from node text
func extractNameFromText(node *TreeSitterNode) string {
	if node.Root.ChildCount() == 0 {
		return node.extractNodeText(node.Root)
	}
	return ""
}

// extractNameFromProps extracts name from node properties (legacy)
func extractNameFromProps(node *TreeSitterNode) string {
	return extractNameFromText(node)
}

// extractTokenFromNode extracts a token from a node using the specified source
func extractTokenFromNode(node *TreeSitterNode, source string) string {
	switch source {
	case "text", "self":
		return node.extractSelfToken()
	default:
		return node.extractChildOrDescendantToken(source)
	}
}

// extractSelfToken extracts token from the node's own text
func (tn *TreeSitterNode) extractSelfToken() string {
	if tn.Root.ChildCount() == 0 {
		return tn.extractNodeText(tn.Root)
	}
	return ""
}

// extractChildOrDescendantToken extracts token from child or descendant
func (tn *TreeSitterNode) extractChildOrDescendantToken(source string) string {
	if tn.isChildToken(source) {
		return tn.extractChildToken(source)
	}
	if tn.isDescendantToken(source) {
		return tn.extractDescendantToken(source)
	}
	return ""
}

// isChildToken checks if the source is a child token
func (tn *TreeSitterNode) isChildToken(source string) bool {
	_, ok := strings.CutPrefix(source, "child:")
	return ok
}

// extractChildToken extracts token from a child field
func (tn *TreeSitterNode) extractChildToken(source string) string {
	after, _ := strings.CutPrefix(source, "child:")
	childNode := tn.Root.ChildByFieldName(after)
	if !childNode.IsNull() {
		return tn.extractNodeText(childNode)
	}
	return ""
}

// isDescendantToken checks if the source is a descendant token
func (tn *TreeSitterNode) isDescendantToken(source string) bool {
	_, ok := strings.CutPrefix(source, "descendant:")
	return ok
}

// extractDescendantToken extracts token from a descendant
func (tn *TreeSitterNode) extractDescendantToken(source string) string {
	after, _ := strings.CutPrefix(source, "descendant:")
	return extractTokenFromDescendant(tn, after)
}

// extractTokenFromDescendant finds the first descendant of the specified type and extracts its token
func extractTokenFromDescendant(node *TreeSitterNode, nodeType string) string {
	return findDescendantToken(node.Root, node.Source, nodeType)
}

// findDescendantToken recursively searches for a descendant of the specified type
func findDescendantToken(root sitter.Node, source []byte, nodeType string) string {
	if root.Type() == nodeType {
		return extractNodeText(root, source)
	}

	for i := uint32(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		if result := findDescendantToken(child, source, nodeType); result != "" {
			return result
		}
	}

	return ""
}

// extractNodeText extracts text from a Tree-sitter node with source
func extractNodeText(root sitter.Node, source []byte) string {
	start := root.StartByte()
	end := root.EndByte()
	if int(end) <= len(source) {
		return string(source[start:end])
	}
	return ""
}

// Token returns the string token for this node, if any.
func (tn *TreeSitterNode) Token() string {
	if tn.Root.ChildCount() == 0 {
		return tn.extractNodeText(tn.Root)
	}
	return ""
}

// Positions returns the source code positions for this node.
func (tn *TreeSitterNode) Positions() *node.Positions {
	return &node.Positions{
		StartLine:   int(tn.Root.StartPoint().Row),
		StartCol:    int(tn.Root.StartPoint().Column),
		StartOffset: int(tn.Root.StartByte()),
		EndLine:     int(tn.Root.EndPoint().Row),
		EndCol:      int(tn.Root.EndPoint().Column),
		EndOffset:   int(tn.Root.EndByte()),
	}
}

// hasMappedType checks if a type is mapped
func (tn *TreeSitterNode) hasMappedType(typ string) bool {
	_, ok := tn.Mapping[typ]
	return ok
}
