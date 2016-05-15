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
	"github.com/spf13/cobra"
)

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for all currently inserted tasks to finish",
	Long: `Wait for all currently inserted tasks to finish.
Returns 0 if all tasks exited with success, otherwise returns 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewUnixConn(Viper)
		if err != nil {
			panic(fmt.Errorf("Error connecting to server: %v", err))
		}
		defer c.Close()
		req := &server.Request{
			Type: server.REQUEST_WAIT,
		}
		err = client.SendRequest(c, req)
		if err != nil {
			panic(fmt.Errorf("Error sending request: %v", err))
		}
		resp, err := client.ReceiveResponse(c)
		if err != nil {
			panic(fmt.Errorf("Error receiving response: %v", err))
		}
		if resp.Type != server.RESPONSE_WAIT {
			panic(fmt.Errorf("Error in server response: %v", resp.Message))
		}
		ExitCode = resp.Wait.ExitStatus

		if Viper.GetBool("wait.no_shutdown") {
			return
		}

		req = &server.Request{
			Type: server.REQUEST_SHUTDOWN,
		}
		err = client.SendRequest(c, req)
		if err != nil {
			panic(fmt.Errorf("Error sending request: %v", err))
		}
		resp, err = client.ReceiveResponse(c)
		if err != nil {
			panic(fmt.Errorf("Error receiving response: %v", err))
		}
		if resp.Type != server.RESPONSE_OK {
			panic(fmt.Errorf("Error in server response: %v", resp.Message))
		}

	},
}

func init() {
	RootCmd.AddCommand(waitCmd)
	waitCmd.Flags().BoolP("no_shutdown", "n", false, "Do not shut down server after wait is complete")
	Viper.BindPFlag("wait.no_shutdown", waitCmd.Flags().Lookup("no_shutdown"))
}
