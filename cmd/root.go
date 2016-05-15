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
	"flag"
	"fmt"
	"os"

	"github.com/akramer/lateral/getsid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The main function exits with this code during a normal exit.
// Set this to the desired value before the Run func returns.
var ExitCode int

var cfgFile string

// The viper instance that will be passed to the implementation of lateral.
// Not using the global viper makes testing easier.
var Viper = viper.New()

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lateral <command>",
	Short: "lateral is an easy-to-use process parallelizer",
	Long: `Lateral is designed to make it a no-brainer to parallelize processing that would
otherwise be done sequentially. It's designed to be a more powerful 'xargs -P'
while also being low-friction and having as few surprises as possible.
`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.lateral/config.yaml)")
	RootCmd.PersistentFlags().StringP("socket", "s", "", "UNIX domain socket path (default $HOME/.lateral/socket.$SESSIONID)")
	Viper.BindPFlag("socket", RootCmd.PersistentFlags().Lookup("socket"))

	// glog flags
	RootCmd.PersistentFlags().Bool("logtostderr", false, "log to standard error instead of files")
	Viper.BindPFlag("logtostderr", RootCmd.PersistentFlags().Lookup("logtostderr"))
	RootCmd.PersistentFlags().Bool("alsologtostderr", false, "log to standard error as well as files")
	Viper.BindPFlag("alsologtostderr", RootCmd.PersistentFlags().Lookup("alsologtostderr"))
	RootCmd.PersistentFlags().String("stderrthreshold", "ERROR", "logs at or above this threshold go to stderr")
	Viper.BindPFlag("stderrthreshold", RootCmd.PersistentFlags().Lookup("stderrthreshold"))
	RootCmd.PersistentFlags().IntP("v", "v", 0, "log level for V logs")
	Viper.BindPFlag("v", RootCmd.PersistentFlags().Lookup("v"))
	RootCmd.PersistentFlags().String("vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	Viper.BindPFlag("vmodule", RootCmd.PersistentFlags().Lookup("vmodule"))
	RootCmd.PersistentFlags().String("log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")
	Viper.BindPFlag("log_backtrace_at", RootCmd.PersistentFlags().Lookup("log_backtrace_at"))
	RootCmd.PersistentFlags().String("log_dir", "", "If non-empty, write log files in this directory")
	Viper.BindPFlag("log_dir", RootCmd.PersistentFlags().Lookup("log_dir"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		Viper.SetConfigFile(cfgFile)
	}

	Viper.SetConfigName("config")         // name of config file (without extension)
	Viper.AddConfigPath("$HOME/.lateral") // adding home directory as first search path
	Viper.SetEnvPrefix("lateral")
	Viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	Viper.ReadInConfig()

	// glog uses the flag library, not pflags.
	// Parsing an empty argv suppresses a warning and allows pass-through of viper values.
	flag.CommandLine.Parse([]string{})
	flag.Set("logtostderr", fmt.Sprintf("%v", Viper.GetBool("logtostderr")))
	flag.Set("alsologtostderr", fmt.Sprintf("%v", Viper.GetBool("alsologtostderr")))
	flag.Set("stderrthreshold", fmt.Sprintf("%v", Viper.GetString("stderrthreshold")))
	flag.Set("v", fmt.Sprintf("%v", Viper.GetInt("v")))
	flag.Set("vmodule", fmt.Sprintf("%v", Viper.GetString("vmodule")))
	flag.Set("log_backtrace_at", fmt.Sprintf("%v", Viper.GetString("log_backtrace_at")))
	flag.Set("log_dir", fmt.Sprintf("%v", Viper.GetString("log_dir")))

	if Viper.GetString("socket") == "" {
		Viper.Set("socket", defaultSocketPath())
	}
}

func defaultSocketPath() string {
	home := os.Getenv("HOME")
	sid, err := getsid.Getsid(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error determining sid, using 0: %v", err)
	}
	name := home + "/.lateral/socket." + fmt.Sprintf("%d", sid)
	return name
}
