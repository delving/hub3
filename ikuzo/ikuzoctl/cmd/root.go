// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint:gochecknoinits

// Package cmd is the commandline-interface for the Ikuzo library.
package cmd

import (
	"fmt"
	"os"
	"strings"

	hub3Cfg "github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/ikuzoctl/cmd/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// version of the application. (Injected at build time)
	version = ""
	// buildStamp is the timestamp of the application. (Injected at build time)
	buildStamp = "1970-01-01 UTC"
	// buildAgent is the agent that created the current build. (Injected at build time)
	buildAgent string
	// gitHash of the current build. (Injected at build time.)
	gitHash string
)

var (
	cfgFile string
	cfg     config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ikuzo",
	Short: "Ikuzo is a high performance Linked Open Data Publication platform",
	Long: `Ikuzo can be run in Silo (all services under a single command) or
	as a series of microservices where all the components can be started as
	their own service running from the same binary and/or Docker image.
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is hub3.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
			log.Fatal().Err(err).Msg("unable to find home-dir")
			// os.Exit(1)
		}

		// Search config in home directory with name ".ikuzo" (without extension).
		viper.AddConfigPath("/etc/default")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName("hub3")
	}

	viper.SetEnvPrefix("HUB3")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// set default config values
	config.SetViperDefaults()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Str("configPath", viper.ConfigFileUsed()).Msg("unable to read configuration file")

		switch err.(type) {
		case viper.ConfigParseError:
			log.Fatal().Err(err).Str("configPath", viper.ConfigFileUsed()).Msg("unable to read configuration file")
		default:
			log.Warn().Err(err).Str("configPath", viper.ConfigFileUsed()).Msg("unable to read configuration file")
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal().Err(err).Msg("unable to decode configuration into struct")
	}

	// TODO(kiivihal): remove this with next release
	hub3Cfg.InitConfig()
}
