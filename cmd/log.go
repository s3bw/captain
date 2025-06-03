package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	sebtable "github.com/s3bw/table"
	"gorm.io/gorm"
)

type checkBox string

const (
	done    checkBox = "▣"
	notDone checkBox = "☐"
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

func fmtBox(task Do) checkBox {
	if task.Completed {
		return done
	}
	return notDone
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
		re := lipgloss.NewRenderer(os.Stdout)
		baseStyle := re.NewStyle().Padding(0, 1)

		// Header
		// headers := []string{"", "#", "do", "at", "doc", "type", "prio", "for"}
		headers := []string{"", "#", "Do", "At", "Doc", "Type", "Prio", "For"}

		var data [][]string

		for _, task := range tasks {
			docIndicator := ""
			if task.Doc.ID != 0 { // If Doc exists, it will have a non-zero ID
				docIndicator = "✻"
			}

			tag := ""
			if len(task.Tags) > 0 {
				tag = task.Tags[0].Name
			}

			taskType := fmtDo(task)
			checkBx := fmtBox(task)
			prio := fmtPrio(task)

			description := task.Description
			if task.Sensitive && !unhide {
				description = strings.Repeat("⠿", len(task.Description))
			}

			data = append(data, []string{
				string(checkBx),
				strconv.Itoa(int(task.ID)),
				description,
				fmtDate(task),
				docIndicator,
				taskType,
				prio,
				tag,
			},
			)
		}

		headerStyle := baseStyle.Foreground(lipgloss.Color("37")).Bold(true)

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
			Headers(headers...).
			Width(0).
			Rows(data...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}

				even := row%2 == 0

				switch col {
				case 0:
					if checkBox(data[row][0]) == done {
						return baseStyle.Foreground(lipgloss.Color("43"))
					}
					return baseStyle.Foreground(lipgloss.Color("138"))
				case 2, 4, 6, 7:
					if even {
						return baseStyle.Foreground(lipgloss.Color("223"))
					}
					return baseStyle.Foreground(lipgloss.Color("216"))
				}
				return baseStyle.Foreground(lipgloss.Color("245"))
			})

		fmt.Println(t)
	}
}

func CrewLog(crew []Mate) {
	if len(crew) == 0 {
		fmt.Println("We've got no crew!")
		return
	}

	tbl := sebtable.New("name", "count")
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
