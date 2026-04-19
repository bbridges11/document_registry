package document

import "context"

type Service struct {
	commands *CommandService
	queries  *QueryService
}

func NewService(commands *CommandService, queries *QueryService) UseCase {
	return &Service{commands: commands, queries: queries}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	doc, ver, valResult, err := s.commands.CreateDocumentWithVersion(ctx, CreateDocumentWithVersionCommand{
		Name:         input.Name,
		Description:  input.Description,
		Tags:         input.Tags,
		DocumentType: input.DocumentType,
		Version:      input.Version,
		Content:      input.Content,
		Metadata:     input.Metadata,
		CreatedBy:    input.Actor.UserID,
	})
	if err != nil {
		// Check if it's a validation error with results
		if valResult != nil && !valResult.Valid {
			// Return partial output with validation errors
			return CreateOutput{
				ValidationResult: toValidationResultView(valResult),
			}, err
		}
		return CreateOutput{}, err
	}

	output := CreateOutput{
		Document: toDocumentView(doc),
		Version:  toVersionView(ver),
	}

	if valResult != nil {
		output.ValidationResult = toValidationResultView(valResult)
	}

	return output, nil
}

func (s *Service) Get(ctx context.Context, input GetInput) (DocumentView, error) {
	dto, err := s.queries.GetDocument(ctx, GetDocumentQuery{ID: input.ID}, input.Actor.UserID)
	if err != nil {
		return DocumentView{}, err
	}
	return toDocumentDTOView(dto), nil
}

func (s *Service) List(ctx context.Context, input ListInput) ([]DocumentView, error) {
	dtos, err := s.queries.ListDocuments(ctx, ListDocumentsQuery{Limit: input.Limit, Offset: input.Offset}, input.Actor.UserID)
	if err != nil {
		return nil, err
	}
	views := make([]DocumentView, 0, len(dtos))
	for _, dto := range dtos {
		views = append(views, toDocumentDTOView(dto))
	}
	return views, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	return s.commands.UpdateDocument(ctx, UpdateDocumentCommand{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Tags:        input.Tags,
		UpdatedBy:   input.Actor.UserID,
	})
}

func (s *Service) Search(ctx context.Context, input SearchInput) (SearchOutput, error) {
	return s.queries.SearchDocuments(ctx, input)
}
