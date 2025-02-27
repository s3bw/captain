package cmd

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	title     string
	do        Do
	options   []string
	style     lipgloss.Style
	cursor    int
	confirmed bool
	quitting  bool
}

func newConfirmationModel(do Do, title string, options []string, style lipgloss.Style) confirmModel {
	return confirmModel{
		do:      do,
		title:   title,
		options: options,
		style:   style,
		cursor:  0,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "right":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.confirmed = m.cursor == 0
			m.quitting = true
			return m, tea.Quit
		case "q", "ctrl+c":
			m.confirmed = false
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

var (
	greenStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	redStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	normalStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m confirmModel) View() string {
	var b strings.Builder

	// Show task details
	b.WriteString(m.style.Render(m.title))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("ID: %s\n", m.style.Render(fmt.Sprintf("%d", m.do.ID))))
	b.WriteString(fmt.Sprintf("Description: %s\n", m.style.Render(m.do.Description)))
	b.WriteString("\n")

	// Show options with arrow indicator
	for i, option := range m.options {
		if i == m.cursor {
			b.WriteString("→ ") // Arrow indicator
			b.WriteString(m.style.Render(option))
		} else {
			b.WriteString("  ") // Space for alignment
			b.WriteString(normalStyle.Render(option))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑/↓ arrows to select • enter to confirm"))

	return b.String()
}

func Confirmation(do Do, title string, style lipgloss.Style) bool {
	m := newConfirmationModel(do, title, []string{"Yes", "No"}, style)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("could not run program: %v", err)
	}
	return finalModel.(confirmModel).confirmed
}
