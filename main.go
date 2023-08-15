package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

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

	rootCmd.AddCommand(cmdVersion, cmdHelp)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
