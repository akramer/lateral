package server

type RequestType int

const (
	REQUEST_RUN RequestType = iota
	REQUEST_WAIT
	REQUEST_GETPID
	REQUEST_KILL
)

type Request struct {
	Type   RequestType
	HasFds bool
	// List of Fds to be transferred by SendRequest
	Fds []int
	// Filled in on receiving side - list of fd numbers corresponding to
	// the original FD numbers above
	ReceivedFds []int
	Run         *RequestRun
}

type RequestRun struct {
	Args []string
	Env  []string
	Cwd  string
}

// The server's response.
// If OK or ERR, message will contain useful text.
type ResponseType int

const (
	RESPONSE_ERR ResponseType = iota
	RESPONSE_OK
	RESPONSE_GETPID
	RESPONSE_WAIT
)

type Response struct {
	Type    ResponseType
	Message string
	Getpid  *ResponseGetpid
	Wait    *ResponseWait
}

type ResponseGetpid struct {
	Pid int
}

type ResponseWait struct {
	ExitStatus int
}
