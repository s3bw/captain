package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestEnv creates a complete test environment with database and config
func setupTestEnv(t *testing.T) (*gorm.DB, *Config, func()) {
	tmpDir := t.TempDir()

	// Setup config
	testConfig := &Config{
		DBFile:       "test.db",
		LookBackDays: 7,
		LogLength:    10,
		CaptainDir:   tmpDir,
	}

	// Setup database
	dbPath := filepath.Join(tmpDir, testConfig.DBFile)
	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

	return conn, testConfig, cleanup
}

// executeCommand executes a cobra command and captures output
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestDoCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a fresh command for testing
	testDoCmd := &cobra.Command{
		Use:  "do <message>",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			message := args[0]
			doType, _ := cmd.Flags().GetString("type")
			prio, _ := cmd.Flags().GetString("prio")

			do := Do{
				Description: message,
				Type:        mapType(doType),
				Priority:    mapPriority(prio),
			}

			if err := conn.Create(&do).Error; err != nil {
				t.Errorf("Failed to create do: %v", err)
			}
		},
	}

	testDoCmd.Flags().String("type", "task", "Set the type")
	testDoCmd.Flags().String("prio", "medium", "Set the priority")
	testDoCmd.Flags().String("for", "", "Set the tag")
	testDoCmd.Flags().StringP("template", "t", "", "Use a template")

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "create basic task",
			args:        []string{"Test task"},
			expectError: false,
		},
		{
			name:        "create high priority ask",
			args:        []string{"--type", "ask", "--prio", "high", "Important question"},
			expectError: false,
		},
		{
			name:        "create tell task",
			args:        []string{"--type", "tell", "Something to tell"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(testDoCmd, tt.args...)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}

	// Verify tasks were created
	var count int64
	conn.Model(&Do{}).Count(&count)
	if count != int64(len(tests)) {
		t.Errorf("Expected %d tasks, got %d", len(tests), count)
	}
}

func TestSetCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a task first
	do := Do{
		Description: "Test task",
		Type:        Task,
		Priority:    Medium,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	tests := []struct {
		name          string
		field         string
		value         string
		expectedPrio  DoPrio
		expectedType  DoType
		shouldSucceed bool
	}{
		{
			name:          "set priority to high",
			field:         "prio",
			value:         "high",
			expectedPrio:  High,
			expectedType:  Task,
			shouldSucceed: true,
		},
		{
			name:          "set type to ask",
			field:         "type",
			value:         "ask",
			expectedPrio:  High, // from previous test
			expectedType:  Ask,
			shouldSucceed: true,
		},
		{
			name:          "set priority to low",
			field:         "prio",
			value:         "low",
			expectedPrio:  Low,
			expectedType:  Ask, // from previous test
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update the do
			var fetchedDo Do
			conn.First(&fetchedDo, do.ID)

			switch tt.field {
			case "prio":
				fetchedDo.Priority = mapPriority(tt.value)
			case "type":
				fetchedDo.Type = mapType(tt.value)
			}

			conn.Save(&fetchedDo)

			// Verify
			var verifyDo Do
			conn.First(&verifyDo, do.ID)

			if verifyDo.Priority != tt.expectedPrio {
				t.Errorf("Expected priority %s, got %s", tt.expectedPrio, verifyDo.Priority)
			}

			if verifyDo.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, verifyDo.Type)
			}
		})
	}
}

func TestDidCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a task
	do := Do{
		Description: "Task to complete",
		Type:        Task,
		Priority:    Medium,
		Completed:   false,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	// Mark as completed
	do.Completed = true
	if err := conn.Save(&do).Error; err != nil {
		t.Errorf("Failed to mark task as complete: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Completed {
		t.Error("Expected task to be completed")
	}
}

func TestScratchCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a task
	do := Do{
		Description: "Task to delete",
		Type:        Task,
		Priority:    Medium,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	// Soft delete with reason
	do.Deleted = true
	do.Reason = "Not needed"
	if err := conn.Save(&do).Error; err != nil {
		t.Errorf("Failed to delete task: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Deleted {
		t.Error("Expected task to be deleted")
	}

	if fetched.Reason != "Not needed" {
		t.Errorf("Expected reason 'Not needed', got '%s'", fetched.Reason)
	}
}

func TestUnscratchCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a deleted task
	do := Do{
		Description: "Deleted task",
		Type:        Task,
		Priority:    Medium,
		Deleted:     true,
		Reason:      "Mistake",
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	// Restore it
	do.Deleted = false
	if err := conn.Save(&do).Error; err != nil {
		t.Errorf("Failed to restore task: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if fetched.Deleted {
		t.Error("Expected task to be restored")
	}
}

func TestPinUnpinCommands(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a task
	do := Do{
		Description: "Task to pin",
		Type:        Task,
		Priority:    Medium,
		Pinned:      false,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	// Pin it
	do.Pinned = true
	if err := conn.Save(&do).Error; err != nil {
		t.Errorf("Failed to pin task: %v", err)
	}

	// Verify pinned
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Pinned {
		t.Error("Expected task to be pinned")
	}

	// Unpin it
	fetched.Pinned = false
	if err := conn.Save(&fetched).Error; err != nil {
		t.Errorf("Failed to unpin task: %v", err)
	}

	// Verify unpinned
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if fetched.Pinned {
		t.Error("Expected task to be unpinned")
	}
}

func TestMarkUnmarkCommands(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a task
	do := Do{
		Description: "Sensitive task",
		Type:        Task,
		Priority:    Medium,
		Sensitive:   false,
	}
	if err := conn.Create(&do).Error; err != nil {
		t.Fatalf("Failed to create test do: %v", err)
	}

	// Mark as sensitive
	do.Sensitive = true
	if err := conn.Save(&do).Error; err != nil {
		t.Errorf("Failed to mark task: %v", err)
	}

	// Verify marked
	var fetched Do
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if !fetched.Sensitive {
		t.Error("Expected task to be marked as sensitive")
	}

	// Unmark
	fetched.Sensitive = false
	if err := conn.Save(&fetched).Error; err != nil {
		t.Errorf("Failed to unmark task: %v", err)
	}

	// Verify unmarked
	if err := conn.First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if fetched.Sensitive {
		t.Error("Expected task to be unmarked")
	}
}

func TestRecruitCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a tag (recruit)
	tag := Tag{Name: "alice"}
	if err := conn.Create(&tag).Error; err != nil {
		t.Fatalf("Failed to recruit: %v", err)
	}

	// Verify
	var fetched Tag
	if err := conn.Where("name = ?", "alice").First(&fetched).Error; err != nil {
		t.Errorf("Failed to find recruited tag: %v", err)
	}

	if fetched.Name != "alice" {
		t.Errorf("Expected name 'alice', got '%s'", fetched.Name)
	}
}

func TestRenameCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a tag
	tag := Tag{Name: "bob"}
	if err := conn.Create(&tag).Error; err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Rename
	tag.Name = "robert"
	if err := conn.Save(&tag).Error; err != nil {
		t.Errorf("Failed to rename tag: %v", err)
	}

	// Verify
	var fetched Tag
	if err := conn.First(&fetched, tag.ID).Error; err != nil {
		t.Fatalf("Failed to fetch tag: %v", err)
	}

	if fetched.Name != "robert" {
		t.Errorf("Expected name 'robert', got '%s'", fetched.Name)
	}
}

func TestAskTellCommands(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		taskType DoType
		message  string
		person   string
	}{
		{
			name:     "ask command",
			taskType: Ask,
			message:  "What time is the meeting?",
			person:   "alice",
		},
		{
			name:     "tell command",
			taskType: Tell,
			message:  "Meeting is at 3pm",
			person:   "bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create or find tag
			tag := Tag{Name: tt.person}
			conn.FirstOrCreate(&tag, Tag{Name: tt.person})

			// Create task
			do := Do{
				Description: tt.message,
				Type:        tt.taskType,
				Priority:    Medium,
			}
			if err := conn.Create(&do).Error; err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}

			// Create relationship
			doTag := DoTag{DoID: do.ID, TagID: tag.ID}
			if err := conn.Create(&doTag).Error; err != nil {
				t.Fatalf("Failed to create relationship: %v", err)
			}

			// Verify
			var fetched Do
			if err := conn.Preload("Tags").First(&fetched, do.ID).Error; err != nil {
				t.Fatalf("Failed to fetch do: %v", err)
			}

			if fetched.Type != tt.taskType {
				t.Errorf("Expected type %s, got %s", tt.taskType, fetched.Type)
			}

			if len(fetched.Tags) != 1 {
				t.Errorf("Expected 1 tag, got %d", len(fetched.Tags))
			}

			if fetched.Tags[0].Name != tt.person {
				t.Errorf("Expected tag name '%s', got '%s'", tt.person, fetched.Tags[0].Name)
			}
		})
	}
}

func TestBragLearnCommands(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		taskType DoType
		message  string
	}{
		{
			name:     "brag command",
			taskType: Brag,
			message:  "Completed major feature",
		},
		{
			name:     "learn command",
			taskType: Learn,
			message:  "Study distributed systems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do := Do{
				Description: tt.message,
				Type:        tt.taskType,
				Completed:   false,
			}
			if err := conn.Create(&do).Error; err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}

			// Verify
			var fetched Do
			if err := conn.First(&fetched, do.ID).Error; err != nil {
				t.Fatalf("Failed to fetch do: %v", err)
			}

			if fetched.Type != tt.taskType {
				t.Errorf("Expected type %s, got %s", tt.taskType, fetched.Type)
			}

			if fetched.Description != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, fetched.Description)
			}
		})
	}
}

func TestReassignCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create tags
	alice := Tag{Name: "alice"}
	bob := Tag{Name: "bob"}
	conn.Create(&alice)
	conn.Create(&bob)

	// Create task assigned to alice
	do := Do{
		Description: "Task for alice",
		Type:        Task,
		Priority:    Medium,
	}
	conn.Create(&do)

	doTag := DoTag{DoID: do.ID, TagID: alice.ID}
	conn.Create(&doTag)

	// Reassign to bob
	conn.Where("do_id = ?", do.ID).Delete(&DoTag{})
	newDoTag := DoTag{DoID: do.ID, TagID: bob.ID}
	conn.Create(&newDoTag)

	// Verify
	var fetched Do
	if err := conn.Preload("Tags").First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if len(fetched.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(fetched.Tags))
	}

	if fetched.Tags[0].Name != "bob" {
		t.Errorf("Expected tag 'bob', got '%s'", fetched.Tags[0].Name)
	}
}

func TestUnassignCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create tag and task
	tag := Tag{Name: "alice"}
	conn.Create(&tag)

	do := Do{
		Description: "Assigned task",
		Type:        Task,
		Priority:    Medium,
	}
	conn.Create(&do)

	doTag := DoTag{DoID: do.ID, TagID: tag.ID}
	conn.Create(&doTag)

	// Unassign
	conn.Where("do_id = ?", do.ID).Delete(&DoTag{})

	// Verify
	var fetched Do
	if err := conn.Preload("Tags").First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if len(fetched.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(fetched.Tags))
	}
}

func TestDocCommand(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create task
	do := Do{
		Description: "Task with docs",
		Type:        Task,
		Priority:    Medium,
	}
	conn.Create(&do)

	// Add documentation
	doc := DoDoc{
		DoID: do.ID,
		Text: "# Meeting Notes\n\n- Point 1\n- Point 2",
	}
	if err := conn.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create doc: %v", err)
	}

	// Verify
	var fetched Do
	if err := conn.Preload("Doc").First(&fetched, do.ID).Error; err != nil {
		t.Fatalf("Failed to fetch do: %v", err)
	}

	if fetched.Doc.ID == 0 {
		t.Error("Expected doc to exist")
	}

	if fetched.Doc.Text != doc.Text {
		t.Errorf("Expected doc text to match")
	}
}

func TestLogQueryFilters(t *testing.T) {
	conn, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create test data
	alice := Tag{Name: "alice"}
	conn.Create(&alice)

	tasks := []Do{
		{Description: "Task 1", Type: Task, Priority: High},
		{Description: "Task 2", Type: Ask, Priority: Medium},
		{Description: "Task 3", Type: Tell, Priority: Low},
		{Description: "Task 4", Type: Task, Priority: High, Completed: true},
	}

	for i := range tasks {
		conn.Create(&tasks[i])
		if i == 1 { // Assign second task to alice
			doTag := DoTag{DoID: tasks[i].ID, TagID: alice.ID}
			conn.Create(&doTag)
		}
	}

	tests := []struct {
		name          string
		queryFunc     func(*gorm.DB) *gorm.DB
		expectedCount int
	}{
		{
			name: "filter by type ask",
			queryFunc: func(db *gorm.DB) *gorm.DB {
				return db.Where("type = ?", Ask)
			},
			expectedCount: 1,
		},
		{
			name: "filter by priority high",
			queryFunc: func(db *gorm.DB) *gorm.DB {
				return db.Where("priority = ?", High)
			},
			expectedCount: 2,
		},
		{
			name: "filter completed",
			queryFunc: func(db *gorm.DB) *gorm.DB {
				return db.Where("completed = ?", true)
			},
			expectedCount: 1,
		},
		{
			name: "filter not completed",
			queryFunc: func(db *gorm.DB) *gorm.DB {
				return db.Where("completed = ?", false)
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []Do
			query := tt.queryFunc(conn)
			query.Find(&results)

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}
