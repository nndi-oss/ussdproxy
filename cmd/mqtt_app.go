package cmd

import (
	"fmt"
	"os"

	"github.com/nndi-oss/ussdproxy/app/mqtt"
	"github.com/nndi-oss/ussdproxy/pkg/server"
	"github.com/spf13/cobra"
)

var mqttAppCmd = &cobra.Command{
	Use:   "mqtt-proxy",
	Short: "Starts the mqtt proxy server",
	Long:  `Starts the mqtt proxy server`,
	Run: func(cmd *cobra.Command, args []string) {
		mqttApplication := mqtt.NewMQTTApplication("tcp://localhost:1883", "user", "password", "some/topic")

		s := server.NewUssdProxyServer(config, mqttApplication)
		addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
		s.ListenAndServe(addr)
		os.Exit(1)
	},
}
