# Brew Sentinel

**Your Homebrew packages, dependencies, and security findings in one terminal interface.**

Brew Sentinel helps you explore your installed formulae and casks, understand why
packages are installed, and review changes before applying them through Homebrew.

[Português](README.pt-BR.md) · [Getting started](#getting-started) · [Report an issue](https://github.com/hformaggi/brew-sentinel/issues)

## Features

- **Package inventory** — inspect installed and available versions, outdated packages,
  pinned packages, and installation reasons.
- **Dependency explorer** — navigate dependencies and dependents, with installation
  receipts used as evidence where available.
- **Vulnerability reports** — scan formulae through `brew vulns --json`, with findings,
  formula patches, and coverage gaps shown together.
- **Package management** — search the Homebrew catalog, install, upgrade, uninstall,
  clean up, and remove orphaned dependencies through preview and confirmation flows.
- **Search and filters** — combine package names, descriptions, types, and installation
  reasons to narrow down your inventory.

## Getting started

### Requirements

- macOS with [Homebrew](https://brew.sh/) installed and `brew` available on `PATH`.
- Go **1.27 or later** to build from source.
- Git to clone the repository.
- For vulnerability reports, a Homebrew version that provides `brew vulns --json`
  and access to the scanner's external service.

### Homebrew

```sh
brew install hformaggi/tap/brew-sentinel
```

### Build from source

```sh
git clone https://github.com/hformaggi/brew-sentinel.git
cd brew-sentinel
go build -o bin/brew-sentinel ./cmd/brew-sentinel
./bin/brew-sentinel
```

## Usage

Use the arrow keys to navigate and **Enter** to open a package or tool.
The sidebar provides access to search, filters, sorting, maintenance, and reports.

| Key | Action |
| --- | --- |
| ↑ / ↓ | Browse items or scroll text |
| ← / → | Switch panes, tabs, or buttons |
| Enter | Open or activate the selection |
| Esc | Go back or cancel |
| `/` | Search packages |
| Ctrl+P | Open the command list |
| `?` | Show help |
| `q` | Quit |

### Search examples

Enter these expressions in the package search field. Filters can be combined.

| Query | Result |
| --- | --- |
| `type:formula reason:explicit` | Explicitly installed formulae |
| `type:cask chrome` | Casks matching “chrome” |
| `outdated:true` | Packages with an available update in local metadata |
| `reason:dependency` | Packages installed as dependencies |

### Command-line options

```sh
./bin/brew-sentinel --lang en      # Use English for this launch
./bin/brew-sentinel --lang pt      # Use Portuguese for this launch
./bin/brew-sentinel --lang system  # Follow the system locale
./bin/brew-sentinel --json         # Print package inventory without opening the UI
./bin/brew-sentinel --version      # Print the build version
./bin/brew-sentinel --help         # List available options
```

## How changes and reports work

Browsing packages does not change your installation. Install and upgrade actions
show a Homebrew dry run before confirmation; uninstall shows known dependents
and leaves final validation to Homebrew. Cleanup and orphan removal have separate
previews. The app does not use force flags or bypass dependency checks.

Available versions come from local Homebrew metadata. Brew Sentinel does not run
`brew update` automatically, so refresh that metadata yourself when needed.

Vulnerability reports cover formulae, with no confirmed coverage for casks or
all retained kegs. An empty report means no known matches in the assessed scope;
check the reported coverage and rescan after upgrading. Scanning uses Homebrew's
external service. Inventory and operation results stay in memory unless you save
JSON output yourself.

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
```

Tests use synthetic inventories, fake Homebrew commands, temporary directories,
and simulated terminal screens. They do not install or remove real packages.

For filter benchmarks:

```sh
go test ./internal/tui -bench . -run '^$'
```

Contributions are welcome through [issues](https://github.com/hformaggi/brew-sentinel/issues)
and [pull requests](https://github.com/hformaggi/brew-sentinel/pulls).
For bug reports, include your macOS, Homebrew, and Brew Sentinel versions,
reproduction steps, and relevant output with personal information removed.

## License

[BSD 3-Clause](LICENSE) © 2026 hformaggi.
