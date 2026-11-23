package models

import "time"

type UserResponse struct {
	ID         uint     `json:"id"`
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	IsActive   bool     `json:"is_active"`
	IsVerified bool     `json:"is_verified"`
	Roles      []string `json:"roles"`
	CreatedAt  time.Time
}

func ToUserResponse(user User) UserResponse {
	roles := make([]string, 0, len(user.Roles))

	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	return UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		Roles:      roles,
		CreatedAt:  user.CreatedAt,
	}
}