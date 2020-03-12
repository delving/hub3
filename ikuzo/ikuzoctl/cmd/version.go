package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version-information of Ikuzo",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(kiivihal): add proper versioning from build injection

		fmt.Println("Ikuzo v0.1.0; BuildDate: unknown")
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}
