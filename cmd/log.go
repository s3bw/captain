package cmd

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/s3bw/table"
	"gorm.io/gorm"
)

func fmtDo(task Do) string {
	switch task.Type {
	case Task:
		return color.New(color.FgGreen).Sprintf("task")
	case Ask:
		return color.New(color.FgYellow).Sprintf("ask")
	case Tell:
		return color.New(color.FgBlue).Sprintf("tell")
	case Brag:
		return color.New(color.FgMagenta).Sprintf("brag")
	default:
		return "unknown"
	}
}

func fmtBox(task Do) string {
	if task.Completed {
		return "▣"
	}
	return "☐"
}

func DoLog(conn *gorm.DB, n int) {

	var tasks []Do

	query := conn
	query = query.Not("deleted = ?", true).Limit(n).Order("completed, created_at DESC")

	if err := query.Find(&tasks).Error; err != nil {
		log.Fatalf("could not fetch tasks: %v", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
	} else {
		// Header
		tbl := table.New("", "", "do", "did at", "doc", "type")
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		tbl.WithHeaderFormatter(headerFmt)

		columnFmt := color.New(color.FgYellow).SprintfFunc()
		tbl.WithColumnFormatters(0, columnFmt)

		formatID := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(1, formatID)

		formatTime := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(3, formatTime)

		// Should we join with tag here can display the
		// tag??
		for _, task := range tasks {
			taskType := fmtDo(task)
			checkBox := fmtBox(task)

			tbl.AddRow(checkBox, task.ID, task.Description, task.CreatedAt.Format("02-Jan-06 15:04"), "", taskType)
		}

		tbl.Print()
	}
}
