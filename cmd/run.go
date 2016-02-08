// Copyright © 2016 Adam Kramer <akramer@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"os/exec"

	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the given command in the lateral server",
	Long:  `A longer description that spans multiple lines and likely contains examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			glog.Errorln("No command specified")
			ExitCode = 1
			return
		}
		c, err := client.NewUnixConn(Viper)
		if err != nil {
			glog.Errorln("Error connecting to server:", err)
			ExitCode = 1
			return
		}
		defer c.Close()
		wd, err := os.Getwd()
		if err != nil {
			glog.Errorln("Error determining working directory")
			ExitCode = 1
			return
		}
		exe, err := exec.LookPath(args[0])
		if err != nil {
			glog.Errorln("Failed to find executable", args[0])
			ExitCode = 1
			return
		}
		req := &server.Request{
			Type:   server.REQUEST_RUN,
			HasFds: true,
			Fds:    []int{0, 1, 2},
			Run: &server.RequestRun{
				Exe:  exe,
				Args: args,
				Env:  os.Environ(),
				Cwd:  wd,
			},
		}
		err = client.SendRequest(c, req)
		if err != nil {
			glog.Errorln("Error sending request:", err)
			ExitCode = 1
			return
		}
		resp, err := client.ReceiveResponse(c)
		if err != nil {
			glog.Errorln("Error receiving response:", err)
			ExitCode = 1
			return
		}
		if resp.Type != server.RESPONSE_OK {
			glog.Errorln("Error in server response:", resp.Message)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

}
