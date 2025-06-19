# Track CLI

A powerful, cross-platform CLI tool to track, download, and manage GitHub repository releases and their binaries. Track CLI is designed for developers, power users, and sysadmins who want robust, automated binary management with advanced configuration.

---

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [Add a Repository](#add-a-repository)
  - [List Tracked Repositories](#list-tracked-repositories)
  - [Update Repositories](#update-repositories)
  - [Remove a Repository](#remove-a-repository)
  - [Rollback to Previous Version](#rollback-to-previous-version)
  - [Tidy Old Versions](#tidy-old-versions)
  - [Configuration](#configuration)
  - [Advanced Asset Matching](#advanced-asset-matching)
- [Example Config](#example-config)
- [Tips & Tricks](#tips--tricks)
- [FAQ](#faq)
- [Contributing](#contributing)

---

## Features
- 🚀 Track and update GitHub releases for multiple repositories
- 🧠 Smart asset selection for your OS/arch (Windows, Linux, macOS)
- 📦 Download, extract, and manage binaries in versioned folders
- 🔗 Global accessibility via shims (Windows) or symlinks (Linux/macOS) in the `track/latest` folder on all platforms
- ⚙️ Flexible config: per-repo and global options (filters, prerelease, asset priorities, etc.)
- 🧹 One-command cleanup of old versions (`track tidy`)
- 🔒 No environment variable hacks
- 📝 Easy config editing and CLI config toggling

---

## Installation

1. **Build from source:**
   ```sh
   go build -o track ./cmd/track
   ```
2. **(Optional)** Place the binary in your PATH or use as-is.

---

## Quick Start

![Track CLI Demo](https://user-images.githubusercontent.com/6759207/210175302-2e2e7b2e-6e7e-4b7e-8e7e-2e2e7b2e6e7e.gif)

```sh
track add BurntSushi/ripgrep
track list
track update
track set 1 prerelease true
track tidy
```

---

## Commands

### Add a Repository
```sh
track add BurntSushi/ripgrep
track add jesseduffield/lazygit
```

### List Tracked Repositories
```sh
track list
```

### Update Repositories
```sh
track update           # Update all
track update 1         # Update repo #1 from the list
```

### Remove a Repository
```sh
track remove BurntSushi/ripgrep
track remove 2         # Remove by list number
```

### Rollback to Previous Version
```sh
track rollback 1 v14.0.0
```

### Tidy Old Versions
Delete all previous versions (keep only the current):
```sh
track tidy
```

### Configuration

#### Open the config file in your editor
```sh
track config
track --config
```
If the editor fails to open, the CLI will print the config file path for manual editing.

#### Where is the config file stored?
- **Windows:** `%LOCALAPPDATA%\track\config.json` (all program data and config are in `%LOCALAPPDATA%\track`)
- **Linux/macOS:** `~/.local/share/track/config.json` (all program data and config are in `~/.local/share/track`)

#### Set per-repo options from the CLI
```sh
track set 1 prerelease true
track set 2 MatcherMode strict
track set 1 AssetFilter ".*musl.*"
track set 1 AssetPriority x86_64,amd64
track set 2 PreferredArchives .zip,.tar.gz
```
Supported fields: `prerelease`, `MatcherMode`, `AssetFilter`, `AssetExclude`, `InstallName`, `AssetPriority`, `PreferredArchives`, `FallbackArch`, `FallbackOS`.

---

## Advanced Asset Matching

- Asset selection can be tuned globally or per-repo using config or `track set`.
- Supports asset priorities, preferred archive types, fallback arch/os, strict/relaxed mode, and regex filters.
- Example: Prefer musl builds, or .zip over .tar.gz, or fallback to arm64 if x86_64 is missing.

---

## Example Config

```json
{
  "global": {
    "data_dir": "/Users/you/.local/share/track",
    "default_asset_priority": ["x86_64", "amd64"],
    "preferred_archive_types": [".zip", ".tar.gz"],
    "matcher_mode": "strict"
  },
  "repos": {
    "BurntSushi/ripgrep": {
      "include_prerelease": false,
      "asset_filter": ".*musl.*",
      "matcher_mode": "strict"
    },
    "jesseduffield/lazygit": {
      "include_prerelease": true,
      "asset_priority": ["x86_64", "amd64"],
      "preferred_archives": [".zip"]
    }
  }
}
```

---

## Tips & Tricks
- Use `track set` to quickly toggle or set config fields without editing JSON.
- Use `track tidy` regularly to save disk space.
- Use asset filters to avoid unwanted builds (e.g. ARM on AMD64).
- Use `track list` to see repo numbers for use in other commands.
- All shims (Windows) and symlinks (Linux/macOS) are created in the `track/latest` folder for easy access.

---

## FAQ

**Q: Where is the config file stored?**
A: See [Configuration](#configuration) above. All program data and config are in a single folder for your OS.

**Q: How do I add a repo with a custom binary name?**
A: Use `track set <repo#|repo> InstallName <name>` after adding the repo.

**Q: How do I prefer a specific asset type?**
A: Use `track set <repo#|repo> AssetPriority x86_64,amd64` or `track set <repo#|repo> PreferredArchives .zip,.tar.gz`.

**Q: How do I remove all old versions?**
A: Use `track tidy`.

---

## Contributing

Pull requests, issues, and suggestions are welcome! Please open an issue or PR on GitHub.

---

**Track CLI** is open source and extensible. Star the repo and share your feedback!
