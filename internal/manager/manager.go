package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/user/track/internal/archiver"
	"github.com/user/track/internal/config"
	"github.com/user/track/internal/downloader"
	"github.com/user/track/internal/gh"
)

type Manager struct {
	Cfg *config.Config
}

func New() (*Manager, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}
	// Use GITHUB_TOKEN from env if set
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return NewWithToken(token)
	}
	return &Manager{Cfg: cfg}, nil
}

func NewWithToken(token string) (*Manager, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}
	mgr := &Manager{Cfg: cfg}
	if token != "" {
		os.Setenv("GITHUB_TOKEN", token)
	}
	return mgr, nil
}

func (m *Manager) UpdateRepo(repoPath string, force bool) error {
	fmt.Printf("Checking for updates for %s...\n", repoPath)

	repoCfg, ok := m.Cfg.Repos[repoPath]
	if !ok {
		return fmt.Errorf("repository '%s' not tracked", repoPath)
	}

	owner, name, _ := strings.Cut(repoPath, "/")
	client := gh.NewClient(context.Background(), "")

	latestRelease, err := client.GetLatestRelease(context.Background(), owner, name, repoCfg.IncludePrerelease)
	if err != nil {
		return fmt.Errorf("failed to get latest release for %s: %w", repoPath, err)
	}

	latestVersion := latestRelease.GetTagName()

	installName := repoCfg.InstallName
	if installName == "" {
		installName = name
	}
	latestDir := filepath.Join(m.Cfg.Global.DataDir, "latest")
	var binaryExists bool
	if runtime.GOOS == "windows" {
		shimPath := filepath.Join(latestDir, installName+".cmd")
		if fi, err := os.Stat(shimPath); err == nil && !fi.IsDir() {
			binaryExists = true
		}
	} else {
		binPath := filepath.Join(latestDir, installName)
		if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
			binaryExists = true
		}
	}

	if !force && latestVersion == repoCfg.CurrentVersion && binaryExists {
		fmt.Printf("'%s' is already up-to-date (version %s).\n", repoPath, latestVersion)
		return nil
	}

	if latestVersion != repoCfg.CurrentVersion {
		fmt.Printf("New version found for %s: %s (current: %s)\n", repoPath, latestVersion, repoCfg.CurrentVersion)
	} else {
		fmt.Printf("Reinstalling current version for %s: %s\n", repoPath, latestVersion)
	}

	if err := m.InstallVersion(repoPath, latestRelease); err != nil {
		return err
	}
	return nil
}

func (m *Manager) InstallVersion(repoPath string, release *github.RepositoryRelease) error {
	repoCfg := m.Cfg.Repos[repoPath]
	version := release.GetTagName()
	_, name, _ := strings.Cut(repoPath, "/")

	asset, err := gh.FindCompatibleAsset(release, repoCfg, &m.Cfg.Global)
	if err != nil {
		return fmt.Errorf("could not find compatible asset for %s in version %s: %w", repoPath, version, err)
	}
	fmt.Printf("Found compatible asset: %s\n", asset.GetName())

	repoDir := filepath.Join(m.Cfg.Global.DataDir, name)
	versionDir := filepath.Join(repoDir, "general", version)
	archivePath := filepath.Join(versionDir, asset.GetName())

	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("could not create version directory: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.GetBrowserDownloadURL())
	if err := downloader.DownloadFile(asset.GetBrowserDownloadURL(), archivePath); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	fmt.Printf("Extracting %s...\n", asset.GetName())
	if err := archiver.Extract(archivePath, versionDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	installName := repoCfg.InstallName
	if installName == "" {
		installName = name
	}

	executablePath, err := archiver.FindExecutable(versionDir, name, installName)
	if err != nil {
		return fmt.Errorf("could not find executable in archive for %s: %w", repoPath, err)
	}

	// --- Fix: Use correct versionDir for symlinks ---
	relativeExecPath, err := filepath.Rel(m.Cfg.Global.DataDir, executablePath)
	if err == nil {
		executablePath = filepath.Join(m.Cfg.Global.DataDir, relativeExecPath)
	}

	if runtime.GOOS == "windows" {
		globalLatestDir := filepath.Join(m.Cfg.Global.DataDir, "latest")
		os.MkdirAll(globalLatestDir, 0755)
		shimPath := filepath.Join(globalLatestDir, installName+".cmd")
		cmdContent := "@echo off\r\n\"" + executablePath + "\" %*\r\n"
		os.WriteFile(shimPath, []byte(cmdContent), 0755)
		fmt.Printf("Created Windows shim: %s\n", shimPath)
	} else {
		globalLatestDir := filepath.Join(m.Cfg.Global.DataDir, "latest")
		os.MkdirAll(globalLatestDir, 0755)
		symlinkPath := filepath.Join(globalLatestDir, installName)
		os.Remove(symlinkPath)
		err := os.Symlink(executablePath, symlinkPath)
		if err != nil {
			fmt.Printf("Failed to create symlink: %v\n", err)
		} else {
			fmt.Printf("Created symlink: %s -> %s\n", symlinkPath, executablePath)
		}

		// --- Add symlink to ~/.local/bin for Linux/macOS ---
		homeDir, err := os.UserHomeDir()
		if err == nil {
			userBin := filepath.Join(homeDir, ".local", "bin")
			if err := os.MkdirAll(userBin, 0755); err == nil {
				// Remove all symlinks in ~/.local/bin that point to any previous ripgrep/lazygit/track version
				entries, _ := os.ReadDir(userBin)
				for _, entry := range entries {
					if entry.Type()&os.ModeSymlink != 0 && (entry.Name() == installName) {
						os.Remove(filepath.Join(userBin, entry.Name()))
					}
				}
				userBinSymlink := filepath.Join(userBin, installName)
				_ = os.Remove(userBinSymlink)
				err := os.Symlink(executablePath, userBinSymlink)
				if err == nil {
					fmt.Printf("Created symlink: %s -> %s\n", userBinSymlink, executablePath)
				} else {
					fmt.Printf("Failed to create symlink in ~/.local/bin: %v\n", err)
				}
			}
		}
	}

	repoCfg.CurrentVersion = version
	if err := m.Cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config after update: %w", err)
	}

	fmt.Printf("Successfully installed %s version %s.\n", repoPath, version)
	return nil
}

func (m *Manager) AddRepo(repoPath string) error {
	if _, exists := m.Cfg.Repos[repoPath]; exists {
		return fmt.Errorf("repository '%s' is already being tracked", repoPath)
	}
	newRepo := &config.Repo{
		Path: repoPath,
	}
	m.Cfg.Repos[repoPath] = newRepo
	if err := m.Cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Successfully added '%s' to tracked repositories.\n", repoPath)
	return nil
}
