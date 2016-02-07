package server

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func makeTestInstance(v *viper.Viper) *instance {
	return newInstance(v)
}

func makeTestViper() *viper.Viper {
	v := viper.New()
	v.Set("start.concurrency", 10)
	return v
}

func TestRunGetpid(t *testing.T) {
	i := makeTestInstance(makeTestViper())
	r := Request{Type: REQUEST_GETPID}
	resp, err := i.runGetpid(&r)
	if err != nil {
		t.Error("got error", err)
	} else if resp.Type == RESPONSE_ERR {
		t.Error("got error", resp.Message)
	} else if resp.Getpid.Pid != os.Getpid() {
		t.Error("Pid didn't match")
	}
}

func TestRunWait(t *testing.T) {
	i := makeTestInstance(makeTestViper())
	r := Request{Type: REQUEST_WAIT}
	resp, err := i.runWait(&r)
	if err != nil {
		t.Error("got error", err)
	} else if resp.Type == RESPONSE_ERR {
		t.Error("got error", resp.Message)
	} else if resp.Wait.ExitStatus != 0 {
		t.Error("Exit status wasn't 0")
	}

	runCmd := &Request{
		Type: REQUEST_RUN,
		Run: &RequestRun{
			Args: []string{"/bin/false"},
			Env:  os.Environ(),
		},
	}
	_, err = i.runRun(runCmd)
	if err != nil {
		t.Error("got error", err)
	} else if resp.Type == RESPONSE_ERR {
		t.Error("got error", resp.Message)
	}

	resp, err = i.runWait(&r)
	if err != nil {
		t.Error("got error", err)
	} else if resp.Type == RESPONSE_ERR {
		t.Error("got error", resp.Message)
	} else if resp.Wait.ExitStatus != 1 {
		t.Error("Exit status wasn't 1")
	}
}
