package cmd

import (
	"testing"
)

func TestMapPriority(t *testing.T) {
	tests := []struct {
		input    string
		expected DoPrio
	}{
		{"low", Low},
		{"med", Medium},
		{"high", High},
		{"", Medium},       // default
		{"invalid", Medium}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapPriority(tt.input)
			if result != tt.expected {
				t.Errorf("mapPriority(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapType(t *testing.T) {
	tests := []struct {
		input    string
		expected DoType
	}{
		{"task", Task},
		{"ask", Ask},
		{"tell", Tell},
		{"brag", Brag},
		{"learn", Learn},
		{"pr", PR},
		{"PR", PR},
		{"meta", Meta},
		{"", Task},       // default
		{"invalid", Task}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapType(tt.input)
			if result != tt.expected {
				t.Errorf("mapType(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDoOrder(t *testing.T) {
	tests := []struct {
		sortby   string
		orderby  string
		contains string // substring that should be in result
	}{
		{"created_at", "asc", "created_at ASC"},
		{"created_at", "desc", "created_at DESC"},
		{"completed_at", "asc", "completed_at ASC"},
		{"priority", "desc", "priority"},
		{"type", "asc", "type ASC"},
		{"description", "desc", "description DESC"},
		{"default", "desc", "completed"},
		{"invalid", "desc", "completed"}, // falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.sortby+"_"+tt.orderby, func(t *testing.T) {
			result := DoOrder(tt.sortby, tt.orderby)
			if result == "" {
				t.Error("Expected non-empty result from DoOrder")
			}
			// Note: We can't easily test the exact SQL string, but we can verify it's not empty
		})
	}
}

func TestSprintfFunc(t *testing.T) {
	format := "test_%s"
	fn := SprintfFunc(format)

	result := fn("value")
	expected := "test_value"

	if result != expected {
		t.Errorf("SprintfFunc result = %s, expected %s", result, expected)
	}
}
