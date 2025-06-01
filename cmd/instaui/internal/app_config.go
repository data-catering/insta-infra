package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig handles application configuration and file management
type AppConfig struct {
	InstaDir string
	Version  string
	EmbedFS  embed.FS
	Logger   *AppLogger
}

// NewAppConfig creates a new application configuration
func NewAppConfig(embedFS embed.FS, version string, logger *AppLogger) *AppConfig {
	return &AppConfig{
		EmbedFS: embedFS,
		Version: version,
		Logger:  logger,
	}
}

// Initialize sets up the insta directory and determines the correct path
func (c *AppConfig) Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	instaDir := os.Getenv("INSTA_HOME")
	if instaDir == "" {
		c.InstaDir = filepath.Join(homeDir, ".insta")
	} else {
		c.InstaDir = instaDir
	}

	c.Logger.Log(fmt.Sprintf("UI insta directory: %s", c.InstaDir))
	return c.ensureComposeFiles()
}

// ensureComposeFiles extracts Docker Compose files from embedded resources if needed
func (c *AppConfig) ensureComposeFiles() error {
	if err := os.MkdirAll(c.InstaDir, 0755); err != nil {
		return fmt.Errorf("failed to create insta directory: %w", err)
	}

	if c.needsSync() {
		c.Logger.Log(fmt.Sprintf("Performing one-time file synchronization for version %s...", c.Version))
		return c.syncFiles()
	}

	return nil
}

// needsSync checks if files need to be synchronized
func (c *AppConfig) needsSync() bool {
	versionFilePath := filepath.Join(c.InstaDir, ".version_synced")
	syncedVersionBytes, err := os.ReadFile(versionFilePath)

	if err != nil {
		if !os.IsNotExist(err) {
			c.Logger.Log(fmt.Sprintf("Warning: failed to read version sync marker: %v", err))
		}
		return true
	}

	return strings.TrimSpace(string(syncedVersionBytes)) != c.Version
}

// syncFiles synchronizes embedded files to the insta directory
func (c *AppConfig) syncFiles() error {
	// Extract docker-compose files
	if err := c.extractComposeFile("docker-compose.yaml"); err != nil {
		return err
	}
	if err := c.extractComposeFile("docker-compose-persist.yaml"); err != nil {
		return err
	}

	// Extract data files
	if err := c.extractDataFiles(); err != nil {
		return fmt.Errorf("failed to extract data files: %w", err)
	}

	// Update version marker
	versionFilePath := filepath.Join(c.InstaDir, ".version_synced")
	if err := os.WriteFile(versionFilePath, []byte(c.Version), 0644); err != nil {
		c.Logger.Log(fmt.Sprintf("Warning: failed to write version marker: %v", err))
	}

	c.Logger.Log("File synchronization complete.")
	return nil
}

// extractComposeFile extracts a single compose file
func (c *AppConfig) extractComposeFile(filename string) error {
	content, err := c.EmbedFS.ReadFile("resources/" + filename)
	if err != nil {
		return fmt.Errorf("failed to read embedded %s: %w", filename, err)
	}

	targetPath := filepath.Join(c.InstaDir, filename)
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

// extractDataFiles extracts data files from embedded resources
func (c *AppConfig) extractDataFiles() error {
	dataDir := filepath.Join(c.InstaDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	return fs.WalkDir(c.EmbedFS, "resources/data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip persist directories
		if strings.Contains(path, "persist") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel("resources/data", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() {
			targetDir := filepath.Join(dataDir, relPath)
			return os.MkdirAll(targetDir, 0755)
		}

		content, err := c.EmbedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		targetFile := filepath.Join(dataDir, relPath)
		targetDir := filepath.Dir(targetFile)

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		return os.WriteFile(targetFile, content, 0755)
	})
}
