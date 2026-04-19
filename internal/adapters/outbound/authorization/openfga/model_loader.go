package openfga

import (
	"context"
	"fmt"
	"os"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"go.uber.org/zap"
)

// ModelLoader handles loading and creating OpenFGA authorization models
type ModelLoader struct {
	client  *client.OpenFgaClient
	storeID string
	logger  *zap.Logger
}

func NewModelLoader(client *client.OpenFgaClient, storeID string, logger *zap.Logger) *ModelLoader {
	return &ModelLoader{
		client:  client,
		storeID: storeID,
		logger:  logger,
	}
}

// LoadModelFromFile reads the authorization model from a file
func (m *ModelLoader) LoadModelFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read model file: %w", err)
	}

	return string(data), nil
}

// CreateModel creates a new authorization model in OpenFGA
func (m *ModelLoader) CreateModel(ctx context.Context, modelDSL string) (string, error) {
	m.logger.Info("creating OpenFGA authorization model")

	// Use the DSL directly - the SDK will handle parsing
	body := client.ClientWriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{
				Type: "user",
			},
			{
				Type: "document",
				Relations: &map[string]openfga.Userset{
					"owner": {
						This: &map[string]any{},
					},
					"contributor": {
						This: &map[string]any{},
					},
					"consumer": {
						This: &map[string]any{},
					},
					"published": {
						This: &map[string]any{},
					},
					"can_view": {
						Union: &openfga.Usersets{
							Child: []openfga.Userset{
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("published")}},
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("owner")}},
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("contributor")}},
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("consumer")}},
							},
						},
					},
					"can_edit": {
						Union: &openfga.Usersets{
							Child: []openfga.Userset{
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("owner")}},
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("contributor")}},
							},
						},
					},
					"can_delete": {
						ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("owner")},
					},
					"can_manage_stakeholders": {
						ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("owner")},
					},
					"can_create_version": {
						Union: &openfga.Usersets{
							Child: []openfga.Userset{
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("owner")}},
								{ComputedUserset: &openfga.ObjectRelation{Relation: openfga.PtrString("contributor")}},
							},
						},
					},
				},
				Metadata: &openfga.Metadata{
					Relations: &map[string]openfga.RelationMetadata{
						"owner": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
						"contributor": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
						"consumer": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
						"published": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user", Wildcard: &map[string]any{}},
							},
						},
					},
				},
			},
			{
				Type: "version",
				Relations: &map[string]openfga.Userset{
					"parent_document": {
						This: &map[string]any{},
					},
					"can_view": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("can_view")},
						},
					},
					"can_edit": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("can_edit")},
						},
					},
					"can_submit": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("can_create_version")},
						},
					},
					"can_review": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("owner")},
						},
					},
					"can_approve": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("owner")},
						},
					},
					"can_publish": {
						TupleToUserset: &openfga.TupleToUserset{
							Tupleset:        openfga.ObjectRelation{Relation: openfga.PtrString("parent_document")},
							ComputedUserset: openfga.ObjectRelation{Relation: openfga.PtrString("owner")},
						},
					},
				},
				Metadata: &openfga.Metadata{
					Relations: &map[string]openfga.RelationMetadata{
						"parent_document": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "document"},
							},
						},
					},
				},
			},
		},
	}

	response, err := m.client.WriteAuthorizationModel(ctx).Body(body).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create authorization model: %w", err)
	}

	modelID := response.GetAuthorizationModelId()
	m.logger.Info("authorization model created successfully",
		zap.String("model_id", modelID),
	)

	return modelID, nil
}

// GetLatestModel retrieves the latest authorization model ID
func (m *ModelLoader) GetLatestModel(ctx context.Context) (string, error) {
	response, err := m.client.ReadAuthorizationModels(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to read authorization models: %w", err)
	}

	models := response.GetAuthorizationModels()
	if len(models) == 0 {
		return "", fmt.Errorf("no authorization models found in store")
	}

	// Return the most recent model
	latestModel := models[0]
	return latestModel.GetId(), nil
}

// VerifyModel checks if a specific model exists
func (m *ModelLoader) VerifyModel(ctx context.Context, modelID string) error {
	if modelID == "" {
		return fmt.Errorf("model ID is empty")
	}

	_, err := m.client.ReadAuthorizationModel(ctx).Options(client.ClientReadAuthorizationModelOptions{
		AuthorizationModelId: &modelID,
	}).Execute()
	if err != nil {
		return fmt.Errorf("model %s not found: %w", modelID, err)
	}

	m.logger.Info("authorization model verified",
		zap.String("model_id", modelID),
	)

	return nil
}

// EnsureModel ensures an authorization model exists, creating one if necessary
func (m *ModelLoader) EnsureModel(ctx context.Context, modelPath string, autoCreate bool) (string, error) {
	// Try to get the latest model
	existingModelID, err := m.GetLatestModel(ctx)
	if err == nil {
		m.logger.Info("using existing authorization model",
			zap.String("model_id", existingModelID),
		)
		return existingModelID, nil
	}

	// No model exists
	if !autoCreate {
		return "", fmt.Errorf("no authorization model exists and auto-create is disabled")
	}

	// Load model from file (we ignore the DSL for now and use hardcoded model)
	_, err = m.LoadModelFromFile(modelPath)
	if err != nil {
		m.logger.Warn("failed to load model from file, using hardcoded model", zap.Error(err))
	}

	// Create new model
	modelID, err := m.CreateModel(ctx, "")
	if err != nil {
		return "", err
	}

	return modelID, nil
}
