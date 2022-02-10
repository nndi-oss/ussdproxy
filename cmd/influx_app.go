package cmd

import (
	"log"

	"github.com/nndi-oss/ussdproxy/app/influx"
	"github.com/nndi-oss/ussdproxy/server"
	"github.com/spf13/cobra"
)

var influxAppCmd = &cobra.Command{
	Use:   "influx-proxy",
	Short: "Starts the influx proxy server",
	Long:  `Starts the influx proxy server`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := "localhost:3000"
		influxApplication := influx.NewInfluxApp(addr, "test_database", "user", "password")
		if err := server.ListenAndServe(addr, influxApplication); err != nil {
			log.Fatalf("Failed to start Influx Application. Error %s", err)
		}
	},
}
