package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-homedir"
	ussdproxyconfig "github.com/nndi-oss/ussdproxy/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var config = &ussdproxyconfig.UssdProxyConfig{}
var logger hclog.Logger

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/ussdproxy.yml)")
	rootCmd.PersistentFlags().Bool("vvv", true, "Verbose output")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(echoAppCmd)
	rootCmd.AddCommand(influxAppCmd)
	rootCmd.AddCommand(mqttAppCmd)
	rootCmd.AddCommand(adminCmd)
}

func initConfig() {
	var home string
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetConfigName("ussdproxy")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/ussdproxy")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	if err := viper.Unmarshal(config); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	logpath := path.Join(home, "logs", "ussdproxy.log")
	if config.Logging.Path != "" {
		logpath = path.Join(config.Logging.Path, "ussdproxy.log")
	}

	logger = hclog.New(&hclog.LoggerOptions{
		Name:  logpath,
		Level: hclog.LevelFromString(config.Logging.Level),
	})
}

var rootCmd = &cobra.Command{
	Use:   "ussdproxy",
	Short: "UssdProxy",
	Long:  `More info: https://github.com/nndi-oss/ussdproxy`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ussdproxyconfig.Validate(*config); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}
