package uast

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// FormatValidationErrors formats validation errors with better readability
func FormatValidationErrors(errors []gojsonschema.ResultError, contextName string) string {
	if len(errors) == 0 {
		return ""
	}

	var errorMsgs []string
	errorMsgs = append(errorMsgs, fmt.Sprintf("❌ Validation failed for %s", contextName))
	errorMsgs = append(errorMsgs, "")
	errorMsgs = append(errorMsgs, "📋 Validation Errors:")
	errorMsgs = append(errorMsgs, "")

	// Group errors by type for better organization
	roleErrors := make([]gojsonschema.ResultError, 0)
	typeErrors := make([]gojsonschema.ResultError, 0)
	otherErrors := make([]gojsonschema.ResultError, 0)

	for _, verr := range errors {
		if strings.Contains(verr.Description(), "must be one of the following") {
			roleErrors = append(roleErrors, verr)
		} else if strings.Contains(verr.Description(), "type") {
			typeErrors = append(typeErrors, verr)
		} else {
			otherErrors = append(otherErrors, verr)
		}
	}

	// Report role errors with pretty formatting
	if len(roleErrors) > 0 {
		errorMsgs = append(errorMsgs, "🔑 Role Validation Errors:")
		for _, verr := range roleErrors {
			actualValue := getActualValueFromError(verr)
			expectedRoles := extractExpectedRoles(verr.Description())
			suggestion := getRoleSuggestion(verr.Description(), actualValue)

			errorMsgs = append(errorMsgs, fmt.Sprintf("   • %s", verr.Field()))
			errorMsgs = append(errorMsgs, fmt.Sprintf("     Expected: %s", expectedRoles))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			if suggestion != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     💡 %s", suggestion))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Report type errors
	if len(typeErrors) > 0 {
		errorMsgs = append(errorMsgs, "🏷️  Type Validation Errors:")
		for _, verr := range typeErrors {
			actualValue := getActualValueFromError(verr)
			expectedTypes := extractExpectedTypes(verr.Description())

			errorMsgs = append(errorMsgs, fmt.Sprintf("   • %s", verr.Field()))
			errorMsgs = append(errorMsgs, fmt.Sprintf("     Expected: %s", expectedTypes))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Report other errors
	if len(otherErrors) > 0 {
		errorMsgs = append(errorMsgs, "⚠️  Other Validation Errors:")
		for _, verr := range otherErrors {
			actualValue := getActualValueFromError(verr)

			errorMsgs = append(errorMsgs, fmt.Sprintf("   • %s", verr.Field()))
			errorMsgs = append(errorMsgs, fmt.Sprintf("     %s", verr.Description()))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Add helpful suggestions
	errorMsgs = append(errorMsgs, "💡 Suggestions:")
	errorMsgs = append(errorMsgs, "   • Update mappings to use canonical UAST types and roles")
	errorMsgs = append(errorMsgs, "   • Replace non-canonical types with valid UAST types")
	errorMsgs = append(errorMsgs, "   • Use roles from the UAST schema specification")
	errorMsgs = append(errorMsgs, "   • Consider using 'Synthetic' type for unmapped nodes")
	errorMsgs = append(errorMsgs, "")

	return strings.Join(errorMsgs, "\n")
}

// getActualValueFromError extracts the actual value from the error field path
func getActualValueFromError(verr gojsonschema.ResultError) string {
	// This is a simplified version - in practice, we'd need the actual data
	// For now, we'll extract what we can from the error message
	description := verr.Description()
	if strings.Contains(description, "Got:") {
		parts := strings.Split(description, "Got:")
		if len(parts) > 1 {
			gotPart := strings.TrimSpace(parts[1])
			// Extract the quoted value
			if strings.HasPrefix(gotPart, `"`) {
				endQuote := strings.Index(gotPart[1:], `"`)
				if endQuote != -1 {
					return gotPart[1 : endQuote+1]
				}
			}
		}
	}
	return ""
}

// extractExpectedRoles extracts the expected roles from the error description
func extractExpectedRoles(description string) string {
	if strings.Contains(description, "must be one of the following:") {
		parts := strings.Split(description, "must be one of the following:")
		if len(parts) > 1 {
			roles := strings.TrimSpace(parts[1])
			// Clean up the roles list and format it nicely
			roles = strings.ReplaceAll(roles, `"`, "")
			roles = strings.ReplaceAll(roles, `,`, ", ")
			return roles
		}
	}
	return description
}

// extractExpectedTypes extracts the expected types from the error description
func extractExpectedTypes(description string) string {
	if strings.Contains(description, "must be one of the following:") {
		parts := strings.Split(description, "must be one of the following:")
		if len(parts) > 1 {
			types := strings.TrimSpace(parts[1])
			// Clean up the types list and format it nicely
			types = strings.ReplaceAll(types, `"`, "")
			types = strings.ReplaceAll(types, `,`, ", ")
			return types
		}
	}
	return description
}

// getRoleSuggestion provides helpful suggestions for role errors
func getRoleSuggestion(description, actualValue string) string {
	if strings.Contains(description, "must be one of the following") {
		if actualValue == "" {
			return "Node is missing required roles. Add appropriate roles from the UAST schema."
		}

		// Provide specific suggestions based on common patterns
		switch actualValue {
		case "Package":
			return "Consider adding Declaration role: roles: \"Package\", \"Declaration\""
		case "Body":
			return "This appears to be a Block node. Ensure it has roles: \"Body\""
		case "Name":
			return "This appears to be an Identifier node. Ensure it has roles: \"Name\""
		default:
			return fmt.Sprintf("Replace %q with a valid role from the UAST schema", actualValue)
		}
	}
	return ""
}
