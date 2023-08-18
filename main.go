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
	var rootCmd = &cobra.Command{Use: "myapp"}

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

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run Haargos",
		Run: func(cmd *cobra.Command, args []string) {
			haargosClient.Run(
				haargos.RunParams{
					UserID:         "07957eee-0d3d-4e09-8d25-465bb1a82806",
					InstallationID: "f2687b3e-d6f7-4cbd-a58b-48000752c2a9",
					Token:          "ba4d8180-88b1-4645-9d0b-d4980a86be05",
					HaConfigPath:   haConfigPath,
				},
			)
		},
	}

	cmdRun.Flags().StringVarP(&haConfigPath, "ha-config", "c", "", "Path to the Home Assistant configuration")
	rootCmd.AddCommand(cmdVersion, cmdHelp, cmdRun)
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error sending request request: %v", err)
		os.Exit(1)
	}
}
