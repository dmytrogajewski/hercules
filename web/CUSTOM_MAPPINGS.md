# Custom UAST Mappings in Web Interface

The UAST Mapping Development Service now supports interactive development of custom UAST mappings through the web interface.

## 🚀 Getting Started

1. Start the development environment:
   ```bash
   make uast-dev
   ```

2. Open your browser to: http://localhost:3000

3. You'll see the UAST Mapping Development Service with a new "Custom Mappings" toggle button.

## 🔧 Using Custom Mappings

### 1. Enable Custom Mappings Panel

Click the "Custom Mappings" button in the top toolbar to expand the custom mappings panel.

### 2. Add Custom Mappings

- Click "Add Mapping" to create a new custom UAST mapping
- Fill in the mapping details:
  - **Name**: A unique identifier for your mapping
  - **Extensions**: File extensions this mapping applies to (e.g., `.json`, `.config`)
  - **UAST DSL**: The mapping rules in UAST DSL format

### 3. Load Example Mappings

The interface provides example mappings you can load:
- **custom_json**: Overrides the built-in JSON parser with custom node types
- **simple_config**: Adds support for `.config` and `.cfg` files

### 4. Test Your Mappings

1. Select a language that matches your mapping's extensions
2. Enter code in the input area
3. The parser will automatically use your custom mapping instead of the built-in one
4. Check the UAST output to see your custom node types

## 📝 UAST DSL Format

Custom mappings use the UAST DSL syntax:

```dsl
[language "base_language", extensions: ".ext1", ".ext2"]

node_type <- (pattern) => uast(
    type: "CustomType",
    token: "self",
    children: "child1", "child2"
)
```

### Example: Custom JSON Parser

```dsl
[language "json", extensions: ".json"]

_value <- (_value) => uast(
    type: "CustomValue"
)

document <- (document) => uast(
    type: "CustomDocument"
)

object <- (object) => uast(
    token: "self",
    type: "CustomObject"
)
```

## 🔄 Interactive Development Workflow

1. **Create**: Add a new custom mapping
2. **Edit**: Modify the DSL rules in real-time
3. **Test**: Enter code and see immediate results
4. **Iterate**: Refine your mapping based on the output
5. **Query**: Use UAST queries to test specific patterns

## 🎯 Use Cases

### Override Built-in Parsers
- Customize how existing languages are parsed
- Add domain-specific node types
- Implement experimental parsing strategies

### Add New Language Support
- Create parsers for custom file formats
- Add support for configuration files
- Implement DSLs for your projects

### Testing and Validation
- Test UAST mapping rules before committing
- Validate parsing behavior with real code
- Debug mapping issues interactively

## 🔍 Verifying Custom Mappings

When custom mappings are active, you'll see:
- Custom node types in the UAST output (e.g., `"CustomDocument"`, `"CustomObject"`)
- Different structure compared to built-in parsers
- Your mapping rules taking precedence over built-in ones

## 🛠️ API Integration

The web interface automatically sends custom mappings to the backend API:

```json
{
  "code": "{\"test\": \"value\"}",
  "language": "json",
  "uastmaps": {
    "custom_json": {
      "extensions": [".json"],
      "uast": "[language \"json\", extensions: \".json\"]..."
    }
  }
}
```

## 🚨 Troubleshooting

### Mapping Not Applied
- Check that file extensions match your mapping
- Verify DSL syntax is correct
- Ensure the base language is supported by tree-sitter

### Parse Errors
- Review the DSL syntax in the mapping
- Check for typos in node type names
- Verify pattern matching rules

### Performance Issues
- Keep mappings focused and efficient
- Avoid overly complex pattern matching
- Test with smaller code samples first

## 📚 Next Steps

- Explore the UAST DSL documentation for advanced features
- Create mappings for your specific use cases
- Share successful mappings with the community
- Integrate custom mappings into your development workflow 