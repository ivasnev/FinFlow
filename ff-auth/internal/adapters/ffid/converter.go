package ffid

import (
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
)

// convertUserDTO преобразует api.UserDTO в адаптерный UserDTO
func convertUserDTO(apiUser *api.UserDTO) *UserDTO {
	if apiUser == nil {
		return nil
	}

	var avatarID *string
	if apiUser.AvatarId != nil {
		str := apiUser.AvatarId.String()
		avatarID = &str
	}

	return &UserDTO{
		ID:        apiUser.Id,
		Email:     string(apiUser.Email),
		Nickname:  apiUser.Nickname,
		Name:      apiUser.Name,
		Phone:     apiUser.Phone,
		AvatarID:  avatarID,
		Birthdate: apiUser.Birthdate,
		CreatedAt: apiUser.CreatedAt,
		UpdatedAt: apiUser.UpdatedAt,
	}
}
