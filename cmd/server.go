package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nndi-oss/ussdproxy/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the server",
	Long:  `Starts the server`,
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewUssdProxyServer()
		err := godotenv.Load()
		if err != nil {
			log.Println("Failed to load .env")
			os.Exit(1)
			return
		}
		addr := "localhost:3000"
		s.ListenAndServe(addr)
		os.Exit(1)
	},
}
