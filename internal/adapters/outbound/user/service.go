package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type ServiceAdapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewServiceAdapter(cfg config.UserServiceConfig) outbound.UserService {
	return &ServiceAdapter{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

type userResponse struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

func (a *ServiceAdapter) GetByID(ctx context.Context, userID string) (user *outbound.User, err error) {
	defer err2.Handle(&err)

	url := fmt.Sprintf("%s/users/%s", a.baseURL, userID)
	req := try.To1(http.NewRequestWithContext(ctx, "GET", url, nil))

	resp := try.To1(a.httpClient.Do(req))
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.CodeDependency, fmt.Sprintf("user service returned %d", resp.StatusCode))
	}

	var userResp userResponse
	try.To(json.NewDecoder(resp.Body).Decode(&userResp))

	return a.toUser(&userResp), nil
}

func (a *ServiceAdapter) GetByIDs(ctx context.Context, userIDs []string) (users []*outbound.User, err error) {
	defer err2.Handle(&err)

	url := fmt.Sprintf("%s/users?ids=%s", a.baseURL, strings.Join(userIDs, ","))
	req := try.To1(http.NewRequestWithContext(ctx, "GET", url, nil))

	resp := try.To1(a.httpClient.Do(req))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.CodeDependency, fmt.Sprintf("user service returned %d", resp.StatusCode))
	}

	var userResps []userResponse
	try.To(json.NewDecoder(resp.Body).Decode(&userResps))

	users = make([]*outbound.User, 0, len(userResps))
	for i := range userResps {
		users = append(users, a.toUser(&userResps[i]))
	}

	return users, nil
}

func (a *ServiceAdapter) GetApprovalRoles(ctx context.Context, userID string) (roles []approval.ApprovalRole, err error) {
	defer err2.Handle(&err)

	url := fmt.Sprintf("%s/users/%s/approval-roles", a.baseURL, userID)
	req := try.To1(http.NewRequestWithContext(ctx, "GET", url, nil))

	resp := try.To1(a.httpClient.Do(req))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.CodeDependency, fmt.Sprintf("user service returned %d", resp.StatusCode))
	}

	var roleStrs []string
	try.To(json.NewDecoder(resp.Body).Decode(&roleStrs))

	roles = make([]approval.ApprovalRole, 0, len(roleStrs))
	for _, roleStr := range roleStrs {
		roles = append(roles, approval.ApprovalRole(roleStr))
	}

	return roles, nil
}

func (a *ServiceAdapter) HealthCheck(ctx context.Context) (err error) {
	defer err2.Handle(&err)

	url := fmt.Sprintf("%s/health", a.baseURL)
	req := try.To1(http.NewRequestWithContext(ctx, "GET", url, nil))

	resp := try.To1(a.httpClient.Do(req))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(errors.CodeDependency, "user service unhealthy")
	}

	return nil
}

func (a *ServiceAdapter) toUser(resp *userResponse) *outbound.User {
	roles := make([]approval.ApprovalRole, 0, len(resp.Roles))
	for _, roleStr := range resp.Roles {
		roles = append(roles, approval.ApprovalRole(roleStr))
	}

	return &outbound.User{
		ID:    resp.ID,
		Name:  resp.Name,
		Email: resp.Email,
		Roles: roles,
	}
}
