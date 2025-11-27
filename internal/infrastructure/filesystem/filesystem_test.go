package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileWriter_WriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writer := NewFileWriter()

	tests := []struct {
		name    string
		path    string
		content []byte
		wantErr bool
	}{
		{
			name:    "Write simple file",
			path:    filepath.Join(tmpDir, "test.txt"),
			content: []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "Write nested file (create dirs)",
			path:    filepath.Join(tmpDir, "nested", "dir", "test.txt"),
			content: []byte("nested content"),
			wantErr: false,
		},
		{
			name:    "Empty path error",
			path:    "",
			content: []byte("content"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.WriteFile(tt.path, tt.content, 0644)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file content
				content, err := os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("failed to read written file: %v", err)
				}
				if string(content) != string(tt.content) {
					t.Errorf("file content = %s, want %s", string(content), string(tt.content))
				}
			}
		})
	}
}

func TestDirectoryManager_CreateDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dir_manager_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewDirectoryManager()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Create simple dir",
			path:    filepath.Join(tmpDir, "newdir"),
			wantErr: false,
		},
		{
			name:    "Create nested dir",
			path:    filepath.Join(tmpDir, "a", "b", "c"),
			wantErr: false,
		},
		{
			name:    "Empty path error",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateDir(tt.path, 0755)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				info, err := os.Stat(tt.path)
				if err != nil {
					t.Errorf("directory was not created: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("path is not a directory")
				}
			}
		})
	}
}

func TestDirectoryManager_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exists_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewDirectoryManager()
	
	// Create a file to test existence
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "Existing directory",
			path:    tmpDir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Existing file",
			path:    testFile,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Non-existent path",
			path:    filepath.Join(tmpDir, "ghost"),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Exists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
