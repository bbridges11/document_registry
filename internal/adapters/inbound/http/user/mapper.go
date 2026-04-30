package user

import (
	appUser "github.com/bbridges_11/document-registry/internal/application/user"
)

// ToUserResponse converts application DTO to HTTP response
func ToUserResponse(dto appUser.UserDTO) UserResponse {
	return UserResponse{
		ID:         dto.ID,
		ExternalID: dto.ExternalID,
		Email:      dto.Email,
		Name:       dto.Name,
		Role:       dto.Role,
		Active:     dto.Active,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  dto.UpdatedAt,
	}
}

// ToUsersResponse converts application DTOs to HTTP response
func ToUsersResponse(dtos []appUser.UserDTO) UsersResponse {
	users := make([]UserResponse, len(dtos))
	for i, dto := range dtos {
		users[i] = ToUserResponse(dto)
	}
	return UsersResponse{
		Users: users,
		Count: len(users),
	}
}
