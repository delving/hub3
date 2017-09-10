// Copyright © 2017 Delving B.V. <info@delving.eu>
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
	"log"
	"os"
	"strings"

	"bitbucket.org/delving/rapid/config"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// path to the config file
	cfgFile string
	// Version of the application. (Injected at build time)
	Version = "0.1-SNAPSHOT"
	// GoVersion is the Golang version of the application. (Injected at build time)
	GoVersion = ""
	// BuildStamp is the timestamp of the application. (Injected at build time)
	BuildStamp = "1970-01-01 UTC"
	// BuildAgent is the agent that created the current build. (Injected at build time)
	BuildAgent string
	// GitHash of the current build. (Injected at build time.)
	GitHash string

	// General configuration object
	Config config.RawConfig
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rapid",
	Short: "RAPID: A Golang-based Linked Open Data discovery platform.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string, goversion string, buildstamp string, buildagent string, githash string) {
	Version = version
	GoVersion = goversion
	BuildStamp = buildstamp
	BuildAgent = buildagent
	GitHash = githash

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rapid.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".rapid" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".rapid")
	}

	viper.SetEnvPrefix("RAPID")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// setting defaults
	viper.SetDefault("port", 3001)
	viper.SetDefault("orgId", "rapid")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatal(
			fmt.Sprintf("unable to decode into struct, %v", err),
		)
	}

}
