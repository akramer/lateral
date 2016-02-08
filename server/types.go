package server

type RequestType int

const (
	REQUEST_RUN RequestType = iota
	REQUEST_WAIT
	REQUEST_GETPID
	REQUEST_KILL
	REQUEST_SHUTDOWN
	REQUEST_CONFIG
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
	Config      *RequestConfig
}

type RequestRun struct {
	// Full path to the binary
	Exe  string
	Args []string
	Env  []string
	Cwd  string
}

type RequestConfig struct {
	// nil indicates lack of presence
	Parallel *int
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
