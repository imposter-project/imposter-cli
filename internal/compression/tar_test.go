package compression

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestTarGz(t *testing.T, testDir string) string {
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

	// Create tar.gz archive
	tarGzFile := filepath.Join(testDir, "test.tar.gz")
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add files to tar
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
			hdr := &tar.Header{
				Name:     f.path,
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
		} else {
			hdr := &tar.Header{
				Name: f.path,
				Mode: 0644,
				Size: int64(len(f.content)),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write([]byte(f.content)); err != nil {
				t.Fatal(err)
			}
		}
	}

	return tarGzFile
}

func TestExtractTarGz(t *testing.T) {
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

	// Create test tar.gz file
	tarGzFile := createTestTarGz(t, sourceDir)

	// Extract to destination
	destDir := filepath.Join(tempDir, "dest")
	if err := ExtractTarGz(tarGzFile, destDir); err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
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

func TestExtractTarGzInvalidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	destDir := filepath.Join(tempDir, "dest")
	err = ExtractTarGz("nonexistent.tar.gz", destDir)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestExtractTarGzDirectoryTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compression_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create malicious tar.gz with directory traversal
	tarGzFile := filepath.Join(tempDir, "malicious.tar.gz")
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add malicious file with directory traversal
	hdr := &tar.Header{
		Name: "../../../foo/bar",
		Mode: 0644,
		Size: 4,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tempDir, "dest")
	err = ExtractTarGz(tarGzFile, destDir)
	if err == nil {
		t.Error("Expected error for directory traversal attempt")
	} else if !strings.Contains(err.Error(), "invalid file path") && !strings.Contains(err.Error(), "unexpected EOF") {
		t.Errorf("Expected directory traversal or EOF error, got: %v", err)
	}
}
