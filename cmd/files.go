package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/s3bw/vfs/browser"
	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Browse promoted .do files in VFS (interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn(&cfg)
		filesDir := filepath.Join(cfg.CaptainDir, "files")

		// Launch the VFS browser with captain's database
		if err := browser.RunBrowser(conn, filesDir); err != nil {
			fmt.Printf("Error running browser: %v\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(filesCmd)
}
