package cmd

import (
	"fmt"

	"github.com/delving/hub3/ikuzo"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version-information of Ikuzo",
	Run: func(cmd *cobra.Command, args []string) {
		info := ikuzo.NewBuildVersionInfo(version, gitHash, buildAgent, buildStamp)
		fmt.Printf("%s\n", info)
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}
