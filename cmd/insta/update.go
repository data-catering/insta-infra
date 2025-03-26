package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// checkForUpdates checks if a new version is available and prints information if one is found.
func (a *App) checkForUpdates() error {
	// Get current version
	currentVersion := version

	// Get latest version from GitHub API
	resp, err := http.Get("https://api.github.com/repos/data-catering/insta-infra/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	// Remove 'v' prefix for comparison
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// Compare versions
	if latestVersion != currentVersion {
		fmt.Printf("A new version is available: %s\n", release.TagName)
		fmt.Printf("Download at: %s\n", release.HTMLURL)
		fmt.Printf("Current version: %s\n", version)
		return nil
	}

	return nil
}

// update downloads and installs the latest version of the application.
func (a *App) update() error {
	// Get current version
	currentVersion := version

	// Get latest version from GitHub API
	resp, err := http.Get("https://api.github.com/repos/data-catering/insta-infra/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name        string `json:"name"`
			BrowserURL  string `json:"browser_download_url"`
			ContentType string `json:"content_type"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	// Remove 'v' prefix for comparison
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// Check if update is needed
	if latestVersion == currentVersion {
		fmt.Println("You are already running the latest version.")
		return nil
	}

	// Determine the correct asset based on OS and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH
	var assetName string
	var assetURL string

	switch osName {
	case "windows":
		assetName = fmt.Sprintf("insta-%s-windows-amd64.zip", release.TagName)
	case "darwin":
		assetName = fmt.Sprintf("insta-%s-darwin-%s.tar.gz", release.TagName, arch)
	case "linux":
		assetName = fmt.Sprintf("insta-%s-linux-%s.tar.gz", release.TagName, arch)
	default:
		return fmt.Errorf("unsupported operating system: %s", osName)
	}

	// Find the matching asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no matching release found for %s", assetName)
	}

	// Download the update
	fmt.Printf("Downloading %s...\n", assetName)
	resp, err = http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	// Create temporary directory for the update
	tmpDir, err := os.MkdirTemp("", "insta-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save the downloaded file
	updateFile := filepath.Join(tmpDir, assetName)
	file, err := os.Create(updateFile)
	if err != nil {
		return fmt.Errorf("failed to create update file: %v", err)
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		return fmt.Errorf("failed to save update: %v", err)
	}
	file.Close()

	// Extract the update
	fmt.Println("Extracting update...")
	if strings.HasSuffix(assetName, ".zip") {
		cmd := exec.Command("unzip", "-o", updateFile, "-d", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract zip: %v", err)
		}
	} else {
		cmd := exec.Command("tar", "-xzf", updateFile, "-C", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract tar: %v", err)
		}
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Create backup of current executable
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Copy new executable
	newExec := filepath.Join(tmpDir, "insta")
	if osName == "windows" {
		newExec += ".exe"
	}
	if err := os.Rename(newExec, execPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install update: %v", err)
	}

	// Remove backup
	os.Remove(backupPath)

	fmt.Printf("Successfully updated to version %s\n", release.TagName)
	return nil
}
