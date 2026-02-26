package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var promoteCmd = &cobra.Command{
	Use:   "promote <do_id>",
	Short: "Promote a task to a .do file in VFS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		conn := OpenConn(&cfg)

		// Fetch task with doc
		var do Do
		result := conn.Preload("Doc").
			Where("deleted = ? AND promoted = ?", false, false).
			First(&do, id)
		if result.Error != nil {
			fmt.Printf("No task under id '%v' or already promoted\n", id)
			return
		}

		// Get doc content
		docContent := ""
		if do.Doc.ID != 0 {
			docContent = do.Doc.Text
		}

		// Show task details
		fmt.Printf("\nPromoting task (id=%d):\n", do.ID)
		fmt.Printf("Description: %s\n", do.Description)
		if docContent != "" {
			fmt.Printf("Has documentation: Yes (%d chars)\n", len(docContent))
		} else {
			fmt.Printf("Has documentation: No\n")
		}
		fmt.Println()

		// Prompt for filename
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter filename (without .do extension): ")
		filename, _ := reader.ReadString('\n')
		filename = strings.TrimSpace(filename)

		if filename == "" {
			fmt.Println("Filename cannot be empty. Cancelled.")
			return
		}

		// Initialize VFS
		vfsManager, err := NewVFSManager(conn, cfg.CaptainDir)
		if err != nil {
			fmt.Printf("Error initializing VFS: %v\n", err)
			return
		}

		// Create .do file
		err = vfsManager.CreatePromotedFile(filename, do.Description, docContent)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}

		// Mark as promoted
		do.Promoted = true
		conn.Save(&do)
		vfsManager.Save()

		fmt.Printf("\n✓ Task promoted to file: %s.do\n", filename)
		fmt.Printf("✓ Task marked as promoted (id=%d)\n", do.ID)
	},
}

func init() {
	RootCmd.AddCommand(promoteCmd)
}
