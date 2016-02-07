package server

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func makeTestInstance(v *viper.Viper) *instance {
	i := instance{
		viper: v,
	}
	return &i
}

func makeTestViper() *viper.Viper {
	v := viper.New()
	return v
}

func TestRunGetpid(t *testing.T) {
	i := makeTestInstance(makeTestViper())
	r := Request{Type: REQUEST_GETPID}
	resp, err := runGetpid(i, &r)
	if err != nil {
		t.Error("got error", err)
	} else if resp.Type == RESPONSE_ERR {
		t.Error("got error", resp.Message)
	} else if resp.Getpid.Pid != os.Getpid() {
		t.Error("Pid didn't match")
	}
}
