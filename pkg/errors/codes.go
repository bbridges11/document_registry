package errors

type Code string

const (
	CodeInternal         Code = "internal"
	CodeInvalidArgument  Code = "invalid_argument"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodeUnauthorized     Code = "unauthorized"
	CodeForbidden        Code = "forbidden"
	CodePrecondition     Code = "precondition_failed"
	CodeDependency       Code = "dependency_unavailable"
	CodeInvalidState     Code = "invalid_state"
	CodeVersionConflict  Code = "version_conflict"
	CodeValidationFailed Code = "validation_failed"
	CodeNotSupported     Code = "not_supported"

	// Deprecation-specific error codes
	CodeVersionNotPublished      Code = "version_not_published"
	CodeVersionAlreadyDeprecated Code = "version_already_deprecated"
	CodeVersionDeprecating       Code = "version_deprecating"
	CodeMultiplePublished        Code = "multiple_published"
	CodeCannotDeprecate          Code = "cannot_deprecate"
	CodeDeprecationNotFound      Code = "deprecation_not_found"
	CodeDeprecationNotPending    Code = "deprecation_not_pending"
)
