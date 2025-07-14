package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/ephemeralfiles/eph/pkg/github"
	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

const (
	// WritablePermission is the file permission used for checking write access.
	WritablePermission = 0666
)

// autoupdateCmd represents the autoupdate command.
var autoupdateCmd = &cobra.Command{
	Use:   "autoupdate",
	Short: "autoupdate binary",
	Long: `autoupdate binary.
`,
	Run: func(_ *cobra.Command, _ []string) {
		// check if binary is writable
		if !checkIfBinaryIsWritable(os.Args[0]) {
			fmt.Fprintf(os.Stderr, "binary is not writable\n")
			fmt.Fprintf(os.Stderr, "If authorized, add write permissions\n")
			fmt.Fprintf(os.Stderr, "Try with sudo on Linux/MacOS or ask to your system administrator\n")

			return
		}
		ghSvc := github.NewClient()
		lastVersionFromGithub, err := ghSvc.GetLastVersionFromGithub(GithubRepository)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while getting last version from github: %v\n", err)

			return
		}
		if IsLastVersion(lastVersionFromGithub) {
			if version == "development" {
				fmt.Println("Development version - no update")

				return
			}
			fmt.Println("Already up to date")

			return
		}
		fmt.Println("last version from github:", lastVersionFromGithub)
		err = autoUpdateBinary(lastVersionFromGithub)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while updating binary: %v\n", err)

			return
		}
	},
}

// IsLastVersion checks if the actual version is the last version from github.
func IsLastVersion(lastVersionFromGithub string) bool {
	if version == "development" {
		return true
	}
	if version == lastVersionFromGithub {
		return true
	}
	return false
}

func checkIfBinaryIsWritable(filePath string) bool {
	// Attempt to open the file with read-write permissions without truncating it.
	// This is a simple way to check for writability without modifying the file.
	// #nosec G304 -- filePath is controlled by the application, checking binary permissions
	file, err := os.OpenFile(filePath, os.O_WRONLY, WritablePermission)
	if err != nil {
		// If there's an error opening the file with write permissions, it's not writable.
		// This could be due to the file not existing, not having write permissions, etc.
		return false
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", closeErr)
		}
	}()

	// If the file can be opened with write permissions, it's writable.
	return true
}

func autoUpdateBinary(lastVersionFromGithub string) error {
	// get actual architecture
	arch := runtime.GOARCH
	// 	The possible values for [`runtime.GOARCH`] in Go include:
	// - `amd64` for x86-64 architecture
	// - `386` for x86 (32-bit) architecture
	// - `arm` for 32-bit ARM architecture
	// - `arm64` for ARMv8 64-bit architecture
	// - `ppc64` for 64-bit PowerPC architecture
	// - `ppc64le` for 64-bit PowerPC little endian architecture
	// - `mips`, `mipsle` (MIPS 32-bit, little endian), `mips64`, `mips64le` (MIPS 64-bit, little endian)
	// - `s390x` for IBM System z
	// - `riscv64` for 64-bit RISC-V
	// This list is not exhaustive and can change as Go adds support for more architectures.
	// get actual os
	os := runtime.GOOS
	// 	The possible values for runtime.GOOS in Go include:
	// android
	// darwin for macOS
	// dragonfly
	// freebsd
	// linux
	// netbsd
	// openbsd
	// plan9
	// solaris
	// windows
	fmt.Println("os:", os)
	fmt.Println("arch:", arch)
	url := GenerateBinaryURL(GithubRepository, lastVersionFromGithub, os, arch)
	fmt.Println("url:", url)
	return DoUpdate(url)
}

// GenerateBinaryURL generates the download URL for a binary from GitHub releases.
func GenerateBinaryURL(repository string, version string, os string, arch string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/eph_%s_%s_%s", repository, version, version, os, arch)
}

// DoUpdate updates the binary from the url.
func DoUpdate(url string) error {
	const defaultTimeout = 5 * time.Minute
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error while downloading binary: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error while downloading binary: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		return fmt.Errorf("error while updating binary: %w", err)
	}
	return nil
}
