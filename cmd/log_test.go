package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestFmtBox(t *testing.T) {
	tests := []struct {
		name     string
		do       Do
		expected checkBox
	}{
		{
			name:     "completed task",
			do:       Do{Completed: true},
			expected: done,
		},
		{
			name:     "incomplete task",
			do:       Do{Completed: false},
			expected: notDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtBox(tt.do)
			if result != tt.expected {
				t.Errorf("fmtBox() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestFmtDo(t *testing.T) {
	tests := []struct {
		name     string
		do       Do
		contains string // substring that should be in the colored output
	}{
		{"task type", Do{Type: Task}, "task"},
		{"ask type", Do{Type: Ask}, "ask"},
		{"tell type", Do{Type: Tell}, "tell"},
		{"brag type", Do{Type: Brag}, "brag"},
		{"learn type", Do{Type: Learn}, "learn"},
		{"PR type", Do{Type: PR}, "PR"},
		{"meta type", Do{Type: Meta}, "meta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtDo(tt.do)
			// Strip ANSI codes for comparison
			stripped := stripANSI(result)
			if !strings.Contains(stripped, tt.contains) {
				t.Errorf("fmtDo() output %s does not contain %s", stripped, tt.contains)
			}
		})
	}
}

func TestFmtPrio(t *testing.T) {
	tests := []struct {
		name     string
		do       Do
		contains string
	}{
		{"low priority", Do{Priority: Low}, "low"},
		{"medium priority", Do{Priority: Medium}, "medium"},
		{"high priority", Do{Priority: High}, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtPrio(tt.do)
			stripped := stripANSI(result)
			if !strings.Contains(stripped, tt.contains) {
				t.Errorf("fmtPrio() output %s does not contain %s", stripped, tt.contains)
			}
		})
	}
}

func TestFmtDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		do   Do
	}{
		{
			name: "created task",
			do: Do{
				CreatedAt: now,
				Completed: false,
			},
		},
		{
			name: "completed task",
			do: Do{
				CreatedAt:   now.Add(-24 * time.Hour),
				Completed:   true,
				CompletedAt: &now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtDate(tt.do)
			if result == "" {
				t.Error("Expected non-empty date string")
			}
			// Verify it contains date components
			stripped := stripANSI(result)
			if len(stripped) == 0 {
				t.Error("Expected formatted date to have content")
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "text with ANSI codes",
			input:    "\x1b[31mRed Text\x1b[0m",
			expected: "Red Text",
		},
		{
			name:     "text without ANSI codes",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFmtBool(t *testing.T) {
	tests := []struct {
		name  string
		input bool
	}{
		{"true value", true},
		{"false value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtBool(tt.input)
			if result == "" {
				t.Error("Expected non-empty string")
			}
			stripped := stripANSI(result)
			if tt.input && !strings.Contains(stripped, "true") {
				t.Error("Expected 'true' in output")
			}
			if !tt.input && !strings.Contains(stripped, "false") {
				t.Error("Expected 'false' in output")
			}
		})
	}
}

func TestFmtReason(t *testing.T) {
	do := Do{Reason: "Not needed"}
	result := fmtReason(do)

	if result == "" {
		t.Error("Expected non-empty string")
	}

	stripped := stripANSI(result)
	if !strings.Contains(stripped, "Not needed") {
		t.Errorf("Expected reason 'Not needed' in output, got %s", stripped)
	}
}

func TestWidthFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "plain ASCII",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "with ANSI codes",
			input:    "\x1b[31mred\x1b[0m",
			expected: 3,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WidthFunc(tt.input)
			if result != tt.expected {
				t.Errorf("WidthFunc() = %d, expected %d", result, tt.expected)
			}
		})
	}
}
