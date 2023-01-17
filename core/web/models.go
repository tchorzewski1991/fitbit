package web

const (
	success = "success"
)

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Error(err error) Response {
	return Response{Error: err.Error()}
}

func Success() Response {
	return Response{Message: success}
}
