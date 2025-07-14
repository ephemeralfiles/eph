// Package main is the entry point for the eph CLI application.
// eph is the official command-line client for ephemeralfiles.com,
// providing file upload, download, and management capabilities with
// end-to-end encryption support.
package main

import "github.com/ephemeralfiles/eph/cmd"

func main() {
	cmd.Execute()
}
