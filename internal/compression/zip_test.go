package compression

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestZip(t *testing.T, testDir string) string {
	// Create test files
	testFile1 := filepath.Join(testDir, "test1.txt")
	testFile2 := filepath.Join(testDir, "subdir", "test2.txt")

	if err := os.MkdirAll(filepath.Dir(testFile2), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create zip archive
	zipFile := filepath.Join(testDir, "test.zip")
	file, err := os.Create(zipFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	// Add files to zip
	files := []struct {
		path    string
		content string
		isDir   bool
	}{
		{"test1.txt", "content1", false},
		{"subdir/", "", true},
		{"subdir/test2.txt", "content2", false},
	}

	for _, f := range files {
		if f.isDir {
			// Create directory entry in zip
			header := &zip.FileHeader{
				Name: f.path,
			}
			header.SetMode(0755 | os.ModeDir)
			_, err := w.CreateHeader(header)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			writer, err := w.Create(f.path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Write([]byte(f.content)); err != nil {
				t.Fatal(err)
			}
		}
	}

	return zipFile
}

func TestExtractZip(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test zip file
	zipFile := createTestZip(t, sourceDir)

	// Extract to destination
	destDir := filepath.Join(tempDir, "dest")
	if err := ExtractZip(zipFile, destDir); err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Verify extracted files
	test1Path := filepath.Join(destDir, "test1.txt")
	content1, err := os.ReadFile(test1Path)
	if err != nil {
		t.Fatalf("Failed to read test1.txt: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("Expected 'content1', got '%s'", string(content1))
	}

	test2Path := filepath.Join(destDir, "subdir", "test2.txt")
	content2, err := os.ReadFile(test2Path)
	if err != nil {
		t.Fatalf("Failed to read subdir/test2.txt: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("Expected 'content2', got '%s'", string(content2))
	}

	// Verify directory exists
	subdirPath := filepath.Join(destDir, "subdir")
	if stat, err := os.Stat(subdirPath); err != nil {
		t.Fatalf("Subdir doesn't exist: %v", err)
	} else if !stat.IsDir() {
		t.Error("Subdir is not a directory")
	}
}

func TestExtractZipInvalidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	destDir := filepath.Join(tempDir, "dest")
	err = ExtractZip("nonexistent.zip", destDir)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestExtractZipDirectoryTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create malicious zip with directory traversal
	zipFile := filepath.Join(tempDir, "malicious.zip")
	file, err := os.Create(zipFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	// Add malicious file with directory traversal
	writer, err := w.Create("../../../foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}

	// Close the zip writer to finalize the file
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tempDir, "dest")
	err = ExtractZip(zipFile, destDir)
	if err == nil || !strings.Contains(err.Error(), "invalid file path") {
		t.Errorf("Expected directory traversal error, got: %v", err)
	}
}
