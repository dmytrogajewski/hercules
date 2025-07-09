package uastconvert

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
	uast_nodes "gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// nodeTypeMapping maps Tree-sitter node types to UAST types.
var nodeTypeMapping = map[string]string{
	// Java
	"method_declaration": "uast:FunctionGroup",
	"class_declaration":  "uast:Class",
	"identifier":         "uast:Identifier",
	// Go
	"function_declaration": "uast:FunctionGroup",
	"type_spec":            "uast:Class",
	// Add more as needed
}

// attributeExtractors maps node types to functions that extract key attributes.
var attributeExtractors = map[string]func(*sitter.Node, []byte) map[string]uast_nodes.Node{
	// Java method_declaration: extract Name from child identifier
	"method_declaration": func(node *sitter.Node, source []byte) map[string]uast_nodes.Node {
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				return map[string]uast_nodes.Node{"Name": uast_nodes.String(safeExtractContent(&child, source))}
			}
		}
		return nil
	},
	// Java class_declaration: extract Name from child identifier
	"class_declaration": func(node *sitter.Node, source []byte) map[string]uast_nodes.Node {
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				return map[string]uast_nodes.Node{"Name": uast_nodes.String(safeExtractContent(&child, source))}
			}
		}
		return nil
	},
	// Go function_declaration: extract Name from child identifier
	"function_declaration": func(node *sitter.Node, source []byte) map[string]uast_nodes.Node {
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child.Type() == "identifier" {
				return map[string]uast_nodes.Node{"Name": uast_nodes.String(safeExtractContent(&child, source))}
			}
		}
		return nil
	},
	// Add more as needed
}

// ConvertSitterNodeToUAST iteratively converts a *sitter.Node to a uast_nodes.Object using a stack.
func ConvertSitterNodeToUAST(node *sitter.Node, source []byte) uast_nodes.Object {
	if node == nil || *node == (sitter.Node{}) {
		return nil
	}

	type stackItem struct {
		tsNode   *sitter.Node
		uastObj  uast_nodes.Object
		parent   uast_nodes.Object
		childIdx int
	}

	rootObj := uast_nodes.Object{}
	stack := []stackItem{{tsNode: node, uastObj: rootObj, parent: nil, childIdx: -1}}
	var parentMap = map[*sitter.Node]uast_nodes.Object{node: rootObj}

	for len(stack) > 0 {
		item := &stack[len(stack)-1]
		if item.tsNode == nil || *item.tsNode == (sitter.Node{}) {
			stack = stack[:len(stack)-1]
			continue
		}
		if item.childIdx == -1 {
			// First visit: fill in type, attributes, and position
			tsType := item.tsNode.Type()
			uastType, hasMapping := nodeTypeMapping[tsType]
			if !hasMapping {
				uastType = tsType
			}
			item.uastObj["@type"] = uast_nodes.String(uastType)
			if extractor, ok := attributeExtractors[tsType]; ok {
				attrs := extractor(item.tsNode, source)
				for k, v := range attrs {
					item.uastObj[k] = v
				}
			}
			startPoint := item.tsNode.StartPoint()
			endPoint := item.tsNode.EndPoint()
			item.uastObj["@pos"] = uast_nodes.Object{
				"@start": uast_nodes.Object{
					"line":   uast_nodes.Uint(startPoint.Row),
					"column": uast_nodes.Uint(startPoint.Column),
				},
				"@end": uast_nodes.Object{
					"line":   uast_nodes.Uint(endPoint.Row),
					"column": uast_nodes.Uint(endPoint.Column),
				},
			}
			// TODO: Enable @token extraction when safe (see above)
			item.childIdx = 0
		}
		if item.childIdx < int(item.tsNode.ChildCount()) {
			child := item.tsNode.Child(uint32(item.childIdx))
			if child != (sitter.Node{}) {
				childObj := uast_nodes.Object{}
				if item.uastObj["Children"] == nil {
					item.uastObj["Children"] = uast_nodes.Array{}
				}
				childrenArr := item.uastObj["Children"].(uast_nodes.Array)
				childrenArr = append(childrenArr, childObj)
				item.uastObj["Children"] = childrenArr
				parentMap[&child] = childObj
				// Only push valid children
				if &child != nil && child != (sitter.Node{}) {
					stack = append(stack, stackItem{tsNode: &child, uastObj: childObj, parent: item.uastObj, childIdx: -1})
				}
			}
			item.childIdx++
		} else {
			stack = stack[:len(stack)-1]
		}
	}
	return rootObj
}

// HashNode computes a SHA256 hash of the UAST node for comparison and deduplication.
func HashNode(node uast_nodes.Node) ([]byte, error) {
	// Serialize to JSON for consistent hashing
	data, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize node for hashing: %w", err)
	}
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// SerializeNode converts a UAST node to a serializable format compatible with Protocol Buffers.
func SerializeNode(node uast_nodes.Node) (map[string]interface{}, error) {
	if node == nil {
		return nil, nil
	}

	// Convert uast_nodes types to standard Go types for serialization
	var convert func(interface{}) interface{}
	convert = func(v interface{}) interface{} {
		switch val := v.(type) {
		case uast_nodes.String:
			return string(val)
		case uast_nodes.Int:
			return int(val)
		case uast_nodes.Uint:
			return uint(val)
		case uast_nodes.Float:
			return float64(val)
		case uast_nodes.Bool:
			return bool(val)
		case uast_nodes.Array:
			result := make([]interface{}, len(val))
			for i, item := range val {
				result[i] = convert(item)
			}
			return result
		case uast_nodes.Object:
			result := make(map[string]interface{})
			for key, value := range val {
				result[key] = convert(value)
			}
			return result
		default:
			return v
		}
	}

	serialized := convert(node)
	if obj, ok := serialized.(map[string]interface{}); ok {
		return obj, nil
	}
	return nil, fmt.Errorf("failed to serialize node")
}

// StringifyNode creates a string representation of the node for debugging and comparison.
func StringifyNode(node uast_nodes.Node) string {
	if node == nil {
		return "nil"
	}

	var buffer bytes.Buffer
	var stringify func(uast_nodes.Node, int)
	stringify = func(n uast_nodes.Node, depth int) {
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}

		if obj, ok := n.(uast_nodes.Object); ok {
			// Sort keys for consistent output
			var keys []string
			for key := range obj {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				value := obj[key]
				if str, ok := value.(uast_nodes.String); ok {
					buffer.WriteString(fmt.Sprintf("%s%s: %s\n", indent, key, string(str)))
				} else if child, ok := value.(uast_nodes.Node); ok {
					buffer.WriteString(fmt.Sprintf("%s%s:\n", indent, key))
					stringify(child, depth+1)
				} else {
					buffer.WriteString(fmt.Sprintf("%s%s: %v\n", indent, key, value))
				}
			}
		} else if arr, ok := n.(uast_nodes.Array); ok {
			for i, item := range arr {
				buffer.WriteString(fmt.Sprintf("%s[%d]:\n", indent, i))
				stringify(item, depth+1)
			}
		}
	}

	stringify(node, 0)
	return buffer.String()
}

// safeExtractContent attempts to extract content from a Tree-sitter node safely.
// Returns an empty string if extraction is unsafe.
func safeExtractContent(node *sitter.Node, source []byte) string {
	if node == nil || *node == (sitter.Node{}) || len(source) == 0 {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if int(start) >= 0 && int(end) <= len(source) && int(start) < int(end) {
		return string(source[start:end])
	}
	return ""
}
