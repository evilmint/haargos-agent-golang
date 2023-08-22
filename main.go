package main

import (
	"fmt"
	"os"

	"github.com/evilmint/haargos-agent-golang/haargos"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var log = logrus.New()

func main() {
	var rootCmd = &cobra.Command{Use: "haargos"}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Current agent version: %s\n", version)
		},
	}

	var cmdHelp = &cobra.Command{
		Use:   "help",
		Short: "Print basic help information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(`Usage of this CLI:
  help      Print basic help information
  version   Print the current agent version`)
		},
	}

	haargosClient := &haargos.Haargos{}
	var haConfigPath string
	var z2mPath string
	var zhaPath string
	var installationId string
	var userId string
	var token string

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run Haargos",
		Run: func(cmd *cobra.Command, args []string) {
			if haConfigPath == "" {
				log.Error("The --ha-config flag must be provided")
				os.Exit(1)
			} else if installationId == "" {
				log.Error("The --installation-id flag must be provided.")
				os.Exit(1)
			} else if userId == "" {
				log.Error("The --user-id flag must be provided.")
				os.Exit(1)
			} else if token == "" {
				log.Error("The --token flag must be provided.")
				os.Exit(1)
			}

			haargosClient.Run(
				haargos.RunParams{
					UserID:         userId,
					InstallationID: installationId,
					Token:          token,
					HaConfigPath:   haConfigPath,
					Z2MPath:        z2mPath,
					ZHAPath:        zhaPath,
				},
			)
		},
	}

	cmdRun.Flags().
		StringVarP(&haConfigPath, "ha-config", "c", "", "Path to the Home Assistant configuration")

	cmdRun.Flags().StringVarP(&z2mPath, "z2m-path", "z", "", "Path to Z2M database")
	cmdRun.Flags().StringVarP(&zhaPath, "zha-path", "x", "", "Path to ZHA database")
	cmdRun.Flags().StringVarP(&installationId, "installation-id", "i", "", "Installation ID")
	cmdRun.Flags().StringVarP(&userId, "user-id", "u", "", "User ID")
	cmdRun.Flags().StringVarP(&token, "token", "t", "", "Token")
	rootCmd.AddCommand(cmdVersion, cmdHelp, cmdRun)
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error sending request request: %v", err)
		os.Exit(1)
	}
}
