package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	GithubRepository = "ephemeralfiles/eph"
	DefaultEndpoint  = "https://api.ephemeralfiles.com"
)

var (
	noProgressBar     bool
	configurationFile string
	token             string
	endpoint          string

	fileToUpload string
	uuidFile     string

	renderingType string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eph",
	Short: "ephemeralfiles command line interface",
	Long:  `ephemeralfiles command line interface`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = os.Getenv("HOME")
		if homedir == "" {
			fmt.Println("warn: $HOME not set")
		}
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// -c option to specify the configuration file
	rootCmd.PersistentFlags().StringVarP(&configurationFile, "config", "c",
		filepath.Join(homedir, ".config/eph/default.yml"),
		"configuration file (default is $HOME/.config/eph/default.yml))")

	// upload subcommand parameters
	uploadCmd.PersistentFlags().StringVarP(&fileToUpload, "input", "i", "", "file to upload")
	uploadCmd.PersistentFlags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
	// download subcommand parameters
	downloadCmd.PersistentFlags().StringVarP(&uuidFile, "input", "i", "", "uuid of file to download")
	downloadCmd.PersistentFlags().BoolVarP(&noProgressBar, "no-progress-bar", "n", false, "disable progress bar")
	// list subcommand parameters
	listCmd.PersistentFlags().StringVarP(&renderingType, "rendering", "r", "table", "rendering type (table, json, csv)")
	// remove subcommand parameters
	removeCmd.PersistentFlags().StringVarP(&uuidFile, "input", "i", "", "uuid of file to download")
	// config subcommand parameters
	configCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "ephemeralfiles token")
	configCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "ephemeralfiles endpoint")

	uploadE2ECmd.PersistentFlags().StringVarP(&fileToUpload, "input", "i", "", "file to upload")

	// add subcommands
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(uploadE2ECmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(configCmd)
	if runtime.GOOS != "windows" {
		rootCmd.AddCommand(autoupdateCmd)
	}
}
