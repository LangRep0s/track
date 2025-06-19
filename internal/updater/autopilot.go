package updater

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	version := strings.TrimPrefix(release.TagName, "v")
	assetName := AssetName(version)
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return errors.New("no suitable asset found for this OS/Arch")
	}
	zipPath, err := DownloadAsset(assetURL)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)
	targetDir, err := os.Executable()
	if err != nil {
		return err
	}
	targetDir = filepath.Dir(targetDir)
	if err := Unzip(zipPath, targetDir); err != nil {
		return err
	}
	return nil
}
