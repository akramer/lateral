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

	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// getpidCmd represents the getpid command
var getpidCmd = &cobra.Command{
	Use:   "getpid",
	Short: "Print pid of server to stdout",
	Long:  "Print pid of server to stdout.",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewUnixConn(Viper)
		if err != nil {
			glog.Errorln("Error connecting to server:", err)
			ExitCode = 1
			return
		}
		defer c.Close()
		req := &server.Request{
			Type: server.REQUEST_GETPID,
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
		if resp.Type != server.RESPONSE_GETPID {
			glog.Errorln("Error in server response:", resp.Message)
		}
		fmt.Printf("%d\n", resp.Getpid.Pid)
	},
}

func init() {
	RootCmd.AddCommand(getpidCmd)
}
