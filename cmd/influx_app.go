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
		influxApplication := influx.NewInfluxApp("127.0.0.1:9009", "ussdproxy", "user", "password")
		if err := server.ListenAndServe("localhost:3000", influxApplication); err != nil {
			log.Fatalf("Failed to start Influx Application. Error %s", err)
		}
	},
}
