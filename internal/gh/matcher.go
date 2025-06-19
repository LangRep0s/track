package gh

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/user/track/internal/config"
)

func FindCompatibleAsset(release *github.RepositoryRelease, repoCfg *config.Repo, globalCfg *config.GlobalConfig) (*github.ReleaseAsset, error) {
	
	assetPriority := repoCfg.AssetPriority
	if len(assetPriority) == 0 && globalCfg != nil {
		assetPriority = globalCfg.DefaultAssetPriority
	}
	preferredArchives := repoCfg.PreferredArchives
	if len(preferredArchives) == 0 && globalCfg != nil {
		preferredArchives = globalCfg.PreferredArchiveTypes
	}
	matcherMode := repoCfg.MatcherMode
	if matcherMode == "" && globalCfg != nil {
		matcherMode = globalCfg.MatcherMode
	}
	if matcherMode == "" {
		matcherMode = "strict"
	}
	fallbackArch := repoCfg.FallbackArch
	fallbackOS := repoCfg.FallbackOS

	
	assetFilter := repoCfg.AssetFilter
	if assetFilter == "" && globalCfg != nil {
		assetFilter = globalCfg.DefaultAssetFilter
	}
	assetExclude := repoCfg.AssetExclude

	var filterRe, excludeRe *regexp.Regexp
	if assetFilter != "" {
		filterRe, _ = regexp.Compile(assetFilter)
	}
	if assetExclude != "" {
		excludeRe, _ = regexp.Compile(assetExclude)
	}

	osChecks, archChecks := getSystemKeywords()
	strictOS := runtime.GOOS

	
	matchesPriority := func(name string, priorities []string) bool {
		for _, p := range priorities {
			if strings.Contains(name, strings.ToLower(p)) {
				return true
			}
		}
		return false
	}
	matchesArchive := func(name string, exts []string) bool {
		for _, ext := range exts {
			if strings.HasSuffix(name, strings.ToLower(ext)) {
				return true
			}
		}
		return false
	}

	var candidates []*github.ReleaseAsset
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.GetName())
		
		if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, ".sig") || strings.HasSuffix(name, ".pem") ||
			strings.HasSuffix(name, ".md5") || strings.HasSuffix(name, ".sha512") || strings.Contains(name, "checksum") ||
			strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".blockmap") || strings.Contains(name, "source") {
			continue
		}
		if excludeRe != nil && excludeRe.MatchString(name) {
			continue
		}
		if filterRe != nil && !filterRe.MatchString(name) {
			continue
		}
		
		isOsMatch := false
		for _, osStr := range osChecks {
			if strings.Contains(name, osStr) {
				isOsMatch = true
				break
			}
		}
		if !isOsMatch && matcherMode == "strict" {
			continue
		}
		isArchMatch := false
		for _, archStr := range archChecks {
			if strings.Contains(name, archStr) {
				isArchMatch = true
				break
			}
		}
		if !isArchMatch && matcherMode == "strict" {
			continue
		}
		
		if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
			if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") || strings.Contains(name, "armv8") {
				continue
			}
		}
		
		if strictOS == "windows" && !strings.Contains(name, "windows") {
			continue
		}
		if strictOS == "darwin" && !strings.Contains(name, "darwin") {
			continue
		}
		if strictOS == "linux" && !strings.Contains(name, "linux") {
			continue
		}
		candidates = append(candidates, asset)
	}

	
	if len(assetPriority) > 0 {
		for _, asset := range candidates {
			name := strings.ToLower(asset.GetName())
			if matchesPriority(name, assetPriority) {
				if len(preferredArchives) > 0 && matchesArchive(name, preferredArchives) {
					return asset, nil
				}
				if len(preferredArchives) == 0 {
					return asset, nil
				}
			}
		}
	}
	
	if len(preferredArchives) > 0 {
		for _, asset := range candidates {
			name := strings.ToLower(asset.GetName())
			if matchesArchive(name, preferredArchives) {
				return asset, nil
			}
		}
	}
	
	if len(candidates) > 0 {
		return candidates[0], nil
	}
	
	if matcherMode == "relaxed" && (len(fallbackArch) > 0 || len(fallbackOS) > 0) {
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.GetName())
			archOk := len(fallbackArch) == 0
			osOk := len(fallbackOS) == 0
			for _, arch := range fallbackArch {
				if strings.Contains(name, strings.ToLower(arch)) {
					archOk = true
					break
				}
			}
			for _, osStr := range fallbackOS {
				if strings.Contains(name, strings.ToLower(osStr)) {
					osOk = true
					break
				}
			}
			if archOk && osOk {
				return asset, nil
			}
		}
	}
	return nil, fmt.Errorf("no assets found for your OS (%s) and arch (%s)", runtime.GOOS, runtime.GOARCH)
}

func getSystemKeywords() (os, arch []string) {
	switch runtime.GOOS {
	case "windows":
		os = []string{"windows", "win", "win64", "win32", ".exe", ".msi"}
	case "linux":
		os = []string{"linux", "ubuntu", "debian", "centos", "fedora", "appimage", "elf"}
	case "darwin":
		os = []string{"darwin", "macos", "osx", "apple"}
	}

	switch runtime.GOARCH {
	case "amd64":
		arch = []string{"x86_64", "amd64", "x64", "64bit", "64-bit", "64"}
	case "arm64":
		arch = []string{"arm64", "aarch64", "armv8", "m1", "m2"}
	case "386":
		arch = []string{"386", "i386", "i686", "x86", "32bit", "32-bit", "32"}
	}
	return
}
