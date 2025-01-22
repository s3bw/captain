package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cap",
	Short: "Task manager CLI",
}

var doCmd = &cobra.Command{
	Use:   "do <message>",
	Short: "Add a new task",
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

		fmt.Printf("Added task: (id:%d)\n", do.ID)
	},
}

var didCmd = &cobra.Command{
	Use:   "did <do_id>",
	Short: "Remove a task by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		conn := OpenConn()

		var fetched Do

		result := conn.First(&fetched, id)

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

var logCmd = &cobra.Command{
	Use:   "log --include-done --sort=created_at",
	Short: "Log tasks",
	Run: func(cmd *cobra.Command, args []string) {
		conn := OpenConn()

		// includeDone, _ := cmd.Flags().GetBool("include-done")
		// sortField, _ := cmd.Flags().GetString("sort")

		var tasks []Do

		query := conn
		query = query.Order("created_at DESC")

		if err := query.Find(&tasks).Error; err != nil {
			log.Fatalf("could not fetch tasks: %v", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
		} else {
			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()
			tbl := table.New("Done", "ID", "Description", "Created At")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, task := range tasks {
				if task.Completed {
					tbl.AddRow("▣", task.ID, task.Description, task.CreatedAt)
				} else {
					tbl.AddRow("☐", task.ID, task.Description, task.CreatedAt)
				}
			}

			tbl.Print()
		}
	},
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

		fmt.Printf("Let's ask %s (id:%d)\n", name, do.ID)
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

		fmt.Printf("Let's tell %s (id:%d)\n", name, do.ID)
	},
}

var bragCmd = &cobra.Command{
	Use:   "brag <name> <message>",
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

var docCmd = &cobra.Command{
	Use:   "doc <do_id>",
	Short: "Document the specifics",
	Args:  cobra.ExactArgs(1),
}

var viewCmd = &cobra.Command{
	Use:   "view <do_id>",
	Short: "View the task",
	Args:  cobra.ExactArgs(1),
}

func init() {
	logCmd.Flags().Bool("include-done", false, "Include completed tasks")
	logCmd.Flags().String("sort", "created_at", "Sort tasks by the specifics")

	RootCmd.AddCommand(
		doCmd,
		didCmd,
		editCmd,
		logCmd,
		recruitCmd,
		renameCmd,
		askCmd,
		tellCmd,
		bragCmd,
		docCmd,
		viewCmd,
	)
}
