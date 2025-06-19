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

		// DEBUG: Print all asset names and why they are skipped
		// fmt.Printf("[DEBUG] Checking asset: %s\n", name)
		PrintDebug(globalCfg, "Checking asset: %s", name)
		if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, ".sig") || strings.HasSuffix(name, ".pem") ||
			strings.HasSuffix(name, ".md5") || strings.HasSuffix(name, ".sha512") || strings.Contains(name, "checksum") ||
			strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".blockmap") || strings.Contains(name, "source") {
			PrintDebug(globalCfg, "Skipped (checksum/source): %s", name)
			continue
		}
		if excludeRe != nil && excludeRe.MatchString(name) {
			PrintDebug(globalCfg, "Skipped (excludeRe): %s", name)
			continue
		}
		if filterRe != nil && !filterRe.MatchString(name) {
			PrintDebug(globalCfg, "Skipped (filterRe): %s", name)
			continue
		}

		isOsMatch := false
		for _, osStr := range osChecks {
			if strings.Contains(name, osStr) {
				isOsMatch = true
				break
			}
		}
		isArchMatch := false
		for _, archStr := range archChecks {
			if strings.Contains(name, archStr) {
				isArchMatch = true
				break
			}
		}
		// Strict mode: require both OS and ARCH match
		if matcherMode == "strict" {
			if !isOsMatch || !isArchMatch {
				continue
			}
		}

		if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
			if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") || strings.Contains(name, "armv8") {
				continue
			}
		}
		// For Linux AMD64, strictly prefer AMD64 assets unless user requests ARM64
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			if !strings.Contains(strings.Join(assetPriority, ","), "arm64") && !strings.Contains(assetFilter, "arm64") && !strings.Contains(assetFilter, "aarch64") {
				if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") || strings.Contains(name, "armv8") {
					continue
				}
			}
		}
		// --- Improved Linux matcher for musl/gnu/any ---
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			// Only allow assets with amd64/x86_64/x64 and linux in the name (unless user requests arm64/aarch64)
			if (strings.Contains(name, "amd64") || strings.Contains(name, "x86_64") || strings.Contains(name, "x64")) && strings.Contains(name, "linux") {
				// fmt.Printf("[DEBUG] Candidate for Linux AMD64: %s\n", name)
				PrintDebug(globalCfg, "Candidate for Linux AMD64: %s", name)
				candidates = append(candidates, asset)
				// Do NOT continue here, allow further logic to run for other platforms
			} else {
				// fmt.Printf("[DEBUG] Skipped (not linux/amd64/x86_64/x64): %s\n", name)
				PrintDebug(globalCfg, "Skipped (not linux/amd64/x86_64/x64): %s", name)
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
	// Enhanced Linux AMD64 matcher: sort candidates by best match
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" && len(candidates) > 1 {
		preferredOrder := []string{
			"x86_64-unknown-linux-gnu",
			"amd64-linux-gnu",
			"x86_64-linux-gnu",
			"x86_64-unknown-linux",
			"amd64-unknown-linux-gnu",
			"amd64-unknown-linux",
			"x86_64-linux",
			"amd64-linux",
			"x64-linux",
			"linux-gnu",
			"musl",
			"linux",
		}
		bestIdx := -1
		bestScore := -1
		for idx, asset := range candidates {
			name := strings.ToLower(asset.GetName())
			for score, pat := range preferredOrder {
				if strings.Contains(name, pat) {
					if bestScore == -1 || score < bestScore {
						bestScore = score
						bestIdx = idx
					}
				}
			}
		}
		if bestIdx >= 0 {
			return candidates[bestIdx], nil
		}
	}

	return nil, fmt.Errorf("no assets found for your OS (%s) and arch (%s)", runtime.GOOS, runtime.GOARCH)
}

func getSystemKeywords() (os, arch []string) {
	switch runtime.GOOS {
	case "windows":
		os = []string{"windows", "win", "win64", "win32", ".exe", ".msi"}
	case "linux":
		os = []string{"linux", "ubuntu", "debian", "centos", "fedora", "appimage", "elf", "gnu"} // add gnu for linux-gnu
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

// PrintDebug prints debug messages if debug is enabled in the global config
func PrintDebug(globalCfg *config.GlobalConfig, format string, a ...interface{}) {
	if globalCfg != nil && globalCfg.Debug {
		fmt.Printf("[DEBUG] "+format+"\n", a...)
	}
}
