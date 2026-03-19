package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// =============================================================================
// Output Formatting
// =============================================================================

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	italic = "\033[3m"

	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"

	gray = "\033[90m"
)

func color(c string) string {
	if runtime.GOOS == "windows" {
		return ""
	}
	return c
}

func greenBold(s string) string  { return color(bold+green) + s + color(reset) }
func redBold(s string) string    { return color(bold+red) + s + color(reset) }
func yellowBold(s string) string { return color(bold+yellow) + s + color(reset) }
func cyanBold(s string) string   { return color(bold+cyan) + s + color(reset) }
func dimText(s string) string    { return color(dim) + s + color(reset) }
func boldText(s string) string   { return color(bold) + s + color(reset) }

func success(s string) { fmt.Println(greenBold("✓") + " " + s) }
func failure(s string) { fmt.Println(redBold("✗") + " " + s) }
func skipped(s string) { fmt.Println(dimText("○") + " " + dimText(s)) }
func running(s string) { fmt.Print(cyanBold("→") + " " + s + "... ") }

func printRow(cols ...string) {
	fmt.Println(strings.Join(cols, "  "))
}

func printHeader(cells ...string) {
	fmt.Println()
	for _, cell := range cells {
		fmt.Print(boldText(cell) + "  ")
	}
	fmt.Println()
}

func printKeyValue(key, value string) {
	padding := 30
	if len(key)+len(value) >= padding {
		fmt.Printf("%s %s\n", boldText(key), value)
	} else {
		dots := strings.Repeat(".", padding-len(key)-len(value))
		fmt.Printf("%s %s%s%s\n", boldText(key), dots, color(dim), value)
	}
}

type Config struct {
	Includes []string        `json:"includes"`
	Excludes []string        `json:"excludes"`
	Store    string          `json:"store"`
	Runner   string          `json:"runner"`
	Tasks    map[string]Task `json:"tasks"`
}

type Task struct {
	Includes []string `json:"includes,omitempty"`
}

func (c *Config) GetIncludesForTask(taskName string) []string {
	if task, ok := c.Tasks[taskName]; ok && len(task.Includes) > 0 {
		return task.Includes
	}
	return c.Includes
}

func (c *Config) GetCommandForTask(taskName string) string {
	return fmt.Sprintf("%s %s", c.Runner, taskName)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		runInit()
	case "enforce":
		runEnforce()
	case "up":
		runUpdateBinary()
	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "%s Error: task name required\n", redBold("✗"))
			fmt.Fprintf(os.Stderr, "Usage: taskhash check %s<task>%s\n", cyanBold(""), reset)
			os.Exit(1)
		}
		runCheck(os.Args[2])
	case "update":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "%s Error: task name required\n", redBold("✗"))
			fmt.Fprintf(os.Stderr, "Usage: taskhash update %s<task>%s\n", cyanBold(""), reset)
			os.Exit(1)
		}
		runUpdate(os.Args[2])
	case "install-hook":
		runInstallHook()
	case "remove-hook":
		runRemoveHook()
	case "reinstall-hook":
		runRemoveHook()
		runInstallHook()
	case "--help", "-h", "help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown command: %s\n", redBold("✗"), cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(boldText("taskhash") + " " + dimText("Code signature hashing for task optimization"))
	fmt.Println()
	fmt.Println(boldText("Usage:"))
	fmt.Println("  " + cyanBold("taskhash") + " <command> [options]")
	fmt.Println()
	fmt.Println(boldText("Commands:"))
	fmt.Println("  init              Detect framework, create config, install hook")
	fmt.Println("  init --no-hook    Detect framework, create config only")
	fmt.Println("  enforce           Run all tasks (lint → test → ...)")
	fmt.Println("  up               Update taskhash binary")
	fmt.Println("  check <task>     Check if task hash matches")
	fmt.Println("  update <task>    Update task hash manually")
	fmt.Println("  install-hook     Install pre-commit hook")
	fmt.Println("  remove-hook      Remove pre-commit hook")
	fmt.Println("  reinstall-hook    Reinstall pre-commit hook")
	fmt.Println("  help              Show this help")
	fmt.Println()
	fmt.Println(boldText("Configuration:"))
	fmt.Println("  Runner: Command prefix (default: make)")
	fmt.Println("  Tasks: Each task runs \"{runner} {task}\"")
	fmt.Println("  Includes: Files/dirs to hash (per-task overrides supported)")
}

func loadConfig() (*Config, error) {
	path := "taskhash.json"
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file '%s' not found. Run 'taskhash init' first", path)
		}
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

func saveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadSignatures(path string) (map[string]string, error) {
	sigs := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return sigs, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, &sigs)
	return sigs, err
}

func saveSignatures(path string, sigs map[string]string) error {
	data, err := json.MarshalIndent(sigs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func calculateSignature(includes []string, excludes []string) (string, error) {
	hasher := md5.New()
	var files []string

	for _, include := range includes {
		err := filepath.Walk(include, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			rel, _ := filepath.Rel(".", path)
			for _, ex := range excludes {
				if strings.Contains(rel, ex) {
					return nil
				}
			}

			files = append(files, path)
			return nil
		})
		if err != nil {
			return "", err
		}
	}

	sort.Strings(files)

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		if _, err := io.Copy(hasher, file); err != nil {
			file.Close()
			return "", err
		}
		file.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func runInit() {
	noHook := len(os.Args) >= 3 && os.Args[2] == "--no-hook"

	config := detectFramework()
	path := "taskhash.json"
	if err := saveConfig(config.Config, path); err != nil {
		fmt.Fprintf(os.Stderr, "%s Error writing config: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(boldText("taskhash") + dimText(" initialization"))
	fmt.Println(strings.Repeat("─", 50))
	printKeyValue("Framework", cyanBold(config.detectedFramework))
	printKeyValue("Runner", config.Config.Runner)
	printKeyValue("Includes", strings.Join(config.Config.Includes, ", "))

	taskNames := make([]string, 0, len(config.Config.Tasks))
	for name := range config.Config.Tasks {
		taskNames = append(taskNames, name)
	}
	sort.Strings(taskNames)
	printKeyValue("Tasks", strings.Join(taskNames, ", "))

	success("Config written to " + cyanBold(path))
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	if !noHook {
		runInstallHook()
	}
}

type initConfig struct {
	*Config
	detectedFramework string
}

func detectFramework() *initConfig {
	includes := []string{}
	runner := "make"
	detected := "generic"

	if fileExists("pyproject.toml") {
		detected = "Python (pyproject.toml)"
		includes = append(includes, "src", "tests", "pyproject.toml")
	}

	if fileExists("package.json") {
		detected = "JavaScript/TypeScript (package.json)"
		includes = append(includes, "src", "package.json")
	}

	if fileExists("go.mod") {
		detected = "Go (go.mod)"
		includes = append(includes, ".")
	}

	if fileExists("Cargo.toml") {
		detected = "Rust (Cargo.toml)"
		includes = append(includes, "src", "Cargo.toml")
	}

	if fileExists("Gemfile") {
		detected = "Ruby (Gemfile)"
		includes = append(includes, "lib", "Gemfile")
	}

	if len(includes) == 0 {
		detected = "generic project"
		includes = []string{"src"}
	}

	includes = uniqueStrings(includes)

	return &initConfig{
		Config: &Config{
			Includes: includes,
			Excludes: []string{
				".git",
				"__pycache__",
				"node_modules",
				"vendor",
				".code_signatures.json",
				"taskhash",
				"taskhash.exe",
				".venv",
				"venv",
				"dist",
				"build",
				"*.lock",
			},
			Store:  ".code_signatures.json",
			Runner: runner,
			Tasks: map[string]Task{
				"lint":  {},
				"test":  {},
				"build": {},
			},
		},
		detectedFramework: detected,
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func runEnforce() {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	signatures, err := loadSignatures(config.Store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error loading signatures: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	preferredOrder := []string{"lint", "test", "build"}

	orderedTasks := make([]string, 0, len(config.Tasks))
	seen := make(map[string]bool)

	for _, preferred := range preferredOrder {
		if _, ok := config.Tasks[preferred]; ok {
			orderedTasks = append(orderedTasks, preferred)
			seen[preferred] = true
		}
	}

	for name := range config.Tasks {
		if !seen[name] {
			orderedTasks = append(orderedTasks, name)
		}
	}

	fmt.Println()
	fmt.Println(boldText("taskhash") + dimText(" running tasks..."))
	fmt.Println(strings.Repeat("─", 50))

	failed := false
	for _, taskName := range orderedTasks {
		includes := config.GetIncludesForTask(taskName)
		currentSig, err := calculateSignature(includes, config.Excludes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error calculating signature for %s: %v\n", redBold("✗"), taskName, err)
			os.Exit(1)
		}

		prevSig := signatures[taskName]
		cmd := config.GetCommandForTask(taskName)

		shortCurrent := currentSig
		if len(shortCurrent) > 8 {
			shortCurrent = currentSig[:8]
		}

		if prevSig == currentSig {
			skipped(taskName + dimText(" hash matches"))
		} else {
			shortPrev := prevSig
			if len(shortPrev) > 8 {
				shortPrev = prevSig[:8]
			}
			if prevSig == "" {
				fmt.Printf("%s %s %s\n", cyanBold("→"), boldText(taskName), dimText("no hash, running..."))
			} else {
				fmt.Printf("%s %s %s %s\n", cyanBold("→"), boldText(taskName),
					dimText("changed"), dimText("("+shortPrev+" → "+shortCurrent+")"))
			}

			shell := "/bin/bash"
			if runtime.GOOS == "windows" {
				shell = "cmd.exe"
			}

			runCmd := exec.Command(shell, "-c", cmd)
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			runCmd.Stdin = os.Stdin

			if err := runCmd.Run(); err != nil {
				fmt.Printf("%s %s %s\n", redBold("✗"), boldText(taskName), redBold("failed"))
				failed = true
			} else {
				fmt.Printf("%s %s %s\n", greenBold("✓"), boldText(taskName), dimText("passed"))

				signatures[taskName] = currentSig
				if err := saveSignatures(config.Store, signatures); err != nil {
					fmt.Fprintf(os.Stderr, "%s Error saving signatures: %v\n", redBold("✗"), err)
					os.Exit(1)
				}
			}
		}
	}

	fmt.Println(strings.Repeat("─", 50))
	if failed {
		os.Exit(1)
	}
}

func runCheck(taskName string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if _, ok := config.Tasks[taskName]; !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown task '%s'\n", taskName)
		os.Exit(1)
	}

	includes := config.GetIncludesForTask(taskName)
	currentSig, err := calculateSignature(includes, config.Excludes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating signature: %v\n", err)
		os.Exit(1)
	}

	signatures, err := loadSignatures(config.Store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading signatures: %v\n", err)
		os.Exit(1)
	}

	if signatures[taskName] == currentSig {
		os.Exit(0)
	}
	os.Exit(1)
}

func runUpdate(taskName string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if _, ok := config.Tasks[taskName]; !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown task '%s'\n", taskName)
		os.Exit(1)
	}

	includes := config.GetIncludesForTask(taskName)
	currentSig, err := calculateSignature(includes, config.Excludes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating signature: %v\n", err)
		os.Exit(1)
	}

	signatures, err := loadSignatures(config.Store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading signatures: %v\n", err)
		os.Exit(1)
	}

	signatures[taskName] = currentSig
	if err := saveSignatures(config.Store, signatures); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving signatures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Code signature updated for task '%s': %s\n", taskName, currentSig)
}

func getHookPath() (string, error) {
	result, err := exec.Command("git", "rev-parse", "--git-path", "hooks").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository or git not installed")
	}
	return strings.TrimSpace(string(result)) + "/pre-commit", nil
}

func runInstallHook() {
	hookPath, err := getHookPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	hookDir := filepath.Dir(hookPath)
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "%s Error creating hooks directory: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	executable, err := os.Executable()
	if err != nil {
		executable = "./taskhash"
	}

	content := fmt.Sprintf(`#!/bin/bash
# Installed by taskhash
%s enforce
`, executable)

	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "%s Error writing hook: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	success("Hook installed at " + cyanBold(hookPath))
}

func runRemoveHook() {
	hookPath, err := getHookPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		skipped("No hook found")
		return
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error reading hook: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	if !strings.Contains(string(content), "# Installed by taskhash") {
		fmt.Fprintf(os.Stderr, "%s Hook was not installed by taskhash. Manual removal required.\n", yellowBold("!"))
		os.Exit(1)
	}

	if err := os.Remove(hookPath); err != nil {
		fmt.Fprintf(os.Stderr, "%s Error removing hook: %v\n", redBold("✗"), err)
		os.Exit(1)
	}

	success("Hook removed from " + cyanBold(hookPath))
}

func runUpdateBinary() {
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error: could not determine current executable\n", redBold("✗"))
		os.Exit(1)
	}

	running("Updating taskhash...")

	runCmd := exec.Command("go", "build", "-o", executable, ".")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	if err := runCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Build failed\n", redBold("✗"))
		os.Exit(1)
	}

	success("taskhash updated")
}
