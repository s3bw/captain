package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = conn.AutoMigrate(
		&Do{}, &Tag{}, &DoTag{}, &DoDoc{}, &Template{},
		&FileRecord{}, &DirectoryState{}, &UserPreference{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return conn, cleanup
}

func TestDoModel(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name string
		do   Do
	}{
		{
			name: "create basic task",
			do: Do{
				Description: "Test task",
				Type:        Task,
				Priority:    Medium,
			},
		},
		{
			name: "create high priority ask",
			do: Do{
				Description: "Test question",
				Type:        Ask,
				Priority:    High,
			},
		},
		{
			name: "create completed task",
			do: Do{
				Description: "Completed task",
				Type:        Task,
				Priority:    Low,
				Completed:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := conn.Create(&tt.do).Error; err != nil {
				t.Errorf("Failed to create do: %v", err)
			}

			if tt.do.ID == 0 {
				t.Error("Expected ID to be set after creation")
			}

			if tt.do.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
		})
	}
}

func TestDoCompletion(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	do := Do{
		Description: "Task to complete",
		Type:        Task,
		Priority:    Medium,
		Completed:   false,
	}

	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create do: %v", err)
	}

	// Mark as completed
	do.Completed = true
	now := time.Now()
	do.CompletedAt = &now

	if err := conn.Save(&do).Error; err != nil {
		t.Fatalf("Failed to update do: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Completed {
		t.Error("Expected task to be completed")
	}

	if fetched.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestDoSoftDelete(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	do := Do{
		Description: "Task to delete",
		Type:        Task,
		Priority:    Medium,
	}

	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create do: %v", err)
	}

	// Soft delete
	do.Deleted = true
	do.Reason = "Not needed anymore"

	if err := conn.Save(&do).Error; err != nil {
		t.Fatalf("Failed to update do: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Deleted {
		t.Error("Expected task to be deleted")
	}

	if fetched.Reason != "Not needed anymore" {
		t.Errorf("Expected reason to be 'Not needed anymore', got %s", fetched.Reason)
	}
}

func TestTagModel(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	tag := Tag{Name: "alice"}

	if err := conn.Create(&tag).Error; err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	// Test uniqueness
	duplicate := Tag{Name: "alice"}
	if err := conn.Create(&duplicate).Error; err == nil {
		t.Error("Expected error when creating duplicate tag")
	}
}

func TestDoTagRelationship(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a tag
	tag := Tag{Name: "bob"}
	if err := conn.Create(&tag).Error; err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Create a do
	do := Do{
		Description: "Ask Bob something",
		Type:        Ask,
		Priority:    Medium,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create do: %v", err)
	}

	// Create relationship
	doTag := DoTag{
		DoID:  do.ID,
		TagID: tag.ID,
	}
	if err := conn.Create(&doTag).Error; err != nil {
		t.Fatalf("Failed to create do-tag relationship: %v", err)
	}

	// Verify relationship
	var fetchedDo Do
	if err := conn.Preload("Tags").First(&fetchedDo, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do with tags: %v", err)
	}

	if len(fetchedDo.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(fetchedDo.Tags))
	}

	if fetchedDo.Tags[0].Name != "bob" {
		t.Errorf("Expected tag name 'bob', got %s", fetchedDo.Tags[0].Name)
	}
}

func TestDoDocModel(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	do := Do{
		Description: "Task with docs",
		Type:        Task,
		Priority:    Medium,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create do: %v", err)
	}

	doc := DoDoc{
		DoID: do.ID,
		Text: "# Documentation\n\nThis is a test document.",
	}
	if err := conn.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create doc: %v", err)
	}

	// Verify
	var fetchedDo Do
	if err := conn.Preload("Doc").First(&fetchedDo, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do with doc: %v", err)
	}

	if fetchedDo.Doc.ID == 0 {
		t.Error("Expected doc to be loaded")
	}

	if fetchedDo.Doc.Text != "# Documentation\n\nThis is a test document." {
		t.Errorf("Expected doc text to match, got %s", fetchedDo.Doc.Text)
	}
}

func TestTemplateModel(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	template := Template{
		Name:    "meeting",
		Content: "Meeting with {{ person:string }}",
		Deleted: false,
	}

	if err := conn.Create(&template).Error; err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	if template.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if template.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestTemplateUniqueness(t *testing.T) {
	conn, cleanup := setupTestDB(t)
	defer cleanup()

	template1 := Template{
		Name:    "standup",
		Content: "Standup notes",
		Deleted: false,
	}

	if err := conn.Create(&template1).Error; err != nil {
		t.Fatalf("Failed to create first template: %v", err)
	}

	// Try to create another with same name
	template2 := Template{
		Name:    "standup",
		Content: "Different content",
		Deleted: false,
	}

	if err := conn.Create(&template2).Error; err == nil {
		t.Error("Expected error when creating duplicate template name")
	}
}

func TestDoTypes(t *testing.T) {
	types := []DoType{Task, Ask, Tell, Brag, Learn, PR, Meta}

	for _, doType := range types {
		t.Run(string(doType), func(t *testing.T) {
			if string(doType) == "" {
				t.Error("DoType should not be empty")
			}
		})
	}
}

func TestDoPriorities(t *testing.T) {
	priorities := []DoPrio{Low, Medium, High}

	for _, prio := range priorities {
		t.Run(string(prio), func(t *testing.T) {
			if string(prio) == "" {
				t.Error("DoPrio should not be empty")
			}
		})
	}
}
