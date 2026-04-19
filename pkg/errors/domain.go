package errors

var (
	ErrNotFound        = New(CodeNotFound, "resource not found")
	ErrAlreadyExists   = New(CodeConflict, "resource already exists")
	ErrInvalidVersion  = New(CodeInvalidArgument, "invalid version format")
	ErrVersionConflict = New(CodeVersionConflict, "version conflict")
	ErrInvalidState    = New(CodeInvalidState, "invalid state transition")
	ErrImmutable       = New(CodePrecondition, "resource is immutable")
	ErrUnauthorized    = New(CodeUnauthorized, "unauthorized")
	ErrForbidden       = New(CodeForbidden, "forbidden")
)
