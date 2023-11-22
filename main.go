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
	rootCmd := &cobra.Command{Use: "haargos"}

	rootCmd.AddCommand(createVersionCommand())
	rootCmd.AddCommand(createHelpCommand())
	rootCmd.AddCommand(createRunCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}

// func main() {
// 	gatherer := dockergatherer.NewDockerGatherer("/var/run/docker.sock")

// 	data := gatherer.GatherDocker()

// 	fmt.Print(data)
// }

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
	var haConfigPath, z2mPath, zhaPath, agentToken string

	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Run Haargos",
		Run: func(cmd *cobra.Command, args []string) {
			if haConfigPath == "" || agentToken == "" {
				logMissingFlags(haConfigPath, agentToken)
				os.Exit(1)
			}

			haargosClient := haargos.NewHaargos()
			haargosClient.Run(
				haargos.RunParams{
					AgentToken:   agentToken,
					HaConfigPath: haConfigPath,
					Z2MPath:      z2mPath,
					ZHAPath:      zhaPath,
				},
			)
		},
	}

	cmdRun.Flags().StringVarP(&haConfigPath, "ha-config", "c", "", "Path to the Home Assistant configuration")
	cmdRun.Flags().StringVarP(&z2mPath, "z2m-path", "z", "", "Path to Z2M database")
	cmdRun.Flags().StringVarP(&zhaPath, "zha-path", "x", "", "Path to ZHA database")
	cmdRun.Flags().StringVarP(&agentToken, "agent-token", "", "", "Agent Token")

	return cmdRun
}

func logMissingFlags(haConfigPath, agentToken string) {
	if haConfigPath == "" {
		log.Error("The --ha-config flag must be provided")
	}
	if agentToken == "" {
		log.Error("The --agent-token flag must be provided.")
	}
}
