package main

import (
	"fmt"
	"os"

	"github.com/evilmint/haargos-agent-golang/haargos"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "1.0.0"

var logger = logrus.New()

func main() {
	debugEnabled := os.Getenv("DEBUG") == "true"

	if debugEnabled {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Level = logrus.InfoLevel
	}

	rootCmd := &cobra.Command{Use: "haargos"}

	rootCmd.AddCommand(createVersionCommand())
	rootCmd.AddCommand(createHelpCommand())
	rootCmd.AddCommand(createRunCommand())

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Error executing command: %v", err)
	}
}

func createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Current agent version: %s\n", version)
		},
	}
}

func createHelpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Print basic help information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(`Usage of this CLI:
  help      Print basic help information
  version   Print the current agent version`)
		},
	}
}

func createRunCommand() *cobra.Command {
	var haConfigPath, z2mPath, zhaPath, agentType string
	agentToken := os.Getenv("HAARGOS_AGENT_TOKEN")
	var stage = os.Getenv("STAGE")

	if stage == "" {
		stage = "production"
	} else if stage != "production" && stage != "dev" {
		logger.Fatal("The stage env must be production or dev.")
	}

	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Run Haargos",
		Run: func(cmd *cobra.Command, args []string) {
			if haConfigPath == "" || agentToken == "" {
				logMissingFlags(haConfigPath, agentToken)
			}

			debugEnabled := os.Getenv("DEBUG") == "true"
			haargosClient := haargos.NewHaargos(logger, debugEnabled)
			haargosClient.Run(
				haargos.RunParams{
					AgentToken:   agentToken,
					AgentType:    agentType,
					HaConfigPath: haConfigPath,
					Z2MPath:      z2mPath,
					ZHAPath:      zhaPath,
					Stage:        stage,
				},
			)
		},
	}

	cmdRun.Flags().StringVarP(&haConfigPath, "ha-config", "c", "", "Path to the Home Assistant configuration")
	cmdRun.Flags().StringVarP(&z2mPath, "z2m-path", "z", "", "Path to Z2M database")
	cmdRun.Flags().StringVarP(&zhaPath, "zha-path", "x", "", "Path to ZHA database")
	cmdRun.Flags().StringVarP(&agentType, "agent-type", "t", "bin", "Agent type")

	return cmdRun
}

func logMissingFlags(haConfigPath, agentToken string) {
	if haConfigPath == "" {
		logger.Fatal("The --ha-config flag must be provided")
	}
	if agentToken == "" {
		logger.Fatal("The agent-token environment variable must be provided.")
	}
}
