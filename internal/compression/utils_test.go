package compression

import "testing"

func TestIsArchiveFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "zip file",
			filename: "test.zip",
			expected: true,
		},
		{
			name:     "tar.gz file",
			filename: "test.tar.gz",
			expected: true,
		},
		{
			name:     "zip file with path",
			filename: "/path/to/archive.zip",
			expected: true,
		},
		{
			name:     "tar.gz file with path",
			filename: "/path/to/archive.tar.gz",
			expected: true,
		},
		{
			name:     "text file",
			filename: "test.txt",
			expected: false,
		},
		{
			name:     "executable file",
			filename: "program.exe",
			expected: false,
		},
		{
			name:     "json file",
			filename: "config.json",
			expected: false,
		},
		{
			name:     "tar file without gzip",
			filename: "archive.tar",
			expected: false,
		},
		{
			name:     "gz file without tar",
			filename: "file.gz",
			expected: false,
		},
		{
			name:     "empty filename",
			filename: "",
			expected: false,
		},
		{
			name:     "filename with zip in middle",
			filename: "test.zip.txt",
			expected: false,
		},
		{
			name:     "filename with tar.gz in middle",
			filename: "test.tar.gz.backup",
			expected: false,
		},
		{
			name:     "case sensitivity - ZIP uppercase",
			filename: "test.ZIP",
			expected: false,
		},
		{
			name:     "case sensitivity - TAR.GZ uppercase",
			filename: "test.TAR.GZ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsArchiveFileExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("IsArchiveFileExtension(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}
