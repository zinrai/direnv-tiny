package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var logger *slog.Logger

func init() {
	// デバッグモードの設定
	var level slog.Level
	if os.Getenv("DIRENV_TINY_DEBUG") == "1" {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	// ロガーの設定
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	if len(os.Args) < 2 {
		logger.Error("Invalid usage", "error", "insufficient arguments")
		fmt.Println("Usage: direnv-tiny <hook|export>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hook":
		fmt.Println(bashHook())
	case "export":
		if err := exportEnv(); err != nil {
			logger.Error("Failed to export environment", "error", err)
			os.Exit(1)
		}
	default:
		logger.Error("Unknown command", "command", os.Args[1])
		os.Exit(1)
	}
}

func bashHook() string {
	return `__direnv_hook() {
  local previous_exit_status=$?
  eval "$(direnv-tiny export)"
  return $previous_exit_status
}

if ! [[ "${PROMPT_COMMAND:-}" =~ __direnv_hook ]]; then
  PROMPT_COMMAND="__direnv_hook;${PROMPT_COMMAND:-}"
fi`
}

func exportEnv() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	previousDir := os.Getenv("DIRENV_TINY_PREVIOUS_DIR")

	logger.Debug("Exporting environment",
		"currentDir", currentDir,
		"previousDir", previousDir)

	if currentDir == previousDir {
		logger.Debug("Still in the same directory. No action needed.")
		return nil
	}

	if previousDir != "" {
		previousEnvrcPath := filepath.Join(previousDir, ".envrc")
		if _, err := os.Stat(previousEnvrcPath); err == nil {
			logger.Debug("Unloading previous .envrc", "path", previousEnvrcPath)
			if err := unloadEnv(previousEnvrcPath); err != nil {
				return fmt.Errorf("failed to unload previous .envrc: %w", err)
			}
		}
	}

	envrcPath := filepath.Join(currentDir, ".envrc")
	if _, err := os.Stat(envrcPath); err == nil {
		logger.Debug("Loading .envrc", "path", envrcPath)
		if err := loadEnv(envrcPath); err != nil {
			return fmt.Errorf("failed to load .envrc: %w", err)
		}
	}

	fmt.Printf("export DIRENV_TINY_PREVIOUS_DIR=\"%s\"\n", currentDir)
	return nil
}

func loadEnv(envrcPath string) error {
	file, err := os.Open(envrcPath)
	if err != nil {
		return fmt.Errorf("failed to open .envrc: %w", err)
	}
	defer file.Close()

	return processEnvFile(file, func(key, value string) {
		logger.Debug("Exporting variable", "key", key, "value", value)
		fmt.Printf("export %s=\"%s\"\n", key, value)
	})
}

func unloadEnv(envrcPath string) error {
	file, err := os.Open(envrcPath)
	if err != nil {
		return fmt.Errorf("failed to open .envrc: %w", err)
	}
	defer file.Close()

	return processEnvFile(file, func(key, _ string) {
		logger.Debug("Unsetting variable", "key", key)
		fmt.Printf("unset %s\n", key)
	})
}

func processEnvFile(r io.Reader, process func(key, value string)) error {
	scanner := bufio.NewScanner(r)
	lineRegex := regexp.MustCompile(`^([[:alnum:]_]+)=(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if matches == nil {
			logger.Warn("Invalid line in .envrc", "line", line)
			continue
		}

		key := matches[1]
		value := strings.Trim(matches[2], "\"'")
		process(key, value)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .envrc: %w", err)
	}

	return nil
}
