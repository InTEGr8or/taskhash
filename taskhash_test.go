package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single element",
			input:    []string{"x"},
			expected: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqueStrings(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("uniqueStrings() got %d elements, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("uniqueStrings()[%d] = %s, want %s", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "file exists",
			path:     existingFile,
			expected: true,
		},
		{
			name:     "file does not exist",
			path:     filepath.Join(tmpDir, "nonexistent.txt"),
			expected: false,
		},
		{
			name:     "directory",
			path:     tmpDir,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileExists(tt.path)
			if result != tt.expected {
				t.Errorf("fileExists(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCalculateSignature(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	includes := []string{tmpDir}
	excludes := []string{}

	sig, err := calculateSignature(includes, excludes)
	if err != nil {
		t.Fatalf("calculateSignature() error = %v", err)
	}

	if len(sig) != 32 {
		t.Errorf("calculateSignature() = %s, want 32 character MD5 hash", sig)
	}

	sig2, err := calculateSignature(includes, excludes)
	if err != nil {
		t.Fatalf("calculateSignature() second call error = %v", err)
	}

	if sig != sig2 {
		t.Errorf("calculateSignature() not deterministic: got %s and %s", sig, sig2)
	}

	newContent := []byte("modified content")
	if err := os.WriteFile(file1, newContent, 0644); err != nil {
		t.Fatal(err)
	}

	sig3, err := calculateSignature(includes, excludes)
	if err != nil {
		t.Fatalf("calculateSignature() after modification error = %v", err)
	}

	if sig == sig3 {
		t.Errorf("calculateSignature() should detect file change: same hash %s", sig)
	}
}

func TestCalculateSignatureWithExcludes(t *testing.T) {
	tmpDir := t.TempDir()

	includeFile := filepath.Join(tmpDir, "include.txt")
	excludeDir := filepath.Join(tmpDir, "exclude")
	excludeFile := filepath.Join(excludeDir, "excluded.txt")

	if err := os.MkdirAll(excludeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(includeFile, []byte("include me"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(excludeFile, []byte("exclude me"), 0644); err != nil {
		t.Fatal(err)
	}

	includes := []string{tmpDir}
	excludes := []string{"exclude"}

	sig, err := calculateSignature(includes, excludes)
	if err != nil {
		t.Fatalf("calculateSignature() error = %v", err)
	}

	if sig == "" {
		t.Error("calculateSignature() returned empty hash")
	}
}

func TestLoadSaveSignatures(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "signatures.json")

	sigs := map[string]string{
		"lint":  "abc123",
		"test":  "def456",
		"build": "",
	}

	if err := saveSignatures(storePath, sigs); err != nil {
		t.Fatalf("saveSignatures() error = %v", err)
	}

	loaded, err := loadSignatures(storePath)
	if err != nil {
		t.Fatalf("loadSignatures() error = %v", err)
	}

	if len(loaded) != len(sigs) {
		t.Errorf("loadSignatures() got %d signatures, want %d", len(loaded), len(sigs))
	}

	for key, expected := range sigs {
		if loaded[key] != expected {
			t.Errorf("loadSignatures()[%s] = %s, want %s", key, loaded[key], expected)
		}
	}
}

func TestLoadSignaturesNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "nonexistent.json")

	sigs, err := loadSignatures(storePath)
	if err != nil {
		t.Fatalf("loadSignatures() for non-existent file error = %v", err)
	}

	if len(sigs) != 0 {
		t.Errorf("loadSignatures() for non-existent file got %d signatures, want 0", len(sigs))
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "taskhash.json")

	config := &Config{
		Includes: []string{"src", "lib"},
		Excludes: []string{".git", "node_modules"},
		Store:    ".code_signatures.json",
		Runner:   "make",
		Tasks: map[string]Task{
			"lint":  {},
			"test":  {Includes: []string{"src", "tests"}},
			"build": {},
		},
	}

	if err := saveConfig(config, configPath); err != nil {
		t.Fatalf("saveConfig() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	config2, err := loadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromPath() error = %v", err)
	}

	if len(config2.Includes) != len(config.Includes) {
		t.Errorf("Includes count: got %d, want %d", len(config2.Includes), len(config.Includes))
	}

	if config2.Runner != config.Runner {
		t.Errorf("Runner: got %s, want %s", config2.Runner, config.Runner)
	}

	if len(config2.Excludes) != len(config.Excludes) {
		t.Errorf("Excludes count: got %d, want %d", len(config2.Excludes), len(config.Excludes))
	}

	_ = data
}

func TestConfigGetIncludesForTask(t *testing.T) {
	config := &Config{
		Includes: []string{"global"},
		Tasks: map[string]Task{
			"lint":  {},
			"test":  {Includes: []string{"src", "tests"}},
			"build": {},
		},
	}

	tests := []struct {
		taskName string
		expected []string
	}{
		{"lint", []string{"global"}},
		{"test", []string{"src", "tests"}},
		{"build", []string{"global"}},
	}

	for _, tt := range tests {
		t.Run(tt.taskName, func(t *testing.T) {
			result := config.GetIncludesForTask(tt.taskName)
			if len(result) != len(tt.expected) {
				t.Errorf("GetIncludesForTask(%s) returned %d items, want %d", tt.taskName, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetIncludesForTask(%s)[%d] = %s, want %s", tt.taskName, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestConfigGetCommandForTask(t *testing.T) {
	config := &Config{
		Runner: "make",
	}

	tests := []struct {
		taskName string
		expected string
	}{
		{"lint", "make lint"},
		{"test", "make test"},
		{"build", "make build"},
		{"custom-task", "make custom-task"},
	}

	for _, tt := range tests {
		t.Run(tt.taskName, func(t *testing.T) {
			result := config.GetCommandForTask(tt.taskName)
			if result != tt.expected {
				t.Errorf("GetCommandForTask(%s) = %s, want %s", tt.taskName, result, tt.expected)
			}
		})
	}
}

func loadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.Store == "" {
		config.Store = ".code_signatures.json"
	}
	if config.Runner == "" {
		config.Runner = "make"
	}
	if config.Tasks == nil {
		config.Tasks = map[string]Task{
			"lint": {},
			"test": {},
		}
	}
	return &config, nil
}
