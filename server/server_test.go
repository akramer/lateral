package server

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

type mockConn struct {
	mockRequest Request
	response    *Response
	closed      bool
}

func (m *mockConn) ReadMessage() (*Request, error) {
	return &m.mockRequest, nil
}

func (m *mockConn) WriteMessage(r *Response) error {
	m.response = r
	return nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func makeTestInstance() *instance {
	i := instance{
		viper: viper.New(),
	}
	return &i
}

func TestRunConnection(t *testing.T) {
	i := makeTestInstance()
	m := mockConn{
		mockRequest: Request{
			Type: REQUEST_GETPID,
		},
	}
	runConnection(i, &m)
	if m.response.Type != RESPONSE_GETPID {
		t.Error("Failed to get a response.")
	}
}

func TestRunGetpid(t *testing.T) {
	i := makeTestInstance()
	r := Request{Type: REQUEST_GETPID}
	resp, err := runGetpid(i, &r)
	if err != nil || resp.Type == RESPONSE_ERR {
		t.Error("got error", err, resp.Message)
	} else if resp.Getpid.Pid != os.Getpid() {
		t.Error("Pid didn't match")
	}
}
