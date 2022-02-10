package cmd

import (
	"log"

	"github.com/nndi-oss/ussdproxy/app/echo"
	"github.com/nndi-oss/ussdproxy/server"
	"github.com/spf13/cobra"
)

var echoAppCmd = &cobra.Command{
	Use:   "echo",
	Short: "Starts the echo server",
	Long:  `Starts the echo server`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := "localhost:3000"
		echoAplication := echo.NewEchoApplication()

		if err := server.ListenAndServe(addr, echoAplication); err != nil {
			log.Fatalf("Failed to start Echo Application. Error %s", err)
		}
	},
}
