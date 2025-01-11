/*
Copyright © 2025 YAUHEN SHULITSKI
*/
package cmd

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"
var Logger *log.Logger
var DebugLogger *log.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jsumo",
	Short: "jsumo is a tool to quickly forward your logs from journalctl to SumoLogic",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		Logger = log.New(os.Stdout, "", 0)

		// Extract the flags
		versionFlag, err := cmd.Flags().GetBool("version")
		if err != nil {
			return err
		}

		debugFlag, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		// Handle flags
		if debugFlag {
			DebugLogger = log.New(os.Stdout, "", 0)
		} else {
			DebugLogger = log.New(io.Discard, "", 0)
		}

		if versionFlag {
			Logger.Println(Version)
			return nil
		}

		// Get the receiver URL
		Logger.Printf("Initializing jsumo %s...\n", Version)
		_, err = GetReceiverURL()
		if err != nil {
			return err
		}
		Logger.Println("Initialization complete. Ready to forward journalctl logs to SumoLogic.")

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "print version and exit")
	rootCmd.Flags().BoolP("debug", "d", false, "enable debug mode")
}
