package errors

var (
	ErrDependencyUnavailable = New(CodeDependency, "required dependency unavailable")
	ErrStorageFailure        = New(CodeInternal, "storage operation failed")
	ErrAuthorizationFailure  = New(CodeInternal, "authorization check failed")
	ErrUserServiceFailure    = New(CodeDependency, "user service unavailable")
)
