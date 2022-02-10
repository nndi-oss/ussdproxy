package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Perform administrative tasks",
	Long:  `Perform administrative tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(1)
	},
}
