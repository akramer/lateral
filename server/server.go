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
	// Number of process slots available for use
	slots int
	// When slots is incremented by 1, Signal cond
	// Otherwise, when slots is changed, Broadcast cond
	slotAvailable *sync.Cond
	// Broadcast every time a running task is finished.
	taskFinished *sync.Cond

	errorOccurred    bool
	shuttingDown     bool
	shutdownComplete bool

	pending  []*Request
	running  []*Request
	finished []finishedProcess
}

type finishedProcess struct {
	request *Request
	state   *os.ProcessState
}

var funcMap = map[RequestType]func(*instance, *Request) (*Response, error){
	REQUEST_GETPID:   (*instance).cmdGetpid,
	REQUEST_RUN:      (*instance).cmdRun,
	REQUEST_KILL:     (*instance).cmdKill,
	REQUEST_WAIT:     (*instance).cmdWait,
	REQUEST_SHUTDOWN: (*instance).cmdShutdown,
	REQUEST_CONFIG:   (*instance).cmdConfig,
}

func newInstance(v *viper.Viper) *instance {
	var i = instance{
		viper: v,
		slots: v.GetInt("start.parallel"),
	}
	i.slotAvailable = sync.NewCond(&i.m)
	i.taskFinished = sync.NewCond(&i.m)
	return &i
}

func (i *instance) broker() {
}

// Run the server's accept loop, waiting for connections from l.
// Correct shutdown procedure is:
// set slots to a number such that no new processes will run
// wait for all running processes to finish
// set shittingDown to true
// close the listening socket
func Run(v *viper.Viper, l *net.UnixListener) {
	i := newInstance(v)
	i.listener = l
	for {
		c, err := l.AcceptUnix()
		i.m.Lock()
		sdc := i.shutdownComplete
		i.m.Unlock()
		if sdc {
			glog.Infoln("Shutdown complete. closing listener.")
			if c != nil {
				c.Close()
			}
			l.Close()
			return
		} else if err != nil {
			glog.Errorln("Accept() failed on unix socket:", err)
			return
		}
		go i.connectionHandler(c)
	}
}

// Helper func, sends an error response to c.
func sendError(c *net.UnixConn, err error) {
	writeResponse(c, errorResponse(err))
}

func errorResponse(err error) *Response {
	return &Response{
		Type:    RESPONSE_ERR,
		Message: err.Error(),
	}
}

func (i *instance) connectionHandler(c *net.UnixConn) {
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

func (i *instance) cmdGetpid(req *Request) (*Response, error) {
	pid := os.Getpid()
	r := Response{
		Type:   RESPONSE_GETPID,
		Getpid: &ResponseGetpid{pid},
	}
	return &r, nil
}

// Delete target from r. Returns the new slice.
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

// Wait for a request slot to open, consume it, and move the request from the pending to the running queue.
// Consumes a slot.
func (i *instance) getRunSlot(req *Request) {
	i.m.Lock()
	defer i.m.Unlock()
	for i.slots <= 0 {
		i.slotAvailable.Wait()
	}
	i.slots--
	i.pending = del(i.pending, req)
	i.running = append(i.running, req)
}

// Remove request from the running queue and add the finished queue.
// Frees up a slot.
func (i *instance) putRunSlot(req *Request, ps *os.ProcessState) {
	i.m.Lock()
	defer i.m.Unlock()
	i.finished = append(i.finished, finishedProcess{
		request: req,
		state:   ps,
	})
	i.running = del(i.running, req)
	i.slots++
	i.slotAvailable.Signal()
	i.taskFinished.Broadcast()
}

func (i *instance) doRunInGoroutine(req *Request) {
	i.getRunSlot(req)
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
	p, err := os.StartProcess(req.Run.Exe, req.Run.Args, attr)
	for _, v := range attr.Files {
		if v != nil {
			v.Close()
		}
	}
	if err != nil {
		glog.Errorln("Error running command:", err)
		i.putRunSlot(req, nil)
		return
	}
	ps, err := p.Wait()
	i.putRunSlot(req, ps)
}

func (i *instance) cmdRun(req *Request) (*Response, error) {
	if req.Run == nil {
		return nil, fmt.Errorf("Missing RequestRun struct")
	}
	i.m.Lock()
	defer i.m.Unlock()
	if i.shuttingDown {
		return nil, fmt.Errorf("Cannot send requests to a shutting down server.")
	}
	i.pending = append(i.pending, req)
	go i.doRunInGoroutine(req)
	return &Response{Type: RESPONSE_OK}, nil
}

func (i *instance) cmdKill(req *Request) (*Response, error) {
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

func (i *instance) cmdWait(req *Request) (*Response, error) {
	i.m.Lock()
	for len(i.running) > 0 || len(i.pending) > 0 {
		i.taskFinished.Wait()
	}
	var exitStatus int
	for _, p := range i.finished {
		if p.state != nil && !p.state.Success() {
			exitStatus = 1
		}
	}
	if i.errorOccurred == true {
		exitStatus = 2
	}
	i.m.Unlock()
	resp := &Response{
		Type: RESPONSE_WAIT,
		Wait: &ResponseWait{
			ExitStatus: exitStatus,
		},
	}
	return resp, nil
}

func (i *instance) cmdConfig(req *Request) (*Response, error) {
	if req.Config == nil {
		return nil, fmt.Errorf("Missing RequestConfig struct")
	}
	i.m.Lock()
	defer i.m.Unlock()
	if req.Config.Parallel != nil {
		diff := i.viper.GetInt("start.parallel") - *req.Config.Parallel
		glog.Infof("Changing parallelism: adding %d slots", -diff)
		i.slots -= diff
		i.viper.Set("start.parallel", req.Config.Parallel)
		i.slotAvailable.Broadcast()
	}
	return &Response{Type: RESPONSE_OK}, nil
}

func (i *instance) cmdShutdown(req *Request) (*Response, error) {
	i.m.Lock()
	i.shuttingDown = true
	// Reduce concurrency to 0. If tasks are running, slots will go negative, but
	// will eventually be incremented to 0 once they're finished.
	i.slots -= i.viper.GetInt("start.parallel")
	for i.slots < 0 {
		i.taskFinished.Wait()
	}
	defer i.m.Unlock()
	i.shutdownComplete = true
	i.listener.Close()
	return &Response{Type: RESPONSE_OK}, nil
}
