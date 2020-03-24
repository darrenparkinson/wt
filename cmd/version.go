package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of wt",
	Long:  `All software has versions. This is wt's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WebEx Tool v0.1.0")
	},
}
