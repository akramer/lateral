package server

import (
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type instance struct {
	m     sync.Mutex
	viper *viper.Viper
}

// Run the server's accept loop, waiting for connections from l.
func Run(v *viper.Viper, l Listener) {
	var i = instance{viper: v}
	for {
		c, err := l.Accept()
		if err != nil {
			glog.Errorln("Accept() failed on unix socket:", err)
			return
		}
		go runConnection(&i, c)
	}
}

func runConnection(i *instance, c Conn) {
	defer c.Close()
	req, err := c.ReadMessage()
	if err != nil {
		glog.Errorln("Failed to read a message from socket:", err)
	}
	var resp *Response
	switch req.Type {
	case REQUEST_GETPID:
		resp, err = runGetpid(i, req)
	}
	if err != nil {
		c.WriteMessage(
			&Response{
				Type:    RESPONSE_ERR,
				Message: err.Error(),
			})
	}
	err = c.WriteMessage(resp)
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
