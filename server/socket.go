package server

import (
	"net"

	"github.com/spf13/viper"
)

// Implements the server.Listener interface
type UnixListener struct {
	*net.UnixListener
}

// Open and return a listening unix socket.
func NewUnixListener(v *viper.Viper) (Listener, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: v.GetString("socket")})
	if err != nil {
		return nil, err
	}

	return &UnixListener{l}, nil
}

func (l *UnixListener) Accept() (Conn, error) {
	c, err := l.UnixListener.AcceptUnix()
	if err != nil {
		return nil, err
	}
	return &UnixConn{c}, nil
}

// Implements the server.Conn interface
type UnixConn struct {
	*net.UnixConn
}

func (*UnixConn) ReadMessage() (*Request, error) {

	return nil, nil
}

func (*UnixConn) WriteMessage(*Response) error {
	return nil
}

func (c *UnixConn) Close() error {
	return c.UnixConn.Close()
}
