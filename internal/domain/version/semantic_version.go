package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"golang.org/x/mod/semver"
)

type SemanticVersion struct {
	raw   string
	major int
	minor int
	patch int
}

func NewSemanticVersion(version string) (*SemanticVersion, error) {
	if version == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "version is required")
	}

	// Ensure version starts with 'v' for semver library
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !semver.IsValid(version) {
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("invalid semantic version: %s", version))
	}

	// Parse major.minor.patch
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) != 3 {
		return nil, errors.New(errors.CodeInvalidArgument, "version must be in format x.y.z")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidArgument, "invalid major version", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidArgument, "invalid minor version", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidArgument, "invalid patch version", err)
	}

	return &SemanticVersion{
		raw:   strings.TrimPrefix(version, "v"),
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

func (v *SemanticVersion) String() string {
	return v.raw
}

func (v *SemanticVersion) Compare(other *SemanticVersion) int {
	v1 := "v" + v.raw
	v2 := "v" + other.raw
	return semver.Compare(v1, v2)
}

func (v *SemanticVersion) IsGreaterThan(other *SemanticVersion) bool {
	return v.Compare(other) > 0
}
