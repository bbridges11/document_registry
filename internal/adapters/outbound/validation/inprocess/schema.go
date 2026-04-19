package inprocess

// DefinitionSchema represents the structure of a definition YAML file
type DefinitionSchema struct {
	Kind          string               `yaml:"kind" validate:"required,eq=Definition"`
	Version       string               `yaml:"version" validate:"required,oneof=v1"`
	Inputs        map[string]InputSpec `yaml:"inputs"`
	Components    []ComponentSpec      `yaml:"components" validate:"required,min=1,dive"`
	Relationships []RelationshipSpec   `yaml:"relationships" validate:"dive"`
}

// InputSpec defines an input parameter
type InputSpec struct {
	Type    string   `yaml:"type" validate:"required,oneof=string boolean enum number"`
	Values  []string `yaml:"values"`
	Default any      `yaml:"default"`
}

// ComponentSpec defines a component in the architecture
type ComponentSpec struct {
	Type                string            `yaml:"type" validate:"required"`
	Name                string            `yaml:"name" validate:"required"`
	Required            bool              `yaml:"required"`
	FileScope           string            `yaml:"fileScope" validate:"omitempty,oneof=regional global"`
	FileScopeByTopology map[string]string `yaml:"fileScopeByTopology"`
	IncludeWhen         *ConditionSpec    `yaml:"includeWhen"`
}

// RelationshipSpec defines a relationship between components
type RelationshipSpec struct {
	From RelationshipEndpoint `yaml:"from" validate:"required"`
	To   RelationshipEndpoint `yaml:"to" validate:"required"`
}

// RelationshipEndpoint defines one end of a relationship
type RelationshipEndpoint struct {
	Type string `yaml:"type" validate:"required"`
	Name string `yaml:"name" validate:"required"`
}

// ConditionSpec defines conditions for component inclusion
type ConditionSpec struct {
	InputEquals map[string]any `yaml:"input_equals"`
}
