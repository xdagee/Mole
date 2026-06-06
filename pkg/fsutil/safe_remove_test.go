package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsProtectedPath(t *testing.T) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}
	appData := os.Getenv("APPDATA")

	tests := []struct {
		path     string
		expected bool
	}{
		{filepath.Join(sysDrive, "\\"), true},
		{filepath.Join(sysDrive, "\\Windows"), true},
		{filepath.Join(sysDrive, "\\Windows", "System32"), true},
		{filepath.Join(sysDrive, "\\Program Files"), true},
		{filepath.Join(sysDrive, "\\Program Files (x86)"), true},
		{filepath.Join(sysDrive, "\\ProgramData", "Microsoft"), true},
		{filepath.Join(sysDrive, "\\Users", "Default"), true},
		{filepath.Join(sysDrive, "\\Users", "someuser", "Desktop"), false},
		{filepath.Join(sysDrive, "\\Temp", "mole_test"), false},
	}

	if appData != "" {
		tests = append(tests, struct{path string; expected bool}{
			filepath.Join(appData, "Microsoft", "Windows"), true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsProtectedPath(tt.path); got != tt.expected {
				t.Errorf("IsProtectedPath(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestSafeRemove(t *testing.T) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = "C:"
	}

	protected := filepath.Join(sysDrive, "\\Windows")
	err := SafeRemove(protected, false)
	if err == nil {
		t.Errorf("Expected error when deleting protected path %s, got nil", protected)
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Test dry run
	err = SafeRemove(testFile, true)
	if err != nil {
		t.Errorf("Dry run failed: %v", err)
	}
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("File was deleted during dry run!")
	}

	// Test actual delete
	err = SafeRemove(testFile, false)
	if err != nil {
		t.Errorf("Actual delete failed: %v", err)
	}
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File was not deleted!")
	}
}
