package web

const (
	success = "success"
)

// Response encapsulates the details of generic web response used withing the app.
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Error is a constructor for generic web error response used by various http endpoints.
func Error(err error) Response {
	return Response{Error: err.Error()}
}

// Success is a constructor for generic web success response used by various http endpoints.
func Success() Response {
	return Response{Message: success}
}
