package updater

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubRepo  = "LangRep0s/track"
	apiReleases = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

func GetOSArch() (string, string) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return osName, arch
}

func AssetName(version string) string {
	osName, arch := GetOSArch()
	return fmt.Sprintf("track-%s-%s-%s.zip", version, osName, arch)
}

func FetchLatestRelease() (*ReleaseInfo, error) {
	resp, err := http.Get(apiReleases)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// DownloadAsset downloads the asset to a temp file
func DownloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: %d", resp.StatusCode)
	}
	tmpFile, err := os.CreateTemp("", "track-update-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		inFile, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, inFile)
		outFile.Close()
		inFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateTrack() error {
	release, err := FetchLatestRelease()
	if err != nil {
		return err
	}
	currentVersion := "v1.0.1"
	fmt.Printf("Current track version: %s\n", currentVersion)
	fmt.Printf("Latest available version: %s\n", release.TagName)
	if strings.TrimPrefix(release.TagName, "v") == strings.TrimPrefix(currentVersion, "v") {
		fmt.Println("track CLI is already up-to-date.")
		return nil
	}
	fmt.Println("New update found! Proceeding to update track CLI...")
	var asset *ReleaseAsset
	osName := runtime.GOOS
	arch := runtime.GOARCH
	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if osName == "windows" && !strings.Contains(name, "windows") {
			continue
		}
		if osName == "linux" && !strings.Contains(name, "linux") {
			continue
		}
		if osName == "darwin" && !strings.Contains(name, "darwin") {
			continue
		}
		if arch == "amd64" && !(strings.Contains(name, "amd64") || strings.Contains(name, "x86_64") || strings.Contains(name, "x64")) {
			continue
		}
		if arch == "arm64" && !(strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") || strings.Contains(name, "armv8")) {
			continue
		}
		if arch == "386" && !(strings.Contains(name, "386") || strings.Contains(name, "i386") || strings.Contains(name, "i686") || strings.Contains(name, "x86")) {
			continue
		}
		if strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
			asset = &a
			break
		}
	}
	if asset == nil {
		return errors.New("no suitable asset found for this OS/Arch")
	}
	zipPath, err := DownloadAsset(asset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)
	targetExe, err := os.Executable()
	if err != nil {
		return err
	}
	targetDir := filepath.Dir(targetExe)
	if runtime.GOOS == "windows" {
		// Defer self-replace: extract to temp, spawn a script to replace after exit
		tmpDir, err := os.MkdirTemp("", "track-update-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		if err := Unzip(zipPath, tmpDir); err != nil {
			return err
		}
		newExe := filepath.Join(tmpDir, filepath.Base(targetExe))
		bat := filepath.Join(tmpDir, "replace_track.bat")
		batContent := "@echo off\r\n" +
			"echo Waiting for track.exe to exit...\r\n" +
			":loop\r\n" +
			"tasklist | findstr /I \"track.exe\" >nul\r\n" +
			"if not errorlevel 1 (timeout /t 1 >nul & goto loop)\r\n" +
			"copy /Y \"" + newExe + "\" \"" + targetExe + "\"\r\n" +
			"echo Updated!\r\n" +
			"start \"\" \"" + targetExe + "\"\r\n"
		if err := os.WriteFile(bat, []byte(batContent), 0755); err != nil {
			return err
		}
		fmt.Println("Update downloaded. The CLI will now exit and update itself you may see a terminal open up...")
		// Launch the batch file and exit
		cmd := exec.Command("cmd", "/C", bat)
		cmd.Start()
		os.Exit(0)
		return nil
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		tmpDir, err := os.MkdirTemp("", "track-update-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		if err := Unzip(zipPath, tmpDir); err != nil {
			return err
		}
		newExe := filepath.Join(tmpDir, filepath.Base(targetExe))
		sh := filepath.Join(tmpDir, "replace_track.sh")
		shContent := "#!/bin/sh\n" +
			"echo Waiting for track to exit..." + "\n" +
			"while lsof | grep \"$1\" > /dev/null; do sleep 1; done\n" +
			"cp \"$1\" \"$2\"\n" +
			"chmod +x \"$2\"\n" +
			"echo Updated!\n" +
			"exec \"$2\"\n"
		if err := os.WriteFile(sh, []byte(shContent), 0755); err != nil {
			return err
		}
		fmt.Println("Update downloaded. The CLI will now exit and update itself...")
		cmd := exec.Command("sh", sh, newExe, targetExe)
		cmd.Start()
		os.Exit(0)
		return nil
	}
	if err := Unzip(zipPath, targetDir); err != nil {
		return err
	}
	return nil
}
