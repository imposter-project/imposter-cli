package platform

import (
	"runtime"
	"testing"
)

func TestGetPlatform(t *testing.T) {
	// Test that GetPlatform returns the current runtime OS and architecture
	gotOS, gotArch := GetPlatform()

	// Verify OS matches runtime.GOOS
	expectedOS := runtime.GOOS
	if gotOS != expectedOS {
		t.Errorf("GetPlatform() OS = %v, want %v", gotOS, expectedOS)
	}

	// Verify architecture matches runtime.GOARCH
	expectedArch := runtime.GOARCH
	if gotArch != expectedArch {
		t.Errorf("GetPlatform() arch = %v, want %v", gotArch, expectedArch)
	}

	// Verify non-empty values
	if gotOS == "" {
		t.Error("GetPlatform() returned empty OS")
	}

	if gotArch == "" {
		t.Error("GetPlatform() returned empty architecture")
	}
}
