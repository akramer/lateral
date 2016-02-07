package client

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/akramer/lateral/server"
	"github.com/spf13/viper"
)

func NewUnixConn(v *viper.Viper) (*net.UnixConn, error) {
	c, err := net.DialUnix("unix", nil, &net.UnixAddr{Net: "unix", Name: v.GetString("socket")})
	return c, err
}

func SendRequest(c *net.UnixConn, req *server.Request) error {
	oob := make([]byte, 0)
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = binary.Write(c, binary.BigEndian, uint32(len(payload)))
	if err != nil {
		return err
	}
	n, oobn, err := c.WriteMsgUnix(payload, oob, nil)
	if err != nil {
		return err
	} else if n != len(payload) || oobn != len(oob) {
		return fmt.Errorf("Error writing to socket, expected n=%v got %v, oob=%v got %v", len(payload), n, len(oob), oobn)
	}
	return nil
}

func ReceiveResponse(c *net.UnixConn) (*server.Response, error) {
	var l uint32
	err := binary.Read(c, binary.BigEndian, &l)
	length := int(l)
	if err != nil {
		return nil, err
	}
	payload := make([]byte, length)
	n, err := c.Read(payload)
	if err != nil && err != io.EOF {
		return nil, err
	} else if n != length {
		return nil, fmt.Errorf("Read %d bytes and expected %d bytes", n, length)
	}
	var resp server.Response
	err = json.Unmarshal(payload, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
