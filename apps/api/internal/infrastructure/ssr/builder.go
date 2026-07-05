// Package ssr provides modular SSR server management components.
package ssr

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// Builder handles building the web app for SSR.
type Builder struct {
	logger    *slog.Logger
	webDir    string
	publicDir string
}

// NewBuilder creates a new web app builder.
func NewBuilder(logger *slog.Logger, webDir, publicDir string) *Builder {
	return &Builder{
		logger:    logger,
		webDir:    webDir,
		publicDir: publicDir,
	}
}

// NeedsBuild checks if the web app needs to be built.
func (b *Builder) NeedsBuild() (bool, error) {
	clientPath := filepath.Join(b.webDir, "dist", "client")
	serverPath := filepath.Join(b.webDir, "dist", "server")

	clientExists, err := dirExists(clientPath)
	if err != nil {
		return false, err
	}

	serverExists, err := dirExists(serverPath)
	if err != nil {
		return false, err
	}

	return !clientExists || !serverExists, nil
}

// Build builds the web app using pnpm.
func (b *Builder) Build() error {
	b.logger.Info("Building web app", "dir", b.webDir)

	// Build the web app.
	cmd := exec.Command("pnpm", "run", "build")
	cmd.Dir = b.webDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm build failed: %w", err)
	}

	b.logger.Info("Web app built successfully")

	return nil
}

// CopyAssets copies built client assets to the public directory.
func (b *Builder) CopyAssets() error {
	clientDist := filepath.Join(b.webDir, "dist", "client")

	// Check if client dist exists.
	exists, err := dirExists(clientDist)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("client dist directory not found: %s", clientDist)
	}

	// Check if there are files to copy.
	files, err := os.ReadDir(clientDist)
	if err != nil {
		return fmt.Errorf("failed to read client dist: %w", err)
	}

	if len(files) == 0 {
		b.logger.Warn("No files in client dist to copy")
		return nil
	}

	// Ensure public directory exists.
	//
	if err := os.MkdirAll(b.publicDir, 0o755); err != nil {
		return fmt.Errorf("failed to create public dir: %w", err)
	}

	// Copy assets using shell command.
	// #nosec:G204 - Paths are validated before use (clientDist from known web build).
	cmd := exec.Command("cp", "-r", clientDist+"/", b.publicDir+"/")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	b.logger.Info("Assets copied to public directory", "from", clientDist, "to", b.publicDir)

	return nil
}

// BuildIfNeeded builds the web app if it doesn't exist.
func (b *Builder) BuildIfNeeded() error {
	needsBuild, err := b.NeedsBuild()
	if err != nil {
		return err
	}

	if !needsBuild {
		b.logger.Info("Web app already built, skipping build")
		return nil
	}

	b.logger.Info("Web app needs build")

	if err := b.Build(); err != nil {
		return err
	}

	return b.CopyAssets()
}

// Clean removes build artifacts.
func (b *Builder) Clean() error {
	distPath := filepath.Join(b.webDir, "dist")

	if err := os.RemoveAll(distPath); err != nil {
		return fmt.Errorf("failed to clean dist: %w", err)
	}

	b.logger.Info("Build artifacts cleaned")

	return nil
}

// dirExists checks if a directory exists.
func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return info.IsDir(), nil
}
