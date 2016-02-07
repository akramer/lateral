package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type instance struct {
	m     sync.Mutex
	viper *viper.Viper
}

var funcMap = map[RequestType]func(i *instance, req *Request) (*Response, error){
	REQUEST_GETPID: runGetpid,
	REQUEST_RUN:    runRun,
	REQUEST_KILL:   runKill,
}

// Open and return a listening unix socket.
func NewUnixListener(v *viper.Viper) (*net.UnixListener, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: v.GetString("socket")})
	if err != nil {
		return nil, err
	}

	return l, nil
}

func readRequest(c *net.UnixConn) (*Request, error) {
	var l uint32
	err := binary.Read(c, binary.BigEndian, &l)
	length := int(l)
	if err != nil {
		return nil, err
	}
	payload := make([]byte, length)
	n, err := c.Read(payload)
	if err != nil {
		return nil, err
	} else if n != length {
		return nil, fmt.Errorf("Payload was %d bytes rather than reported size of %d", n, length)
	}
	req := &Request{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return nil, err
	}
	if !req.HasFds {
		return req, nil
	}

	payload = make([]byte, 1)
	// TODO: does this buffer need to be configurable?
	oob := make([]byte, 8192)
	n, oobn, _, _, err := c.ReadMsgUnix(payload, oob)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("Error reading OOB filedescriptors")
	}
	oob = oob[0:oobn]
	scm, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, fmt.Errorf("Error parsing socket control message: %v", err)
	}
	var fds []int
	for i := 0; i < len(scm); i++ {
		tfds, err := syscall.ParseUnixRights(&scm[i])
		if err == syscall.EINVAL {
			continue // Wasn't a UnixRights Control Message
		} else if err != nil {
			return nil, fmt.Errorf("Error parsing unix rights: %v", err)
		}
		fds = append(fds, tfds...)
	}
	if len(fds) == 0 {
		return nil, fmt.Errorf("Failed to receive any FDs on a request with HasFds == true")
	}
	req.ReceivedFds = fds
	return req, nil
}

func writeResponse(c *net.UnixConn, resp *Response) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	err = binary.Write(c, binary.BigEndian, uint32(len(payload)))
	if err != nil {
		return err
	}
	n, err := c.Write(payload)
	if err != nil {
		return err
	} else if n != len(payload) {
		return fmt.Errorf("Failed to write full payload, expected %v, wrote %v", len(payload), n)
	}

	return nil
}

// Run the server's accept loop, waiting for connections from l.
func Run(v *viper.Viper, l *net.UnixListener) {
	var i = instance{viper: v}
	for {
		c, err := l.AcceptUnix()
		if err != nil {
			glog.Errorln("Accept() failed on unix socket:", err)
			return
		}
		go runConnection(&i, c)
	}
}

func sendError(c *net.UnixConn, err error) {
	writeResponse(c, &Response{
		Type:    RESPONSE_ERR,
		Message: err.Error(),
	})
}

func runConnection(i *instance, c *net.UnixConn) {
	defer c.Close()
	for {
		req, err := readRequest(c)
		if err == io.EOF {
			return // Client closed the connection.
		}
		if err != nil {
			glog.Errorln("Failed to read a message from socket:", err)
		}
		var resp *Response
		f, t := funcMap[req.Type]
		if t != true {
			sendError(c, fmt.Errorf("unknown request type"))
			continue
		}
		resp, err = f(i, req)
		if err != nil {
			sendError(c, err)
			continue
		}
		err = writeResponse(c, resp)
		if err != nil {
			glog.Errorln("Failed to write a message to socket:", err)
			return
		}
	}
}

func runGetpid(i *instance, req *Request) (*Response, error) {
	pid := os.Getpid()
	r := Response{
		Type:   RESPONSE_GETPID,
		Getpid: &ResponseGetpid{pid},
	}
	return &r, nil
}

func runRun(in *instance, req *Request) (*Response, error) {
	fmt.Printf("Received FDs! Original: %v, received: %v", req.Fds, req.ReceivedFds)
	if req.Run == nil {
		return nil, fmt.Errorf("Missing RequestRun struct")
	}
	var max int
	for _, v := range req.Fds {
		if v > max {
			max = v
		}
	}
	f := make([]*os.File, max+1)
	for i, v := range req.Fds {
		f[v] = os.NewFile(uintptr(req.ReceivedFds[i]), "fd")
	}
	attr := &os.ProcAttr{
		Env:   req.Run.Env,
		Dir:   req.Run.Cwd,
		Files: f,
	}
	p, err := os.StartProcess(req.Run.Args[0], req.Run.Args, attr)
	if err != nil {
		return nil, err
	}
	p.Wait()
	return &Response{Type: RESPONSE_ERR}, nil
}

func runKill(in *instance, req *Request) (*Response, error) {
	glog.Infoln("Server going down with SIGKILL")
	glog.Flush()

	pgid, err := syscall.Getpgid(0)
	if err != nil {
		glog.Errorln("Failed to get pgid", err)
		os.Exit(1)
	}

	err = syscall.Kill(-pgid, syscall.SIGKILL)
	if err != nil {
		glog.Errorln("Failed to kill our process group", err)
		os.Exit(1)
	}

	// We'll never get here...
	return nil, nil
}
