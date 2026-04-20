package response

import "github.com/gin-gonic/gin"

type SuccessEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorEnvelope struct {
	Success bool         `json:"success"`
	Error   ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func JSONSuccess(c *gin.Context, statusCode int, data any, message string) {
	c.JSON(statusCode, SuccessEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func JSONError(c *gin.Context, statusCode int, code, message string, details any) {
	c.JSON(statusCode, ErrorEnvelope{
		Success: false,
		Error: ErrorPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
