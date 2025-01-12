/*
Copyright © 2025 YAUHEN SHULITSKI
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version            = "dev"
	Logger             *log.Logger
	DebugLogger        *log.Logger
	UploadQueue        Queue
	FlagVersion        bool
	FlagDebug          bool
	FlagReceiver       string
	FlagReadInterval   time.Duration
	FlagUploadInterval time.Duration
	FlagSourceCategory string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jsumo",
	Short: "jsumo is a tool to quickly forward your logs from journalctl to SumoLogic",
	Long:  `jsumo is a tool to quickly forward your logs from journalctl to SumoLogic. It uses journalctl cursor to ensure that no logs are lost.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		Logger = log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)
		UploadQueue = Queue{}

		// Handle flags
		if FlagDebug {
			DebugLogger = log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)
		} else {
			DebugLogger = log.New(io.Discard, "", 0)
		}

		if FlagVersion {
			fmt.Println(Version)
			return nil
		}

		// Get the receiver URL
		if FlagReceiver == "" {
			Logger.Printf("Initializing jsumo %s...\n", Version)
			receiverURL, err := GetReceiverURL()
			if err != nil {
				return err
			}
			FlagReceiver = receiverURL
		}
		if FlagReceiver == "" {
			return fmt.Errorf("receiver URL is empty")
		}
		Logger.Printf("Initialization complete. Ready to forward journalctl logs to %s\n", FlagReceiver)

		journalReader, err := NewJournalReader()
		if err != nil {
			return err
		}

		// Start reading logs from journalctl every 5 seconds
		tickerJournal := time.NewTicker(FlagReadInterval)
		logReadIsActive := false
		go func() {
			for ; ; <-tickerJournal.C {
				logReadIsActive = true
				err := journalReader.ReadLogs()
				if err != nil {
					Logger.Println(red(err))
				}
				logReadIsActive = false
			}
		}()

		// Start uploading files to SumoLogic
		tickerUploader := time.NewTicker(FlagUploadInterval)
		uploaderIsActive := false
		go func() {
			for ; ; <-tickerUploader.C {
				uploaderIsActive = true
				fileToUpload := UploadQueue.Next()
				if fileToUpload != "" {
					err := uploadFileToSumoSource(fileToUpload, FlagReceiver)
					if err != nil {
						Logger.Println(red(err))
						UploadQueue.ReturnFile(fileToUpload)
						continue
					}
				} else {
					DebugLogger.Println("No files to upload")
				}
				uploaderIsActive = false
			}
		}()

		// Handle graceful shutdown on Ctrl+C or SIGINT signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		Logger.Println(yellow("Shutting down gracefully..."))
		tickerJournal.Stop()
		tickerUploader.Stop()
		for logReadIsActive {
			Logger.Println(yellow("Waiting for log reading to finish..."))
			time.Sleep(1 * time.Second)
		}
		for uploaderIsActive {
			Logger.Println(yellow("Waiting for file upload to finish..."))
			time.Sleep(1 * time.Second)
		}
		Logger.Println("Shutdown complete.")
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
	rootCmd.PersistentFlags().BoolVarP(&FlagVersion, "version", "v", false, "print version and exit")
	rootCmd.PersistentFlags().BoolVarP(&FlagDebug, "debug", "d", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&FlagReceiver, "url", "r", "", "receiver URL. If empty, it will be fetched or created automatically using SumoLogic API")
	rootCmd.PersistentFlags().DurationVar(&FlagReadInterval, "read-interval", 5*time.Second, "interval to read logs from journalctl")
	rootCmd.PersistentFlags().DurationVar(&FlagUploadInterval, "upload-interval", 2*time.Second, "interval to upload files to the receiver URL")
	rootCmd.PersistentFlags().StringVarP(&FlagSourceCategory, "category", "c", "", "override source category with the given value")
}
