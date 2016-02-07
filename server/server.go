package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type instance struct {
	m     sync.Mutex
	viper *viper.Viper
}

// Open and return a listening unix socket.
func NewUnixListener(v *viper.Viper) (*net.UnixListener, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: v.GetString("socket")})
	if err != nil {
		return nil, err
	}

	return l, nil
}

func readMessage(c *net.UnixConn) (*Request, error) {
	// TODO: make this buffer configurable
	payload := make([]byte, 8192)
	oob := make([]byte, 8192)
	n, _, _, _, err := c.ReadMsgUnix(payload, oob)
	if err != nil && err != io.EOF {
		return nil, err
	}
	payload = payload[0:n]
	req := &Request{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func writeMessage(c *net.UnixConn, resp *Response) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	n, err := c.Write(payload)
	if err != nil {
		return err
	} else if n != len(payload) {
		return fmt.Errorf("Failed to write full payload, expected %v, wrote %v", len(payload), n)
	}

	c.CloseWrite()
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

func runConnection(i *instance, c *net.UnixConn) {
	defer c.Close()
	req, err := readMessage(c)
	if err != nil {
		glog.Errorln("Failed to read a message from socket:", err)
	}
	var resp *Response
	switch req.Type {
	case REQUEST_GETPID:
		resp, err = runGetpid(i, req)
	}
	if err != nil {
		writeMessage(c,
			&Response{
				Type:    RESPONSE_ERR,
				Message: err.Error(),
			})
	}
	err = writeMessage(c, resp)
	if err != nil {
		glog.Errorln("Failed to write a message to socket:", err)
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
