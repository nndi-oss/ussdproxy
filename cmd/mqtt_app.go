package cmd

import (
	"log"

	"github.com/nndi-oss/ussdproxy/app/mqtt"
	"github.com/nndi-oss/ussdproxy/server"
	"github.com/spf13/cobra"
)

var mqttAppCmd = &cobra.Command{
	Use:   "mqtt-proxy",
	Short: "Starts the mqtt proxy server",
	Long:  `Starts the mqtt proxy server`,
	Run: func(cmd *cobra.Command, args []string) {
		mqttApplication := mqtt.NewMQTTApplication("tcp://localhost:1883", "user", "password", "some/topic")
		if err := server.ListenAndServe("localhost:3000", mqttApplication); err != nil {
			log.Fatalf("Failed to start MQTT Application. Error %s", err)
		}
	},
}
