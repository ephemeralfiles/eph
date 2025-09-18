package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/ephcli"
	"github.com/spf13/cobra"
)

const (
	// GithubRepository is the GitHub repository identifier for self-updates.
	GithubRepository = "ephemeralfiles/eph"
	// DefaultEndpoint is the default API endpoint for ephemeralfiles.
	DefaultEndpoint  = "https://api.ephemeralfiles.com"
)

var (
	noProgressBar     bool
	debugMode         bool
	configurationFile string
	token             string
	endpoint          string

	fileToUpload string
	uuidFile     string

	renderingType string

	// Transfer method flag.
	clearTransfer bool

	cfg *config.Config
	c   *ephcli.ClientEphemeralfiles
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "eph",
	Short: "ephemeralfiles command line interface",
	Long:  `ephemeralfiles command line interface`,
}

// Execute runs the root command and handles any execution errors.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	_, err := os.UserHomeDir()
	if err != nil {
		if homeEnv := os.Getenv("HOME"); homeEnv == "" {
			fmt.Println("warn: $HOME not set")
		}
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// -c option to specify the configuration file
	rootCmd.PersistentFlags().StringVarP(&configurationFile, "config", "c",
		config.DefaultConfigFilePath(), "configuration file")
	// -d option to enable debug mode
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "enable debug mode (disable progress bar)")

	// upload subcommand parameters
	uploadCmd.PersistentFlags().StringVarP(&fileToUpload, "input", "i", "", "file to upload")
	uploadCmd.PersistentFlags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
	uploadCmd.PersistentFlags().BoolVar(&clearTransfer, "clear", false, "upload without encryption")
	// download subcommand parameters
	downloadCmd.PersistentFlags().StringVarP(&uuidFile, "input", "i", "", "uuid of file to download")
	downloadCmd.PersistentFlags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
	downloadCmd.PersistentFlags().BoolVar(&clearTransfer, "clear", false, "download without encryption")
	// list subcommand parameters
	listCmd.PersistentFlags().StringVarP(&renderingType, "rendering", "r", "table", "rendering type (table, json, csv)")
	// remove subcommand parameters
	removeCmd.PersistentFlags().StringVarP(&uuidFile, "input", "i", "", "uuid of file to download")
	// config subcommand parameters
	configCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "ephemeralfiles token")
	configCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "ephemeralfiles endpoint")


	// add subcommands
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(configCmd)
	if runtime.GOOS != "windows" {
		rootCmd.AddCommand(autoupdateCmd)
	}
}

// InitClient initializes the client.
func InitClient() {
	cfg = config.NewConfig()
	err := cfg.LoadConfiguration(configurationFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %s\n", err)
		os.Exit(1)
	}
	c = ephcli.NewClient(cfg.Token)
	if cfg.Endpoint != "" {
		c.SetEndpoint(cfg.Endpoint)
	}
	if noProgressBar {
		c.DisableProgressBar()
	}
	if debugMode {
		c.SetDebug()
	}
}
