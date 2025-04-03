package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

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
	case Learn:
		return color.New(color.FgCyan).Sprintf("learn")
	case PR:
		return color.New(color.FgHiRed).Sprintf("PR")
	case Meta:
		return color.New(color.FgCyan).Sprintf("meta")
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

func fmtBool(b bool) string {
	if b {
		return color.New(color.FgGreen).Sprintf("true")
	}
	return color.New(color.FgRed).Sprintf("false")
}

func fmtDate(task Do) string {
	date := task.CreatedAt
	colour := color.New(color.FgHiBlack)
	if task.Completed {
		date = *task.CompletedAt
		colour = color.New(color.Underline)
	}

	return colour.Sprintf("%s", date.Format("02-Jan-06 15:04"))
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func WidthFunc(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

func DoLog(conn *gorm.DB, query *gorm.DB, unhide bool) {
	var tasks []Do

	if err := query.Preload("Doc").Preload("Tags").Find(&tasks).Error; err != nil {
		log.Fatalf("could not fetch tasks: %v", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
	} else {
		// Header
		tbl := table.New("", "", "do", "at", "doc", "type", "prio", "for")
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		tbl.WithHeaderFormatter(headerFmt)

		columnFmt := color.New(color.FgYellow).SprintfFunc()
		tbl.WithColumnFormatters(0, columnFmt)

		formatID := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(1, formatID)

		formatTime := color.New(color.FgHiBlack).SprintfFunc()
		tbl.WithColumnFormatters(3, formatTime)

		tbl.WithWidthFunc(WidthFunc)

		for _, task := range tasks {
			docIndicator := ""
			if task.Doc.ID != 0 { // If Doc exists, it will have a non-zero ID
				docIndicator = "+"
			}

			tag := ""
			if len(task.Tags) > 0 {
				tag = task.Tags[0].Name
			}

			taskType := fmtDo(task)
			checkBox := fmtBox(task)
			prio := fmtPrio(task)

			description := task.Description
			if task.Sensitive && !unhide {
				description = strings.Repeat("⠿", len(task.Description))
			}

			tbl.AddRow(
				checkBox,
				task.ID,
				description,
				fmtDate(task),
				docIndicator,
				taskType,
				prio,
				tag,
			)
		}

		tbl.Print()
	}
}

func CrewLog(crew []Mate) {
	if len(crew) == 0 {
		fmt.Println("We've got no crew!")
		return
	}

	tbl := table.New("name", "count")
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	tbl.WithHeaderFormatter(headerFmt)

	for _, mate := range crew {
		tbl.AddRow(mate.Name, mate.Count)
	}

	tbl.Print()
}

func DoDetails(conn *gorm.DB, query *gorm.DB) {
	var task Do	
	if err := query.Preload("Doc").Preload("Tags").First(&task).Error; err != nil {
		log.Fatalf("could not fetch task: %v", err)
	}

	fmt.Printf("[id=%d]: \t%s\n", task.ID, highlightStyle.Render(task.Description))
	// Fix display of tags on a single line
	tagString := ""
	for _, tag := range task.Tags {
		tagString += tag.Name
	}
	fmt.Printf("for: \t\t%s\n", tagString)
	fmt.Printf("type: \t\t%s\n", fmtDo(task))
	fmt.Printf("prio: \t\t%s\n", fmtPrio(task))
	fmt.Printf("pinned: \t%s\n", fmtBool(task.Pinned))
	fmt.Printf("sensitive: \t%s\n", fmtBool(task.Sensitive))
	fmt.Printf("deleted: \t%s\n", fmtBool(task.Deleted))
	fmt.Printf("doc: \t\t%s\n", fmtBool(task.Doc.ID != 0))
	fmt.Printf("created_at: \t%s\n", task.CreatedAt)
	fmt.Printf("completed_at: \t%s\n", task.CompletedAt)
}