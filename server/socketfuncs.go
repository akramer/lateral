package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"syscall"

	"github.com/spf13/viper"
)

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
