package user

type UpdateUserRequest struct {
	Name        string                 `json:"name" validate:"omitempty,min=2,max=50"`
	PhoneNumber string                 `json:"phone_number" validate:"omitempty,e164"`
	Preferences map[string]interface{} `json:"preferences"`
}
