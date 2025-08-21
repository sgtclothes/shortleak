package dto

type LinkRequest struct {
	URL string `json:"url" validate:"required,url"`
}
