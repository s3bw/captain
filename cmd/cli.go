package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

var RootCmd = &cobra.Command{
	Use:   "cap",
	Short: "Task manager CLI",
}

var cfg = LoadConfig()

var doCmd = &cobra.Command{
	Use:   "do <message>",
	Short: "Add a new do",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]

		conn := OpenConn(&cfg)

		do := Do{
			Description: message,
			Type:        Task,
		}

		if err := conn.Create(&do).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("Added do: (id=%d)\n", do.ID)
	},
}

var didCmd = &cobra.Command{
	Use:   "did <do_id>",
	Short: "Complete a do by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		conn := OpenConn(&cfg)

		var fetched Do

		result := conn.Where("deleted = ?", false).First(&fetched, id)
		if result.Error != nil {
			fmt.Printf("No task under %v\n", id)
			return
		}

		if Confirmation(fetched, "Complete this task?", greenStyle) {
			fetched.Completed = true
			now := time.Now()
			fetched.CompletedAt = &now
			conn.Save(&fetched)
			fmt.Printf("Marked %d as done\n", fetched.ID)
		} else {
			fmt.Println("Task completion cancelled")
		}
	},
}

func mapPriority(s string) DoPrio {
	switch s {
	case "low":
		return Low
	case "med":
		return Medium
	case "high":
		return High
	default:
		return Medium
	}
}

var setPrioCmd = &cobra.Command{
	Use:   "set <field> <value> <do_id>",
	Short: "Changes something of a do, right now just priority",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		field := args[0]
		if field != "prio" {
			fmt.Printf("The field '%s' is not supported.", field)
			return
		}
		value := args[1]
		id := args[2]

		conn := OpenConn(&cfg)

		var do Do

		result := conn.Not("deleted = ?", false).First(&do, id)
		if result.Error != nil {
			fmt.Printf("No do with ID '%v'\n", id)
			return
		}

		oldPrio := do.Priority

		do.Priority = mapPriority(value)
		conn.Save(&do)

		fmt.Printf("Do %v updated '%v' -> '%v'\n", field, oldPrio, do.Priority)
	},
}

// TODO
var editCmd = &cobra.Command{
	Use:   "edit <do_id> <new_message>",
	Short: "Edit a task by ID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		newMessage := args[1]
		fmt.Printf("Edited task %s: %s\n", id, newMessage)
	},
}

var scratchCmd = &cobra.Command{
	Use:   "scratch <do_id>",
	Short: "Soft delete a do",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		conn := OpenConn(&cfg)

		var do Do

		result := conn.Not("deleted = ?", true).First(&do, id)
		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		if Confirmation(do, "Delete this task?", redStyle) {
			do.Deleted = true
			conn.Save(&do)
			fmt.Printf("Deleted do %d\n", do.ID)
		} else {
			fmt.Println("Task deletion cancelled")
		}
	},
}

var unscratchCmd = &cobra.Command{
	Use:   "unscratch <do_id>",
	Short: "Revert the soft deleted do",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		conn := OpenConn(&cfg)

		var do Do

		result := conn.First(&do, id)

		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		if Confirmation(do, "Resurrect this task?", redStyle) {
			do.Deleted = false
			conn.Save(&do)
			fmt.Printf("Resurrected %d\n", do.ID)
		} else {
			fmt.Println("Task resurrection cancelled")
		}
	},
}

var All bool

func SprintfFunc(format string) func(string) string {
	return func(value string) string {
		return fmt.Sprintf(format, value)
	}
}

func DoOrder(sortby string, orderby string) string {
	// DESC / ASC
	var so string

	switch orderby {
	case "asc":
		so = "%s " + "ASC"
	default:
		if orderby != "desc" {
			fmt.Printf("No such order: '%s'!\n", orderby)
		}
		so = "%s " + "DESC"
	}

	sortOrder := SprintfFunc(so)

	switch sortby {
	case "created_at":
		return sortOrder("created_at")
	case "completed_at":
		return sortOrder("completed_at")
	case "description":
		return sortOrder("description")
	case "type":
		return sortOrder("type")
	case "priority":
		return sortOrder(`
			CASE priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
				ELSE 2
			END
		`)
	default:
		if sortby != "default" {
			fmt.Printf("No such sort: '%s'!\n", orderby)
		}
		return `
			completed,
			CASE priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
				ELSE 2
			END, created_at DESC
		`
	}
}

var logCmd = &cobra.Command{
	Use:   "log --include-done --sort=created_at",
	Short: "Log tasks",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn(&cfg)
		// cmd.Flags().Bool("include-done", false, "Include completed tasks")
		n, _ := cmd.Flags().GetInt("n")
		sort, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")

		// --target = recruit
		// Fetching by type
		// --type = brag

		// How do we handle these?
		// --not-done
		query := conn.Not("deleted = ?", true).Limit(n).Order(DoOrder(sort, order))
		if !All {
			lookBack := time.Now().AddDate(0, 0, -cfg.LookBackDays)
			query = query.Where("completed_at IS NULL OR completed_at >= ?", lookBack)
		}

		DoLog(conn, query)
	},
}

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Log tasks done today",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn(&cfg)

		oneDayAgo := time.Now().AddDate(0, 0, -1)

		query := conn
		query = query.Not("deleted = ?", true).
			Where("completed_at IS NULL OR completed_at >= ?", oneDayAgo).
			Limit(100).Order(`
			completed,
			CASE priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
				ELSE 2
			END, created_at DESC
		`)
		DoLog(conn, query)
	},
}

var recruitCmd = &cobra.Command{
	Use:   "recruit <name>",
	Short: "Add someone to target",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		conn := OpenConn(&cfg)

		newTag := Tag{
			Name: name,
		}

		if err := conn.Create(&newTag).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("🏴‍☠️  Say welcome the new recruit! '%s'\n", name)
	},
}

type Mate struct {
	Name  string
	Count int64
}

var crewCmd = &cobra.Command{
	Use:   "crew",
	Short: "List the crew",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn(&cfg)

		var crew []Mate

		result := conn.Table("tags").
			Select("tags.name, COUNT(do_tags.tag_id) AS count").
			Joins("LEFT JOIN do_tags ON do_tags.tag_id = tags.id").
			Group("tags.id").
			Order("count DESC").
			Find(&crew)

		if result.Error != nil {
			fmt.Printf("We have no crew aboard! %v\n", result.Error)
			return
		}

		CrewLog(crew)
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename <name> <new_name>",
	Short: "Rename the target",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]

		conn := OpenConn(&cfg)

		var tag Tag

		result := conn.First(&tag, "name = ?", oldName)
		if result.Error != nil {
			fmt.Printf("No tag under '%v'\n", oldName)
			return
		}

		tag.Name = newName
		conn.Save(&tag)

		coloredOldName := color.New(color.FgYellow).Sprintf("%s", oldName)
		coloredName := color.New(color.FgGreen).Sprintf("%s", newName)
		fmt.Printf("We are now calling '%v' -> '%v'\n", coloredOldName, coloredName)
	},
}

var askCmd = &cobra.Command{
	Use:   "ask <name> <message>",
	Short: "Set ask for someone",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		message := args[1]

		conn := OpenConn(&cfg)

		do := &Do{Description: message, Type: Ask}
		tag := &Tag{Name: name}

		conn.Create(&do)
		conn.Create(&tag)

		doTag := DoTag{DoID: do.ID, TagID: tag.ID}
		if err := conn.Create(&doTag).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("Let's ask %s (id=%d)\n", name, do.ID)
	},
}

var tellCmd = &cobra.Command{
	Use:   "tell <name> <message>",
	Short: "Set tell for someone",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		message := args[1]

		conn := OpenConn(&cfg)

		do := &Do{Description: message, Type: Tell}
		tag := &Tag{Name: name}

		conn.Create(&do)
		conn.Create(&tag)

		doTag := DoTag{DoID: do.ID, TagID: tag.ID}
		if err := conn.Create(&doTag).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("Let's tell %s (id=%d)\n", name, do.ID)
	},
}

var bragCmd = &cobra.Command{
	Use:   "brag <message>",
	Short: "Set brag for achievement",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]

		conn := OpenConn(&cfg)

		now := time.Now()
		do := Do{
			Description: message,
			Type:        Brag,
			Completed:   true,
			CompletedAt: &now,
		}

		if err := conn.Create(&do).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("Added brag: (id=%d)\n", do.ID)
	},
}

var reassignCmd = &cobra.Command{
	Use:   "reassign <do_id> <name>",
	Short: "Reassign the do to someone else.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		name := args[1]

		conn := OpenConn(&cfg)

		var do Do

		result := conn.Where("deleted = ?", false).First(&do, id)
		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		var tag Tag

		result = conn.Where("name = ?", name).First(&tag)
		if result.Error != nil {
			fmt.Printf("No recruit called '%v'\n", name)
			return
		}

		doTag := DoTag{DoID: do.ID, TagID: tag.ID}
		if err := conn.Create(&doTag).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		conn.Save(&doTag)
		fmt.Printf("We've reassigned the do to '%s'\n", name)
	},
}

var docCmd = &cobra.Command{
	Use:   "doc <do_id>",
	Short: "Document the specifics",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		conn := OpenConn(&cfg)

		var do Do
		result := conn.Where("deleted = ?", false).First(&do, id)
		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		// Create a temporary file
		tmpfile, err := os.CreateTemp("", "captain-doc-*.md")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		// Write existing doc if it exists
		var existingDoc DoDoc
		result = conn.Where("do_id = ?", do.ID).First(&existingDoc)
		if result.Error == nil {
			tmpfile.WriteString(existingDoc.Text)
		}
		tmpfile.Close()

		// Get editor from environment or fallback to vim
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		// Open editor
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
			log.Fatal(err)
		}

		// Save to database
		doc := DoDoc{
			DoID: do.ID,
			Text: string(content),
		}

		if result.Error == nil {
			// Update existing doc
			existingDoc.Text = string(content)
			conn.Save(&existingDoc)
		} else {
			// Create new doc
			conn.Create(&doc)
		}

		fmt.Printf("Documentation saved for task %d\n", do.ID)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <do_id>",
	Short: "View the do's documentation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		conn := OpenConn(&cfg)

		var do Do
		result := conn.Preload("Doc").Where("deleted = ?", false).First(&do, id)
		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		if do.Doc.ID == 0 {
			fmt.Printf("No documentation for do %d\n", do.ID)
			return
		}

		// Print task details
		fmt.Printf("Documentation (id=%d): %s\n\n", do.ID, highlightStyle.Render(do.Description))
		fmt.Println(strings.Repeat("-", 40))

		// Initialize glamour renderer
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)

		out, err := r.Render(do.Doc.Text)
		if err != nil {
			fmt.Printf("Error rendering markdown: %v\n", err)
			return
		}

		fmt.Print(out)
	},
}

var configCmd = &cobra.Command{
	Use:   "config <key> <value>",
	Short: "Set items in the config file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		err := cfg.Set(key, value)
		if err != nil {
			fmt.Printf("Updated '%s' -> '%s'", key, value)
		}
	},
}

func init() {
	logCmd.Flags().IntP("n", "n", 10, "Limit the number of dos outstanding")
	logCmd.Flags().StringP("sort", "s", "default", "Set the sort")
	logCmd.Flags().StringP("order", "o", "desc", "Set the order")
	logCmd.Flags().BoolVar(&All, "all", false, "return all instead of filtering")

	RootCmd.AddCommand(
		doCmd,
		didCmd,
		scratchCmd,
		unscratchCmd,
		setPrioCmd,
		editCmd,
		// Views
		logCmd,
		todayCmd,
		// Crew Commands
		// - manage
		recruitCmd,
		crewCmd,
		renameCmd,
		// - assignment
		askCmd,
		tellCmd,
		bragCmd,
		reassignCmd,
		// More Details
		docCmd,
		viewCmd,
		// Edit Config
		configCmd,
	)
}
