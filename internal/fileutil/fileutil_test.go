package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFilesWithExtension(t *testing.T) {
	testDir := "testdata"

	tests := []struct {
		name string
		ext  []string
		want []string
	}{
		{
			name: "find json files",
			ext:  []string{".json"},
			want: []string{"test1.json", "test5.json"},
		},
		{
			name: "find yaml files",
			ext:  []string{".yaml", ".yml"},
			want: []string{"test2.yaml", "test3.yml"},
		},
		{
			name: "find non-existent extension",
			ext:  []string{".missing"},
			want: []string{},
		},
		{
			name: "find multiple extensions",
			ext:  []string{".json", ".txt"},
			want: []string{"test1.json", "test4.txt", "test5.json", ".hidden.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindFilesWithExtension(testDir, tt.ext...)
			if len(got) != len(tt.want) {
				t.Errorf("FindFilesWithExtension() got %d files, want %d files", len(got), len(tt.want))
			}
			for _, wantFile := range tt.want {
				found := false
				for _, gotFile := range got {
					if gotFile == wantFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindFilesWithExtension() missing expected file %q", wantFile)
				}
			}
			for _, gotFile := range got {
				found := false
				for _, wantFile := range tt.want {
					if gotFile == wantFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindFilesWithExtension() found unexpected file %q", gotFile)
				}
			}
		})
	}
}

func TestGenerateFilePathAdjacentToFile(t *testing.T) {
	// Create temporary directory for output files
	tempDir, err := os.MkdirTemp("", "fileutil_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Copy test files to temp directory
	testFiles := []string{"test1.json", "test2.yaml"}
	for _, file := range testFiles {
		srcPath := filepath.Join("testdata", file)
		destPath := filepath.Join(tempDir, file)
		if err := CopyFile(srcPath, destPath); err != nil {
			t.Fatalf("Failed to copy test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name           string
		sourceFilePath string
		suffix         string
		forceOverwrite bool
		wantContains   string
	}{
		{
			name:           "simple suffix",
			sourceFilePath: filepath.Join(tempDir, "test1.json"),
			suffix:         "-modified",
			forceOverwrite: false,
			wantContains:   "test1-modified",
		},
		{
			name:           "with extension in suffix",
			sourceFilePath: filepath.Join(tempDir, "test1.json"),
			suffix:         ".yaml",
			forceOverwrite: false,
			wantContains:   "test1.yaml",
		},
		{
			name:           "with directory path",
			sourceFilePath: filepath.Join(tempDir, "test2.yaml"),
			suffix:         "-new",
			forceOverwrite: false,
			wantContains:   "test2-new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilePathAdjacentToFile(tt.sourceFilePath, tt.suffix, tt.forceOverwrite)
			if got == tt.sourceFilePath {
				t.Errorf("GenerateFilePathAdjacentToFile() returned source path unchanged")
			}
			if !contains(got, tt.wantContains) {
				t.Errorf("GenerateFilePathAdjacentToFile() = %v, want containing %v", got, tt.wantContains)
			}
		})
	}
}

func TestCopyDirShallow(t *testing.T) {
	// Create temporary destination directory
	destDir, err := os.MkdirTemp("", "fileutil_test_dest_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Test copying from testdata
	if err := CopyDirShallow("testdata", destDir); err != nil {
		t.Fatalf("CopyDirShallow() error = %v", err)
	}

	// Verify copied files
	expectedFiles := []string{"test1.json", "test2.yaml", "test3.yml", "test4.txt", "test5.json", ".hidden.txt"}
	for _, name := range expectedFiles {
		srcPath := filepath.Join("testdata", name)
		destPath := filepath.Join(destDir, name)

		srcContent, err := os.ReadFile(srcPath)
		if err != nil {
			t.Errorf("Failed to read source file %s: %v", name, err)
			continue
		}

		destContent, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", name, err)
			continue
		}

		if string(destContent) != string(srcContent) {
			t.Errorf("File %s content mismatch. Got %v, want %v", name, string(destContent), string(srcContent))
		}
	}
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name          string
		includeHidden bool
		want          int
	}{
		{
			name:          "exclude hidden files",
			includeHidden: false,
			want:          5, // All non-hidden files
		},
		{
			name:          "include hidden files",
			includeHidden: true,
			want:          6, // All files including hidden
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListFiles("testdata", tt.includeHidden)
			if err != nil {
				t.Fatalf("ListFiles() error = %v", err)
			}
			if len(got) != tt.want {
				t.Errorf("ListFiles() = got %v files, want %v files", len(got), tt.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		wantErr  bool
	}{
		{
			name:     "existing file",
			filePath: filepath.Join("testdata", "test1.json"),
			want:     `{"key": "value1"}`,
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filePath: filepath.Join("testdata", "non_existent.txt"),
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(*got) != tt.want {
				t.Errorf("ReadFile() = %v, want %v", string(*got), tt.want)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return filepath.Base(s) == substr
}
