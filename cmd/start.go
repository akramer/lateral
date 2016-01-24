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
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const MAGICENV = "LAT_MAGIC"

func runStart(cmd *cobra.Command, args []string) {
	// If MAGICENV is set to the socket path, we can be (relatively) sure we're the child process.
	if Viper.GetBool("start.foreground") || os.Getenv(MAGICENV) == Viper.GetString("socket") {
		glog.Infoln("Not forking a child server")
	} else {
		glog.Infoln("forking child...")
		os.Setenv(MAGICENV, Viper.GetString("socket"))
		attr := &syscall.ProcAttr{
			Dir:   "/",
			Env:   os.Environ(),
			Files: []uintptr{0, 1, 2}}
		pid, err := syscall.ForkExec("/proc/self/exe", os.Args, attr)
		if err != nil {
			glog.Errorln("Error forking subprocess: ", err, pid)
		}
	}
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the lateral background server",
	Long: `Start the lateral background server. By default, this creates a new server
for every session. This essentially means each login shell will have its own
server.`,
	Run: runStart,
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("new_server", "n", false, "Print an error and return a non-zero status if the server is already running")
	Viper.BindPFlag("start.new_server", startCmd.Flags().Lookup("new_server"))
	startCmd.Flags().BoolP("foreground", "f", false, "Do not fork off a background server: run in the foreground.")
	Viper.BindPFlag("start.foreground", startCmd.Flags().Lookup("foreground"))
}
