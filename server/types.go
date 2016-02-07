package server

type RequestType int

const (
	REQUEST_RUN RequestType = iota
	REQUEST_WAIT
	REQUEST_GETPID
)

type Request struct {
	Type   RequestType
	HasFds bool
}

// The server's response.
// If OK or ERR, message will contain useful text.
type ResponseType int

const (
	RESPONSE_ERR ResponseType = iota
	RESPONSE_GETPID
)

type Response struct {
	Type    ResponseType
	Message string
	Getpid  *ResponseGetpid
}

type ResponseGetpid struct {
	Pid int
}
