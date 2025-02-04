package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cap",
	Short: "Task manager CLI",
}

var doCmd = &cobra.Command{
	Use:   "do <message>",
	Short: "Add a new do",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]

		conn := OpenConn()

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

		conn := OpenConn()

		var fetched Do

		result := conn.Where("deleted = ?", false).First(&fetched, id)

		if result.Error != nil {
			fmt.Printf("No task under %v\n", id)
			return
		}

		fetched.Completed = true
		now := time.Now()
		fetched.CompletedAt = &now
		conn.Save(&fetched)

		fmt.Printf("Marked %d as done\n", fetched.ID)
	},
}

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

		conn := OpenConn()

		var do Do

		result := conn.First(&do, id)

		if result.Error != nil {
			fmt.Printf("No do under id '%v'\n", id)
			return
		}

		do.Deleted = true
		conn.Save(&do)
		fmt.Printf("Deleted do %d\n", do.ID)
	},
}

var logCmd = &cobra.Command{
	Use:   "log --include-done --sort=created_at",
	Short: "Log tasks",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn()
		// cmd.Flags().Bool("include-done", false, "Include completed tasks")
		// cmd.Flags().String("sort", "created_at", "Sort tasks by the specifics")
		n, _ := cmd.Flags().GetInt("n")

		// includeDone, _ := cmd.Flags().GetBool("include-done")
		// sortField, _ := cmd.Flags().GetString("sort")
		// --type
		// --target
		// --sort

		// How do we handle these?
		// --include-done
		// --done
		DoLog(conn, n)
	},
}

var todayCmd = &cobra.Command{
	Use: "today",
	// We will need to sort this by completed at date time.
	Short: "Log tasks done today",
}

var recruitCmd = &cobra.Command{
	Use:   "recruit <name>",
	Short: "Add someone to target",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		conn := OpenConn()

		newTag := Tag{
			Name: name,
		}

		if err := conn.Create(&newTag).Error; err != nil {
			log.Fatalf("could not insert new row: %v", err)
		}

		fmt.Printf("🏴‍☠️  Say welcome the new recruit! '%s'\n", name)
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename <name> <new_name>",
	Short: "Rename the target",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]

		conn := OpenConn()

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

		conn := OpenConn()

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

		conn := OpenConn()

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

		conn := OpenConn()

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
}

var docCmd = &cobra.Command{
	Use:   "doc <do_id>",
	Short: "Document the specifics",
	Args:  cobra.ExactArgs(1),
}

var viewCmd = &cobra.Command{
	Use:   "view <do_id>",
	Short: "View the do",
	Args:  cobra.ExactArgs(1),
}

func init() {
	logCmd.Flags().IntP("n", "n", 10, "Limit the number of dos outstanding")

	RootCmd.AddCommand(
		doCmd,
		didCmd,
		scratchCmd,
		editCmd,
		logCmd,
		todayCmd,
		recruitCmd,
		renameCmd,
		askCmd,
		tellCmd,
		reassignCmd,
		bragCmd,
		docCmd,
		viewCmd,
	)
}
