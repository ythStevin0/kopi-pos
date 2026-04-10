package model

// APIResponse adalah wrapper standar untuk semua response API.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(message string, data interface{}) APIResponse {
	return APIResponse{Success: true, Message: message, Data: data}
}

func Fail(message string) APIResponse {
	return APIResponse{Success: false, Message: message}
}
