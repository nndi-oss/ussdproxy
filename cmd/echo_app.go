package cmd

import (
	"fmt"
	"os"

	"github.com/nndi-oss/ussdproxy/app/echo"
	"github.com/nndi-oss/ussdproxy/pkg/server"
	"github.com/spf13/cobra"
)

var echoAppCmd = &cobra.Command{
	Use:   "echo",
	Short: "Starts the echo server",
	Long:  `Starts the echo server`,
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewUssdProxyServer(config, echo.NewEchoApplication())
		addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
		s.ListenAndServe(addr)
		os.Exit(1)
	},
}
