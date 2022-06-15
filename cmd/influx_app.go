package cmd

import (
	"fmt"
	"os"

	"github.com/nndi-oss/ussdproxy/app/influx"
	"github.com/nndi-oss/ussdproxy/pkg/server"
	"github.com/spf13/cobra"
)

var influxAppCmd = &cobra.Command{
	Use:   "influx-proxy",
	Short: "Starts the influx proxy server",
	Long:  `Starts the influx proxy server`,
	Run: func(cmd *cobra.Command, args []string) {
		influxApplication := influx.NewInfluxApp("127.0.0.1:9009", "ussdproxy", "user", "password")
		s := server.NewUssdProxyServer(config, influxApplication)
		addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
		s.ListenAndServe(addr)
		os.Exit(1)
	},
}
