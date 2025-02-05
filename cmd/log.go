package cmd

import (
	"fmt"
	"log"
	"regexp"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
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

func fmtPrio(task Do) string {
	light := task.Priority

	switch task.Priority {
	case Low:
		return color.New(color.FgHiBlack).Sprintf("%s", light)
	case Medium:
		return color.New(color.FgHiBlack).Sprintf("%s", light)
	case High:
		return color.New(color.FgRed).Sprintf("%s", light)
	default:
		return color.New(color.FgHiBlack).Sprintf("%s", light)
	}
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func WidthFunc(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

func DoLog(conn *gorm.DB, query *gorm.DB) {

	var tasks []Do

	if err := query.Find(&tasks).Error; err != nil {
		log.Fatalf("could not fetch tasks: %v", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
	} else {
		// Header
		tbl := table.New("", "", "do", "did at", "doc", "type", "prio")
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		tbl.WithHeaderFormatter(headerFmt)

		columnFmt := color.New(color.FgYellow).SprintfFunc()
		tbl.WithColumnFormatters(0, columnFmt)

		formatID := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(1, formatID)

		formatTime := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(3, formatTime)

		tbl.WithWidthFunc(WidthFunc)

		// Should we join with tag here can display the
		// tag??
		for _, task := range tasks {
			taskType := fmtDo(task)
			checkBox := fmtBox(task)
			prio := fmtPrio(task)

			tbl.AddRow(
				checkBox,
				task.ID,
				task.Description,
				task.CreatedAt.Format("02-Jan-06 15:04"),
				"",
				taskType,
				prio,
			)
		}

		tbl.Print()
	}
}
