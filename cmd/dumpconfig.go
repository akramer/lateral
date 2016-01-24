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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var dumpJson, dumpYaml, dumpToml bool

func runDumpconfigCmd(cmd *cobra.Command, args []string) {
	settings := Viper.AllSettings()
	var out []byte
	var err error
	if dumpJson {
		out, err = json.MarshalIndent(settings, "", "  ")
	} else if dumpToml {
		buffer := new(bytes.Buffer)
		enc := toml.NewEncoder(buffer)
		err = enc.Encode(settings)
		out = buffer.Bytes()
	} else /* yaml */ {
		out, err = yaml.Marshal(settings)
	}
	if err != nil {
		glog.Errorln("Failed to marshal the configuration: ", err)
		ExitCode = 1
		return
	}
	fmt.Println(string(out))
}

// dumpconfigCmd represents the dumpconfig command
var dumpconfigCmd = &cobra.Command{
	Use:   "dumpconfig",
	Short: "Dump available configuration options",
	Long: `Dump all configuration options in a format suitable for a configuration file.
Default, if unspecified, is YAML.`,
	Run: runDumpconfigCmd,
}

func init() {
	RootCmd.AddCommand(dumpconfigCmd)

	dumpconfigCmd.Flags().BoolVar(&dumpJson, "json", false, "dump config as json")
	dumpconfigCmd.Flags().BoolVar(&dumpYaml, "yaml", false, "dump config as yaml")
	dumpconfigCmd.Flags().BoolVar(&dumpToml, "toml", false, "dump config as toml")
}
