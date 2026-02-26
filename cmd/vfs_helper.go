package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/s3bw/vfs"
	"gorm.io/gorm"
)

// VFSManager wraps VFS functionality for captain
type VFSManager struct {
	vfs      *vfs.VFS
	storage  *vfs.GormStorage
	filesDir string
}

// NewVFSManager creates a new VFS manager for captain
func NewVFSManager(db *gorm.DB, captainDir string) (*VFSManager, error) {
	filesDir := filepath.Join(captainDir, "files")

	storage := vfs.NewGormStorage(db, filesDir)
	vfsTree, err := storage.LoadVFSFromDB()
	if err != nil {
		return nil, err
	}

	return &VFSManager{
		vfs:      vfsTree,
		storage:  storage,
		filesDir: filesDir,
	}, nil
}

// CreatePromotedFile creates a .do file with task content
func (vm *VFSManager) CreatePromotedFile(filename, description, docContent string) error {
	// Ensure .do extension
	if !strings.HasSuffix(filename, ".do") {
		filename += ".do"
	}

	// Create the file node at root
	node, err := vm.vfs.CreateFile(filename, false)
	if err != nil {
		return err
	}

	// Format content: description as header + doc content as body
	content := fmt.Sprintf("# %s\n\n%s", description, docContent)

	// Set file content using VFS layer (updates node size)
	return vm.vfs.SetFileContent(node, content)
}

// Save persists all VFS changes to database
func (vm *VFSManager) Save() error {
	if err := vm.storage.SaveVFSToDB(vm.vfs); err != nil {
		return err
	}
	return vm.storage.SaveDirectoryStates(vm.vfs)
}
