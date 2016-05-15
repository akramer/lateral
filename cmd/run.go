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
	"fmt"
	"os"
	"os/exec"

	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the given command in the lateral server",
	Long:  `A longer description that spans multiple lines and likely contains examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			panic(fmt.Errorf("No command specified"))
		}
		c, err := client.NewUnixConn(Viper)
		if err != nil {
			panic(fmt.Errorf("Error connecting to server: %v", err))
		}
		defer c.Close()
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("Error determining working directory"))
		}
		exe, err := exec.LookPath(args[0])
		if err != nil {
			panic(fmt.Errorf("Failed to find executable %v", args[0]))
		}
		req := &server.Request{
			Type:   server.REQUEST_RUN,
			HasFds: true,
			// TODO: send all fds over the socket.
			Fds: []int{0, 1, 2},
			Run: &server.RequestRun{
				Exe:  exe,
				Args: args,
				Env:  os.Environ(),
				Cwd:  wd,
			},
		}
		err = client.SendRequest(c, req)
		if err != nil {
			panic(fmt.Errorf("Error sending request: %v", err))
		}
		resp, err := client.ReceiveResponse(c)
		if err != nil {
			panic(fmt.Errorf("Error receiving response: %v", err))
		}
		if resp.Type != server.RESPONSE_OK {
			panic(fmt.Errorf("Error in server response: %v", resp.Message))
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

}
