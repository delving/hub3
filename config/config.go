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

package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Config is the general configuration object
	Config RawConfig

	// CfgFile is the path to the config file
	CfgFile string
)

// RawConfig holds all the configuration blocks.
// These are bound from cli, Environment variables or configuration files by
// Viper.
type RawConfig struct {
	OrgID         string `json:"orgId"`
	Port          int    `json:"port" mapstructure:"port"`
	ElasticSearch `json:"elasticsearch"`
}

// ElasticSearch holds all the configuration values
// It is bound by Viper.
type ElasticSearch struct {
	Urls []string `json:"urls"`
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
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
	viper.SetDefault("ElasticSearch.urls", []string{"http://localhost:9200"})

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
