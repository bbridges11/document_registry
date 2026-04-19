package inprocess

import (
	"fmt"

	"github.com/bbridges_11/document-registry/internal/ports/outbound"
)

// validateSemantics performs deeper validation beyond struct tags
func validateSemantics(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	// Validate relationships reference actual components
	issues = append(issues, validateRelationships(def)...)

	// Validate input references in conditions
	issues = append(issues, validateInputReferences(def)...)

	// Validate component uniqueness
	issues = append(issues, validateComponentUniqueness(def)...)

	// Validate enum inputs have values
	issues = append(issues, validateEnumInputs(def)...)

	// Validate fileScope consistency
	issues = append(issues, validateFileScopeConsistency(def)...)

	return issues
}

// validateRelationships ensures relationship endpoints reference actual components
func validateRelationships(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	// Build component map
	componentMap := make(map[string]bool)
	for _, comp := range def.Components {
		key := comp.Type + ":" + comp.Name
		componentMap[key] = true
	}

	// Check each relationship
	for i, rel := range def.Relationships {
		fromKey := rel.From.Type + ":" + rel.From.Name
		if !componentMap[fromKey] {
			issues = append(issues, outbound.ValidationIssue{
				Field:    fmt.Sprintf("relationships[%d].from", i),
				Rule:     "reference_exists",
				Message:  fmt.Sprintf("component '%s' (type: %s, name: %s) not found in components", fromKey, rel.From.Type, rel.From.Name),
				Severity: "error",
			})
		}

		toKey := rel.To.Type + ":" + rel.To.Name
		if !componentMap[toKey] {
			issues = append(issues, outbound.ValidationIssue{
				Field:    fmt.Sprintf("relationships[%d].to", i),
				Rule:     "reference_exists",
				Message:  fmt.Sprintf("component '%s' (type: %s, name: %s) not found in components", toKey, rel.To.Type, rel.To.Name),
				Severity: "error",
			})
		}
	}

	return issues
}

// validateInputReferences ensures condition references point to defined inputs
func validateInputReferences(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	for i, comp := range def.Components {
		if comp.IncludeWhen != nil && comp.IncludeWhen.InputEquals != nil {
			for inputName := range comp.IncludeWhen.InputEquals {
				if _, exists := def.Inputs[inputName]; !exists {
					issues = append(issues, outbound.ValidationIssue{
						Field:    fmt.Sprintf("components[%d].includeWhen.input_equals.%s", i, inputName),
						Rule:     "input_exists",
						Message:  fmt.Sprintf("input '%s' not defined in inputs section", inputName),
						Severity: "error",
					})
				}
			}
		}
	}

	return issues
}

// validateComponentUniqueness ensures each component type:name combination is unique
func validateComponentUniqueness(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	seen := make(map[string]int)
	for i, comp := range def.Components {
		key := comp.Type + ":" + comp.Name
		if firstIndex, exists := seen[key]; exists {
			issues = append(issues, outbound.ValidationIssue{
				Field:    fmt.Sprintf("components[%d]", i),
				Rule:     "unique",
				Message:  fmt.Sprintf("duplicate component '%s' (first defined at components[%d])", key, firstIndex),
				Severity: "error",
			})
		} else {
			seen[key] = i
		}
	}

	return issues
}

// validateEnumInputs ensures enum type inputs have values defined
func validateEnumInputs(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	for name, input := range def.Inputs {
		if input.Type == "enum" {
			if len(input.Values) == 0 {
				issues = append(issues, outbound.ValidationIssue{
					Field:    fmt.Sprintf("inputs.%s", name),
					Rule:     "enum_values_required",
					Message:  "enum type inputs must define values",
					Severity: "error",
				})
			}

			// Validate default is in values if specified
			if input.Default != nil && len(input.Values) > 0 {
				defaultStr := fmt.Sprintf("%v", input.Default)
				found := false
				for _, val := range input.Values {
					if val == defaultStr {
						found = true
						break
					}
				}
				if !found {
					issues = append(issues, outbound.ValidationIssue{
						Field:    fmt.Sprintf("inputs.%s.default", name),
						Rule:     "enum_default_in_values",
						Message:  fmt.Sprintf("default value '%v' must be one of: %v", input.Default, input.Values),
						Severity: "error",
					})
				}
			}
		}
	}

	return issues
}

// validateFileScopeConsistency ensures fileScope and fileScopeByTopology are not both set
func validateFileScopeConsistency(def *DefinitionSchema) []outbound.ValidationIssue {
	var issues []outbound.ValidationIssue

	for i, comp := range def.Components {
		if comp.FileScope != "" && len(comp.FileScopeByTopology) > 0 {
			issues = append(issues, outbound.ValidationIssue{
				Field:    fmt.Sprintf("components[%d]", i),
				Rule:     "fileScope_exclusive",
				Message:  "component cannot have both fileScope and fileScopeByTopology defined",
				Severity: "error",
			})
		}
	}

	return issues
}
