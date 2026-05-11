# Architecture Definition System — Complete Examples and Generator Logic

## Final Model

This system defines a logical architecture in YAML and generates one or more output files based on runtime inputs.

Current design:

- `topology` is an enum input.
- `optional_components` is an array input.
- Optional components are selected by full component key: `type:name`.
- Components are uniquely identified by `type:name`.
- `resolve.when` controls conditional inclusion.
- `resolve.fileScope` controls fixed placement.
- `resolve.placement: promoted` controls topology-aware placement.
- Relationships are defined separately and become `connectsTo` in the generated output.

---

## Semantics

### Defaults

If a component has no `resolve` block:

```yaml
- type: ECSCluster
  name: apiCluster
```

It means:

- included by default
- placed in the `regional` file

If a relationship has no `resolve.when`, it is active by default as long as both endpoints are active.

### Optional component

```yaml
- type: DynamoDB
  name: userTable
  resolve:
    when:
      contains:
        input: optional_components
        value: DynamoDB:userTable
```

### Promoted placement

```yaml
resolve:
  placement: promoted
```

Means:

- `single_region` → `regional`
- `multi_region` → `global`

### Fixed global placement

```yaml
resolve:
  fileScope: global
```

Means the component is always placed in the global file whenever active.

---

## Condition DSL

Supported condition operators:

```yaml
input_equals:
contains:
all:
any:
not:
```

### Example: array contains

```yaml
when:
  contains:
    input: optional_components
    value: DynamoDB:userTable
```

---

## Complete Definition Example — Basic

```yaml
kind: Definition
version: v1

inputs:
  topology:
    type: enum
    values: [single_region, multi_region]
    default: single_region

  optional_components:
    type: array
    items:
      type: string
    default: []

components:
  - type: ECSCluster
    name: apiCluster

  - type: S3
    name: assetBucket

  - type: DynamoDB
    name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable
      placement: promoted

relationships:
  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: S3
      name: assetBucket

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: DynamoDB
      name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable
```

---

## Complete Definition Example — Multiple Optional Components of Same Type

```yaml
kind: Definition
version: v1

inputs:
  topology:
    type: enum
    values: [single_region, multi_region]
    default: single_region

  optional_components:
    type: array
    items:
      type: string
    default: []

components:
  - type: ECSCluster
    name: apiCluster

  - type: S3
    name: assetBucket

  - type: DynamoDB
    name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable
      placement: promoted

  - type: DynamoDB
    name: auditTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:auditTable
      placement: promoted

relationships:
  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: S3
      name: assetBucket

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: DynamoDB
      name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: DynamoDB
      name: auditTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:auditTable
```

---

## Complete Definition Example — Larger Architecture

```yaml
kind: Definition
version: v1

inputs:
  topology:
    type: enum
    values: [single_region, multi_region]
    default: single_region

  optional_components:
    type: array
    items:
      type: string
    default: []

components:
  - type: ECSCluster
    name: apiCluster

  - type: LoadBalancer
    name: publicALB

  - type: S3
    name: assetBucket

  - type: DynamoDB
    name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable
      placement: promoted

  - type: SQS
    name: eventQueue
    resolve:
      when:
        contains:
          input: optional_components
          value: SQS:eventQueue

  - type: Kinesis
    name: analyticsStream
    resolve:
      when:
        contains:
          input: optional_components
          value: Kinesis:analyticsStream

relationships:
  - from:
      type: LoadBalancer
      name: publicALB
    to:
      type: ECSCluster
      name: apiCluster

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: S3
      name: assetBucket

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: DynamoDB
      name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB:userTable

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: SQS
      name: eventQueue
    resolve:
      when:
        contains:
          input: optional_components
          value: SQS:eventQueue

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: Kinesis
      name: analyticsStream
    resolve:
      when:
        contains:
          input: optional_components
          value: Kinesis:analyticsStream
```

---

## Runtime Input Examples

### Single-region, no optional components

```yaml
topology: single_region
optional_components: []
```

### Single-region, include DynamoDB

```yaml
topology: single_region
optional_components:
  - DynamoDB:userTable
```

### Multi-region, include DynamoDB

```yaml
topology: multi_region
optional_components:
  - DynamoDB:userTable
```

### Multi-region, include multiple optional components

```yaml
topology: multi_region
optional_components:
  - DynamoDB:userTable
  - DynamoDB:auditTable
  - SQS:eventQueue
```

---

## Generated Output Examples

### Case A — Single-region, no optional components

```yaml
kind: File
version: v1
metadata:
  fileName: file1
  fileVersion: v1
spec:
  components:
    - type: ECSCluster
      name: apiCluster
      connectsTo:
        - type: S3
          name: assetBucket
    - type: S3
      name: assetBucket
```

### Case B — Single-region with DynamoDB

```yaml
kind: File
version: v1
metadata:
  fileName: file1
  fileVersion: v1
spec:
  components:
    - type: DynamoDB
      name: userTable
    - type: ECSCluster
      name: apiCluster
      connectsTo:
        - type: DynamoDB
          name: userTable
        - type: S3
          name: assetBucket
    - type: S3
      name: assetBucket
```

### Case C — Multi-region with DynamoDB

Global file:

```yaml
kind: File
version: v1
metadata:
  fileName: globalFile
  fileVersion: v1
spec:
  components:
    - type: DynamoDB
      name: userTable
```

Regional file:

```yaml
kind: File
version: v1
metadata:
  fileName: regionalFile
  fileVersion: v1
spec:
  components:
    - type: ECSCluster
      name: apiCluster
      connectsTo:
        - type: DynamoDB
          name: userTable
          fileName: globalFile
          fileVersion: v1
        - type: S3
          name: assetBucket
    - type: S3
      name: assetBucket
```

---

## Generation Flow

1. Unmarshal definition YAML.
2. Validate definition.
3. Merge runtime inputs with defaults.
4. Validate runtime inputs.
5. Resolve active components.
6. Resolve active relationships.
7. Resolve component placement.
8. Partition active components by `fileScope`.
9. Generate output file or files.
10. Convert relationships into inline `connectsTo`.
11. Add `fileName` and `fileVersion` for cross-file references.

---

## main.go

```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"

	"gopkg.in/yaml.v3"
)

type Definition struct {
	Kind          string              `yaml:"kind"`
	Version       string              `yaml:"version"`
	Inputs        map[string]InputDef `yaml:"inputs"`
	Components    []ComponentDef      `yaml:"components"`
	Relationships []Relationship      `yaml:"relationships"`
}

type InputDef struct {
	Type    string    `yaml:"type"`
	Items   *ItemsDef `yaml:"items,omitempty"`
	Values  []string  `yaml:"values,omitempty"`
	Default any       `yaml:"default,omitempty"`
}

type ItemsDef struct {
	Type string `yaml:"type"`
}

type ComponentDef struct {
	Type    string        `yaml:"type"`
	Name    string        `yaml:"name"`
	Resolve *ResolveBlock `yaml:"resolve,omitempty"`
}

type Relationship struct {
	From    Endpoint      `yaml:"from"`
	To      Endpoint      `yaml:"to"`
	Resolve *ResolveBlock `yaml:"resolve,omitempty"`
}

type Endpoint struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}

type ResolveBlock struct {
	When      *Condition `yaml:"when,omitempty"`
	FileScope string     `yaml:"fileScope,omitempty"`
	Placement string     `yaml:"placement,omitempty"`
}

type ContainsCondition struct {
	Input string `yaml:"input"`
	Value string `yaml:"value"`
}

type Condition struct {
	InputEquals map[string]any     `yaml:"input_equals,omitempty"`
	Contains    *ContainsCondition `yaml:"contains,omitempty"`
	All         []*Condition       `yaml:"all,omitempty"`
	Any         []*Condition       `yaml:"any,omitempty"`
	Not         *Condition         `yaml:"not,omitempty"`
}

type ActiveComponent struct {
	Type      string
	Name      string
	Key       string
	FileScope string
}

type OutputFile struct {
	Kind     string         `yaml:"kind"`
	Version  string         `yaml:"version"`
	Metadata OutputMetadata `yaml:"metadata"`
	Spec     OutputSpec     `yaml:"spec"`
}

type OutputMetadata struct {
	FileName    string `yaml:"fileName"`
	FileVersion string `yaml:"fileVersion"`
}

type OutputSpec struct {
	Components []OutputComponent `yaml:"components"`
}

type OutputComponent struct {
	Type       string            `yaml:"type"`
	Name       string            `yaml:"name"`
	ConnectsTo []OutputReference `yaml:"connectsTo,omitempty"`
}

type OutputReference struct {
	Type        string `yaml:"type"`
	Name        string `yaml:"name"`
	FileName    string `yaml:"fileName,omitempty"`
	FileVersion string `yaml:"fileVersion,omitempty"`
}

type GenerationResult struct {
	Files []OutputFile
}

func main() {
	definitionYAML := `
kind: Definition
version: v1

inputs:
  topology:
    type: enum
    values: [single_region, multi_region]
    default: single_region

  optional_components:
    type: array
    items:
      type: string
    default: []

components:
  - type: ECSCluster
    name: apiCluster

  - type: S3
    name: assetBucket

  - type: DynamoDB
    name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB
      placement: promoted

relationships:
  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: S3
      name: assetBucket

  - from:
      type: ECSCluster
      name: apiCluster
    to:
      type: DynamoDB
      name: userTable
    resolve:
      when:
        contains:
          input: optional_components
          value: DynamoDB
`

	var def Definition
	if err := yaml.Unmarshal([]byte(definitionYAML), &def); err != nil {
		log.Fatalf("failed to parse definition: %v", err)
	}

	if err := ValidateDefinition(def); err != nil {
		log.Fatalf("definition validation failed: %v", err)
	}

	examples := []map[string]any{
		{
			"topology":            "single_region",
			"optional_components": []any{},
		},
		{
			"topology":            "single_region",
			"optional_components": []any{"DynamoDB"},
		},
		{
			"topology":            "multi_region",
			"optional_components": []any{"DynamoDB"},
		},
	}

	for i, inputValues := range examples {
		fmt.Printf("\n=== Example %d ===\n", i+1)

		result, err := Generate(def, inputValues)
		if err != nil {
			log.Fatalf("generation failed: %v", err)
		}

		for _, file := range result.Files {
			out, err := yaml.Marshal(file)
			if err != nil {
				log.Fatalf("failed to marshal output: %v", err)
			}
			fmt.Println(string(out))
		}
	}
}

func Generate(def Definition, providedInputs map[string]any) (GenerationResult, error) {
	inputs, err := mergeInputsWithDefaults(def.Inputs, providedInputs)
	if err != nil {
		return GenerationResult{}, err
	}

	if err := validateInputs(def.Inputs, inputs); err != nil {
		return GenerationResult{}, err
	}

	activeComponents, err := resolveActiveComponents(def.Components, inputs)
	if err != nil {
		return GenerationResult{}, err
	}

	activeMap := make(map[string]ActiveComponent, len(activeComponents))
	for _, c := range activeComponents {
		activeMap[c.Key] = c
	}

	activeRelationships, err := resolveActiveRelationships(def.Relationships, activeMap, inputs)
	if err != nil {
		return GenerationResult{}, err
	}

	scopeBuckets := bucketComponentsByScope(activeComponents)
	fileNames := resolveFileNames(scopeBuckets)
	files := buildOutputFiles(scopeBuckets, activeRelationships, activeMap, fileNames)

	return GenerationResult{Files: files}, nil
}

func mergeInputsWithDefaults(defs map[string]InputDef, provided map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(defs))
	for name, def := range defs {
		if value, ok := provided[name]; ok {
			result[name] = value
		} else {
			result[name] = def.Default
		}
	}

	for name := range provided {
		if _, ok := defs[name]; !ok {
			return nil, fmt.Errorf("unknown input %q", name)
		}
	}

	return result, nil
}

func validateInputs(defs map[string]InputDef, inputs map[string]any) error {
	for name, def := range defs {
		value, ok := inputs[name]
		if !ok {
			return fmt.Errorf("missing input %q", name)
		}

		switch def.Type {
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("input %q must be boolean", name)
			}

		case "enum":
			strVal, ok := value.(string)
			if !ok {
				return fmt.Errorf("input %q must be string for enum", name)
			}
			if !stringInSlice(strVal, def.Values) {
				return fmt.Errorf("input %q has invalid value %q", name, strVal)
			}

		case "array":
			if _, ok := value.([]any); ok {
				break
			}
			if _, ok := value.([]string); ok {
				break
			}
			return fmt.Errorf("input %q must be array", name)

		default:
			return fmt.Errorf("unsupported input type %q for input %q", def.Type, name)
		}
	}
	return nil
}

func resolveActiveComponents(components []ComponentDef, inputs map[string]any) ([]ActiveComponent, error) {
	active := make([]ActiveComponent, 0, len(components))
	seen := map[string]struct{}{}

	for _, c := range components {
		ok, err := resolveWhen(c.Resolve, inputs)
		if err != nil {
			return nil, fmt.Errorf("component %s:%s resolve error: %w", c.Type, c.Name, err)
		}
		if !ok {
			continue
		}

		scope, err := resolveComponentScope(c.Resolve, inputs)
		if err != nil {
			return nil, fmt.Errorf("component %s:%s scope error: %w", c.Type, c.Name, err)
		}

		key := componentKey(c.Type, c.Name)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate active component %s", key)
		}
		seen[key] = struct{}{}

		active = append(active, ActiveComponent{
			Type:      c.Type,
			Name:      c.Name,
			Key:       key,
			FileScope: scope,
		})
	}

	return active, nil
}

func resolveActiveRelationships(
	relationships []Relationship,
	activeComponents map[string]ActiveComponent,
	inputs map[string]any,
) ([]Relationship, error) {
	result := make([]Relationship, 0, len(relationships))

	for _, r := range relationships {
		ok, err := resolveWhen(r.Resolve, inputs)
		if err != nil {
			return nil, fmt.Errorf(
				"relationship %s:%s -> %s:%s resolve error: %w",
				r.From.Type, r.From.Name, r.To.Type, r.To.Name, err,
			)
		}
		if !ok {
			continue
		}

		fromKey := componentKey(r.From.Type, r.From.Name)
		toKey := componentKey(r.To.Type, r.To.Name)

		if _, ok := activeComponents[fromKey]; !ok {
			continue
		}
		if _, ok := activeComponents[toKey]; !ok {
			continue
		}

		result = append(result, r)
	}

	return result, nil
}

func resolveWhen(resolve *ResolveBlock, inputs map[string]any) (bool, error) {
	if resolve == nil || resolve.When == nil {
		return true, nil
	}
	return evalCondition(resolve.When, inputs)
}

func resolveComponentScope(resolve *ResolveBlock, inputs map[string]any) (string, error) {
	if resolve == nil {
		return "regional", nil
	}

	if resolve.FileScope != "" && resolve.Placement != "" {
		return "", errors.New("fileScope and placement cannot both be set")
	}

	if resolve.FileScope != "" {
		return resolve.FileScope, nil
	}

	switch resolve.Placement {
	case "":
		return "regional", nil
	case "promoted":
		topologyValue, ok := inputs["topology"]
		if !ok {
			return "", errors.New("topology input is required for placement=promoted")
		}
		topology, ok := topologyValue.(string)
		if !ok {
			return "", errors.New("topology input must be a string")
		}
		switch topology {
		case "single_region":
			return "regional", nil
		case "multi_region":
			return "global", nil
		default:
			return "", fmt.Errorf("unsupported topology %q for placement=promoted", topology)
		}
	default:
		return "", fmt.Errorf("unsupported placement %q", resolve.Placement)
	}
}

func bucketComponentsByScope(active []ActiveComponent) map[string][]ActiveComponent {
	buckets := map[string][]ActiveComponent{}

	for _, c := range active {
		buckets[c.FileScope] = append(buckets[c.FileScope], c)
	}

	for scope := range buckets {
		sort.Slice(buckets[scope], func(i, j int) bool {
			a := buckets[scope][i]
			b := buckets[scope][j]
			if a.Type == b.Type {
				return a.Name < b.Name
			}
			return a.Type < b.Type
		})
	}

	return buckets
}

func resolveFileNames(scopeBuckets map[string][]ActiveComponent) map[string]string {
	fileNames := map[string]string{}

	if len(scopeBuckets) == 1 {
		for scope := range scopeBuckets {
			switch scope {
			case "global":
				fileNames["global"] = "globalFile"
			default:
				fileNames[scope] = "file1"
			}
		}
		return fileNames
	}

	if _, ok := scopeBuckets["global"]; ok {
		fileNames["global"] = "globalFile"
	}
	if _, ok := scopeBuckets["regional"]; ok {
		fileNames["regional"] = "regionalFile"
	}

	for scope := range scopeBuckets {
		if _, ok := fileNames[scope]; !ok {
			fileNames[scope] = scope + "File"
		}
	}

	return fileNames
}

func buildOutputFiles(
	scopeBuckets map[string][]ActiveComponent,
	relationships []Relationship,
	activeMap map[string]ActiveComponent,
	fileNames map[string]string,
) []OutputFile {
	var scopes []string
	for scope := range scopeBuckets {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)

	files := make([]OutputFile, 0, len(scopes))

	for _, scope := range scopes {
		components := scopeBuckets[scope]
		componentOutputs := make([]OutputComponent, 0, len(components))

		for _, c := range components {
			outComp := OutputComponent{
				Type: c.Type,
				Name: c.Name,
			}

			var refs []OutputReference
			for _, rel := range relationships {
				if rel.From.Type != c.Type || rel.From.Name != c.Name {
					continue
				}

				targetKey := componentKey(rel.To.Type, rel.To.Name)
				target, ok := activeMap[targetKey]
				if !ok {
					continue
				}

				ref := OutputReference{
					Type: target.Type,
					Name: target.Name,
				}

				if target.FileScope != c.FileScope {
					ref.FileName = fileNames[target.FileScope]
					ref.FileVersion = "v1"
				}

				refs = append(refs, ref)
			}

			sort.Slice(refs, func(i, j int) bool {
				if refs[i].Type == refs[j].Type {
					return refs[i].Name < refs[j].Name
				}
				return refs[i].Type < refs[j].Type
			})

			if len(refs) > 0 {
				outComp.ConnectsTo = refs
			}

			componentOutputs = append(componentOutputs, outComp)
		}

		file := OutputFile{
			Kind:    "File",
			Version: "v1",
			Metadata: OutputMetadata{
				FileName:    fileNames[scope],
				FileVersion: "v1",
			},
			Spec: OutputSpec{
				Components: componentOutputs,
			},
		}

		files = append(files, file)
	}

	return files
}

func evalCondition(c *Condition, inputs map[string]any) (bool, error) {
	if c == nil {
		return true, nil
	}

	setCount := 0
	if len(c.InputEquals) > 0 {
		setCount++
	}
	if c.Contains != nil {
		setCount++
	}
	if len(c.All) > 0 {
		setCount++
	}
	if len(c.Any) > 0 {
		setCount++
	}
	if c.Not != nil {
		setCount++
	}

	if setCount != 1 {
		return false, errors.New("condition must contain exactly one operator")
	}

	if len(c.InputEquals) > 0 {
		if len(c.InputEquals) != 1 {
			return false, errors.New("input_equals must contain exactly one key")
		}
		for key, expected := range c.InputEquals {
			actual, ok := inputs[key]
			if !ok {
				return false, fmt.Errorf("input %q not found", key)
			}
			return valuesEqual(actual, expected), nil
		}
	}

	if c.Contains != nil {
		actual, ok := inputs[c.Contains.Input]
		if !ok {
			return false, fmt.Errorf("input %q not found", c.Contains.Input)
		}
		return arrayContains(actual, c.Contains.Value), nil
	}

	if len(c.All) > 0 {
		for _, child := range c.All {
			ok, err := evalCondition(child, inputs)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	}

	if len(c.Any) > 0 {
		for _, child := range c.Any {
			ok, err := evalCondition(child, inputs)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	if c.Not != nil {
		ok, err := evalCondition(c.Not, inputs)
		if err != nil {
			return false, err
		}
		return !ok, nil
	}

	return false, errors.New("invalid condition")
}

func arrayContains(arr any, expected string) bool {
	switch v := arr.(type) {
	case []string:
		for _, item := range v {
			if item == expected {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

func valuesEqual(a, b any) bool {
	aj, errA := json.Marshal(a)
	bj, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return a == b
	}
	return string(aj) == string(bj)
}

func componentKey(componentType, name string) string {
	return componentType + ":" + name
}

func stringInSlice(v string, values []string) bool {
	for _, candidate := range values {
		if candidate == v {
			return true
		}
	}
	return false
}

```

---

## validation.go

```go
package main

import "fmt"

func ValidateDefinition(def Definition) error {
	if def.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if def.Kind != "Definition" {
		return fmt.Errorf("kind must be Definition")
	}
	if def.Version == "" {
		return fmt.Errorf("version is required")
	}

	if len(def.Inputs) == 0 {
		return fmt.Errorf("inputs are required")
	}
	if err := validateInputDefs(def.Inputs); err != nil {
		return err
	}

	if len(def.Components) == 0 {
		return fmt.Errorf("at least one component is required")
	}

	componentSet := buildComponentSet(def.Components)

	if err := validateComponents(def.Components, def.Inputs, componentSet); err != nil {
		return err
	}

	if err := validateRelationships(def.Relationships, def.Components, def.Inputs, componentSet); err != nil {
		return err
	}

	return nil
}

func buildComponentSet(components []ComponentDef) map[string]struct{} {
	componentSet := map[string]struct{}{}
	for _, c := range components {
		if c.Type == "" || c.Name == "" {
			continue
		}
		componentSet[componentKey(c.Type, c.Name)] = struct{}{}
	}
	return componentSet
}

func validateInputDefs(inputs map[string]InputDef) error {
	for name, def := range inputs {
		if name == "" {
			return fmt.Errorf("input name cannot be empty")
		}

		switch def.Type {
		case "boolean":
			if len(def.Values) > 0 {
				return fmt.Errorf("input %q of type boolean cannot define values", name)
			}
			if def.Items != nil {
				return fmt.Errorf("input %q of type boolean cannot define items", name)
			}
			if def.Default != nil {
				if _, ok := def.Default.(bool); !ok {
					return fmt.Errorf("input %q default must be boolean", name)
				}
			}

		case "enum":
			if len(def.Values) == 0 {
				return fmt.Errorf("input %q of type enum must define values", name)
			}
			if def.Items != nil {
				return fmt.Errorf("input %q of type enum cannot define items", name)
			}
			if def.Default != nil {
				defaultValue, ok := def.Default.(string)
				if !ok {
					return fmt.Errorf("input %q default must be string for enum", name)
				}
				if !stringInSlice(defaultValue, def.Values) {
					return fmt.Errorf("input %q default value %q must be one of %v", name, defaultValue, def.Values)
				}
			}

		case "array":
			if def.Items == nil {
				return fmt.Errorf("input %q of type array must define items", name)
			}
			if def.Items.Type != "string" {
				return fmt.Errorf("input %q array items.type must be string", name)
			}
			if len(def.Values) > 0 {
				return fmt.Errorf("input %q of type array cannot define values", name)
			}
			if def.Default != nil {
				switch arr := def.Default.(type) {
				case []any:
					for i, item := range arr {
						if _, ok := item.(string); !ok {
							return fmt.Errorf("input %q default[%d] must be string", name, i)
						}
					}
				case []string:
				default:
					return fmt.Errorf("input %q default must be an array", name)
				}
			}

		default:
			return fmt.Errorf("input %q has unsupported type %q", name, def.Type)
		}
	}
	return nil
}

func validateComponents(
	components []ComponentDef,
	inputs map[string]InputDef,
	componentSet map[string]struct{},
) error {
	seen := map[string]struct{}{}

	for i, c := range components {
		path := fmt.Sprintf("components[%d]", i)

		if c.Type == "" {
			return fmt.Errorf("%s.type is required", path)
		}
		if c.Name == "" {
			return fmt.Errorf("%s.name is required", path)
		}

		key := componentKey(c.Type, c.Name)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%s duplicates component %q", path, key)
		}
		seen[key] = struct{}{}

		if c.Resolve != nil {
			if c.Resolve.FileScope != "" && c.Resolve.Placement != "" {
				return fmt.Errorf("%s (%s) cannot define both fileScope and placement", path, key)
			}

			if c.Resolve.FileScope != "" && !isValidFileScope(c.Resolve.FileScope) {
				return fmt.Errorf("%s (%s) has invalid fileScope %q", path, key, c.Resolve.FileScope)
			}

			if c.Resolve.Placement != "" && !isValidPlacement(c.Resolve.Placement) {
				return fmt.Errorf("%s (%s) has invalid placement %q", path, key, c.Resolve.Placement)
			}

			if c.Resolve.When != nil {
				if err := validateCondition(c.Resolve.When, inputs, componentSet, path+".resolve.when"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateRelationships(
	relationships []Relationship,
	components []ComponentDef,
	inputs map[string]InputDef,
	componentSet map[string]struct{},
) error {
	seen := map[string]struct{}{}

	for i, r := range relationships {
		path := fmt.Sprintf("relationships[%d]", i)

		if r.From.Type == "" {
			return fmt.Errorf("%s.from.type is required", path)
		}
		if r.From.Name == "" {
			return fmt.Errorf("%s.from.name is required", path)
		}
		if r.To.Type == "" {
			return fmt.Errorf("%s.to.type is required", path)
		}
		if r.To.Name == "" {
			return fmt.Errorf("%s.to.name is required", path)
		}

		fromKey := componentKey(r.From.Type, r.From.Name)
		toKey := componentKey(r.To.Type, r.To.Name)

		if _, ok := componentSet[fromKey]; !ok {
			return fmt.Errorf("%s references unknown from component %q", path, fromKey)
		}
		if _, ok := componentSet[toKey]; !ok {
			return fmt.Errorf("%s references unknown to component %q", path, toKey)
		}

		relationshipKey := fromKey + "->" + toKey
		if _, exists := seen[relationshipKey]; exists {
			return fmt.Errorf("%s duplicates relationship %q", path, relationshipKey)
		}
		seen[relationshipKey] = struct{}{}

		if r.Resolve != nil {
			if r.Resolve.FileScope != "" || r.Resolve.Placement != "" {
				return fmt.Errorf("%s cannot define fileScope or placement on relationships", path)
			}
			if r.Resolve.When != nil {
				if err := validateCondition(r.Resolve.When, inputs, componentSet, path+".resolve.when"); err != nil {
					return err
				}
			}
		}
	}

	_ = components
	return nil
}

func validateCondition(c *Condition, inputs map[string]InputDef, componentSet map[string]struct{}, path string) error {
	if c == nil {
		return nil
	}

	setCount := 0
	if len(c.InputEquals) > 0 {
		setCount++
	}
	if c.Contains != nil {
		setCount++
	}
	if len(c.All) > 0 {
		setCount++
	}
	if len(c.Any) > 0 {
		setCount++
	}
	if c.Not != nil {
		setCount++
	}

	if setCount != 1 {
		return fmt.Errorf("%s must contain exactly one operator", path)
	}

	if len(c.InputEquals) > 0 {
		if len(c.InputEquals) != 1 {
			return fmt.Errorf("%s.input_equals must contain exactly one key", path)
		}

		for inputName, expectedValue := range c.InputEquals {
			inputDef, ok := inputs[inputName]
			if !ok {
				return fmt.Errorf("%s.input_equals references unknown input %q", path, inputName)
			}

			switch inputDef.Type {
			case "boolean":
				if _, ok := expectedValue.(bool); !ok {
					return fmt.Errorf("%s.input_equals for %q must use a boolean value", path, inputName)
				}
			case "enum":
				strVal, ok := expectedValue.(string)
				if !ok {
					return fmt.Errorf("%s.input_equals for %q must use a string value", path, inputName)
				}
				if !stringInSlice(strVal, inputDef.Values) {
					return fmt.Errorf("%s.input_equals for %q uses invalid enum value %q", path, inputName, strVal)
				}
			default:
				return fmt.Errorf("%s.input_equals cannot be used with input type %q for %q", path, inputDef.Type, inputName)
			}
		}
		return nil
	}

	if c.Contains != nil {
		if c.Contains.Input == "" {
			return fmt.Errorf("%s.contains.input is required", path)
		}
		if c.Contains.Value == "" {
			return fmt.Errorf("%s.contains.value is required", path)
		}

		inputDef, ok := inputs[c.Contains.Input]
		if !ok {
			return fmt.Errorf("%s.contains references unknown input %q", path, c.Contains.Input)
		}
		if inputDef.Type != "array" {
			return fmt.Errorf("%s.contains requires array input, got %q", path, inputDef.Type)
		}
		if inputDef.Items == nil || inputDef.Items.Type != "string" {
			return fmt.Errorf("%s.contains requires array input with string items", path)
		}

		if c.Contains.Input == "optional_components" {
			if _, ok := componentSet[c.Contains.Value]; !ok {
				return fmt.Errorf("%s.contains.value %q is not a known component key; expected type:name", path, c.Contains.Value)
			}
		}

		return nil
	}

	if len(c.All) > 0 {
		for i, child := range c.All {
			if child == nil {
				return fmt.Errorf("%s.all[%d] cannot be null", path, i)
			}
			if err := validateCondition(child, inputs, componentSet, fmt.Sprintf("%s.all[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	}

	if len(c.Any) > 0 {
		for i, child := range c.Any {
			if child == nil {
				return fmt.Errorf("%s.any[%d] cannot be null", path, i)
			}
			if err := validateCondition(child, inputs, componentSet, fmt.Sprintf("%s.any[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	}

	if c.Not != nil {
		return validateCondition(c.Not, inputs, componentSet, path+".not")
	}

	return fmt.Errorf("%s is invalid", path)
}

func isValidFileScope(scope string) bool {
	return scope == "regional" || scope == "global"
}

func isValidPlacement(placement string) bool {
	return placement == "promoted"
}

```
