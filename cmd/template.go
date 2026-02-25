package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/s3bw/mostxt/src"
	sebtable "github.com/s3bw/table"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage task templates",
	Long:  "Create, list, edit, and delete task templates with mostxt placeholders",
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		conn := OpenConn(&cfg)

		// Check if template already exists
		var existing Template
		result := conn.Where("name = ? AND deleted = ?", name, false).First(&existing)
		if result.Error == nil {
			fmt.Printf("Template '%s' already exists. Use 'captain template edit %s' to modify it.\n", name, name)
			return
		}

		// Create temporary file for editing
		tmpfile, err := os.CreateTemp("", fmt.Sprintf("template-create-%s-*.md", name))
		if err != nil {
			fmt.Printf("Error creating temporary file: %v\n", err)
			return
		}
		defer os.Remove(tmpfile.Name())

		// Open editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		editorCmd := exec.Command(editor, tmpfile.Name())
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		err = editorCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Read the content
		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			fmt.Printf("Error reading template content: %v\n", err)
			return
		}

		contentStr := strings.TrimSpace(string(content))
		if contentStr == "" {
			fmt.Println("Template content is empty. Not saving.")
			return
		}

		// Validate template syntax
		_, err = src.ParseTemplate(contentStr)
		if err != nil {
			fmt.Printf("Error parsing template syntax: %v\n", err)
			fmt.Println("Template not saved. Please fix the syntax and try again.")
			return
		}

		// Save to database
		template := Template{
			Name:    name,
			Content: contentStr,
			Deleted: false,
		}
		conn.Create(&template)
		fmt.Printf("Created template '%s'\n", name)
	},
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn(&cfg)

		var templates []Template
		conn.Where("deleted = ?", false).Order("updated_at DESC").Find(&templates)

		if len(templates) == 0 {
			fmt.Println("No templates found.")
			return
		}

		tbl := sebtable.New("name", "preview", "updated")
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		tbl.WithHeaderFormatter(headerFmt)

		for _, tmpl := range templates {
			// Create preview (first 50 chars, replace newlines with spaces)
			preview := strings.ReplaceAll(tmpl.Content, "\n", " ")
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}

			// Format updated date
			updated := tmpl.UpdatedAt.Format("2006-01-02")

			tbl.AddRow(tmpl.Name, preview, updated)
		}

		tbl.Print()
	},
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		conn := OpenConn(&cfg)

		var template Template
		result := conn.Where("name = ? AND deleted = ?", name, false).First(&template)
		if result.Error != nil {
			fmt.Printf("No template named '%s'\n", name)
			return
		}

		// Create temporary file with current content
		tmpfile, err := os.CreateTemp("", fmt.Sprintf("template-edit-%s-*.md", name))
		if err != nil {
			fmt.Printf("Error creating temporary file: %v\n", err)
			return
		}
		defer os.Remove(tmpfile.Name())

		// Write current content
		_, err = tmpfile.WriteString(template.Content)
		if err != nil {
			fmt.Printf("Error writing to temporary file: %v\n", err)
			return
		}

		// Open editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		editorCmd := exec.Command(editor, tmpfile.Name())
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		err = editorCmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Read the edited content
		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			fmt.Printf("Error reading edited content: %v\n", err)
			return
		}

		contentStr := strings.TrimSpace(string(content))

		// Validate template syntax
		_, err = src.ParseTemplate(contentStr)
		if err != nil {
			fmt.Printf("Error parsing template syntax: %v\n", err)
			fmt.Println("Template not saved. Please fix the syntax and try again.")
			return
		}

		// Update the template
		template.Content = contentStr
		conn.Save(&template)
		fmt.Printf("Updated template '%s'\n", name)
	},
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		conn := OpenConn(&cfg)

		var template Template
		result := conn.Where("name = ? AND deleted = ?", name, false).First(&template)
		if result.Error != nil {
			fmt.Printf("No template named '%s'\n", name)
			return
		}

		// Simple confirmation prompt
		fmt.Printf("Delete template '%s'? (y/n): ", name)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			// Soft delete
			template.Deleted = true
			conn.Save(&template)
			fmt.Printf("Deleted template '%s'\n", name)
		} else {
			fmt.Println("Template deletion cancelled")
		}
	},
}

func init() {
	templateCmd.AddCommand(
		templateCreateCmd,
		templateListCmd,
		templateEditCmd,
		templateDeleteCmd,
	)
}
