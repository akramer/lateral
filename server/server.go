package server

import (
	"encoding/binary"
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
	n, _, _, _, err = c.ReadMsgUnix(payload, oob)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("Error reading OOB filedescriptors")
	}
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
		switch req.Type {
		case REQUEST_GETPID:
			resp, err = runGetpid(i, req)
		}
		if err != nil {
			writeResponse(c,
				&Response{
					Type:    RESPONSE_ERR,
					Message: err.Error(),
				})
		}
		err = writeResponse(c, resp)
		if err != nil {
			glog.Errorln("Failed to write a message to socket:", err)
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
