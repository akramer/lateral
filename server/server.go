package server

import (
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
	viper    *viper.Viper
	listener *net.UnixListener

	// m protects the following members
	m sync.Mutex
	// Number of process slots available
	slots int
	cond  *sync.Cond
	// Has an error occurred at all?
	errorOccurred bool
	shuttingDown  bool
	pending       []*Request
	running       []*Request
	finished      []finishedProcess
}

type finishedProcess struct {
	request *Request
	state   *os.ProcessState
}

var funcMap = map[RequestType]func(*instance, *Request) (*Response, error){
	REQUEST_GETPID: (*instance).runGetpid,
	REQUEST_RUN:    (*instance).runRun,
	REQUEST_KILL:   (*instance).runKill,
	REQUEST_WAIT:   (*instance).runWait,
}

func newInstance(v *viper.Viper) *instance {
	var i = instance{
		viper: v,
		slots: v.GetInt("start.concurrency"),
	}
	i.cond = sync.NewCond(&i.m)
	return &i
}

func (i *instance) broker() {
}

// Run the server's accept loop, waiting for connections from l.
func Run(v *viper.Viper, l *net.UnixListener) {
	i := newInstance(v)
	i.listener = l
	for {
		c, err := l.AcceptUnix()
		i.m.Lock()
		if i.shuttingDown {
			i.m.Unlock()
			return
		}
		i.m.Unlock()
		if err != nil {
			glog.Errorln("Accept() failed on unix socket:", err)
			return
		}
		go i.runConnection(c)
	}
}

// Helper func, sends an error response to c.
func sendError(c *net.UnixConn, err error) {
	writeResponse(c, &Response{
		Type:    RESPONSE_ERR,
		Message: err.Error(),
	})
}

func (i *instance) runConnection(c *net.UnixConn) {
	defer c.Close()
	for {
		req, err := readRequest(c)
		if err == io.EOF {
			return // Client closed the connection.
		}
		if err != nil {
			glog.Errorln("Failed to read a message from socket:", err)
		}
		f, t := funcMap[req.Type]
		if t != true {
			sendError(c, fmt.Errorf("unknown request type"))
			continue
		}
		resp, err := f(i, req)
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

func (i *instance) runGetpid(req *Request) (*Response, error) {
	pid := os.Getpid()
	r := Response{
		Type:   RESPONSE_GETPID,
		Getpid: &ResponseGetpid{pid},
	}
	return &r, nil
}

func del(r []*Request, target *Request) []*Request {
	var i int
	for i = 0; r[i] != target && i < len(r); i++ {
	}
	if i == len(r) {
		return r
	}
	r, r[len(r)-1] = append(r[:i], r[i+1:]...), nil
	return r
}

func (i *instance) doRun(req *Request) {
	i.m.Lock()
	for i.slots <= 0 {
		i.cond.Wait()
	}
	i.slots--
	i.pending = del(i.pending, req)
	i.running = append(i.running, req)
	i.m.Unlock()
	var max int
	for _, v := range req.Fds {
		if v+1 > max {
			max = v + 1
		}
	}
	f := make([]*os.File, max)
	for i, v := range req.Fds {
		f[v] = os.NewFile(uintptr(req.ReceivedFds[i]), "fd")
	}
	attr := &os.ProcAttr{
		Env:   req.Run.Env,
		Dir:   req.Run.Cwd,
		Files: f,
	}
	// TODO: add running process to the running list
	p, err := os.StartProcess(req.Run.Args[0], req.Run.Args, attr)
	for _, v := range attr.Files {
		if v != nil {
			v.Close()
		}
	}
	if err != nil {
		glog.Errorln("Error running command:", err)
		i.m.Lock()
		defer i.m.Unlock()
		i.errorOccurred = true
		i.running = del(i.running, req)
		i.cond.Signal()
		i.slots++
		return
	}
	ps, err := p.Wait()
	i.m.Lock()
	defer i.m.Unlock()
	i.finished = append(i.finished, finishedProcess{
		request: req,
		state:   ps,
	})
	i.running = del(i.running, req)
	i.slots++
	i.cond.Signal()
}

func (i *instance) runRun(req *Request) (*Response, error) {
	if req.Run == nil {
		return nil, fmt.Errorf("Missing RequestRun struct")
	}
	i.m.Lock()
	defer i.m.Unlock()
	i.pending = append(i.pending, req)
	go i.doRun(req)
	return &Response{Type: RESPONSE_OK}, nil
}

func (i *instance) runKill(req *Request) (*Response, error) {
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

func (i *instance) runWait(req *Request) (*Response, error) {
	i.m.Lock()
	for len(i.running) > 0 || len(i.pending) > 0 {
		i.cond.Wait()
	}
	defer i.m.Unlock()
	var exitStatus int
	for _, p := range i.finished {
		if !p.state.Success() {
			exitStatus = 1
		}
	}
	resp := &Response{
		Type: RESPONSE_WAIT,
		Wait: &ResponseWait{
			ExitStatus: exitStatus,
		},
	}
	return resp, nil
}
