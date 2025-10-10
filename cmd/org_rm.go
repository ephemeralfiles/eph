package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	orgRmFile  string
	orgRmForce bool
)

// orgDeleteCmd represents the organization delete file command.
var orgDeleteCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete file from organization",
	Long:  `Delete a file from an organization by file ID.`,
	Run: func(_ *cobra.Command, _ []string) {
		InitClient()

		if orgRmFile == "" {
			fmt.Fprintf(os.Stderr, "Error: --input flag is required\n")
			os.Exit(1)
		}

		// Confirmation prompt unless --force
		if !orgRmForce {
			fmt.Printf("Are you sure you want to delete file %s? (y/N): ", orgRmFile)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
				os.Exit(1)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Deletion cancelled")
				return
			}
		}

		err := c.DeleteOrganizationFile(orgRmFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("File deleted successfully")
	},
}

func init() {
	orgDeleteCmd.Flags().StringVarP(&orgRmFile, "input", "i", "", "file ID to delete (required)")
	orgDeleteCmd.Flags().BoolVarP(&orgRmForce, "force", "f", false, "skip confirmation")
}
