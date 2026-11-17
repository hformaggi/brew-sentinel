# Brew Sentinel

Explore and manage your Homebrew packages from an interactive terminal interface.

See what's installed, understand package dependencies, and review updates and
known vulnerabilities in one place. Built in Go for macOS.

## Features

- Browse installed formulae and casks, versions, and available updates.
- Explore dependencies, dependents, and installation reasons.
- Review vulnerability findings and scan coverage for formulae.
- Search the Homebrew catalog and install, upgrade, or uninstall packages.
- Preview cleanup and orphan removal before confirming.
- Filter your inventory and sort packages by name, disk usage, or dependents.

## Installation

Requires macOS with [Homebrew](https://brew.sh/) installed.

```sh
brew install hformaggi/tap/brew-sentinel
brew-sentinel
```

### Build from source

Requires Go 1.27 or later.

```sh
git clone https://github.com/hformaggi/brew-sentinel.git
cd brew-sentinel
go build -o bin/brew-sentinel ./cmd/brew-sentinel
./bin/brew-sentinel
```

## Usage

Use the sidebar to browse packages, search the catalog, and open maintenance tools
or security reports. Open a package to see its details, dependencies, and actions.
Package changes require confirmation; installation and upgrade include a preview.

Open **Maintenance → Updates** to upgrade all packages, only formulae, or only
applications. The same menu lets you update Homebrew and refresh its package catalog.

| Key | Action |
| --- | --- |
| ↑ / ↓ | Browse items or scroll |
| ← / → | Switch panes, tabs, or buttons |
| Enter | Open or activate the selection |
| Esc | Go back or close a dialog |
| `/` | Search installed packages |
| Ctrl+P | Open the command menu |
| `?` | Show navigation help |
| `q` | Quit from the package list |

### Search and filters

Search by package name or description, or combine filters:

| Query | Result |
| --- | --- |
| `type:formula reason:explicit` | Explicitly installed formulae |
| `type:cask chrome` | Casks matching "chrome" |
| `outdated:true` | Packages with an available update |
| `reason:dependency` | Packages installed as dependencies |

Update availability uses local Homebrew metadata. Refresh it through
**Maintenance → Updates → Update Homebrew** or by running `brew update`.

### Command-line options

The interface supports English and Portuguese. Use `--lang` to choose a language
for the current session, or select one from the command menu.

```sh
brew-sentinel --lang en     # English
brew-sentinel --lang pt     # Portuguese
brew-sentinel --json        # Print the installed package inventory as JSON
brew-sentinel --version     # Show the version
brew-sentinel --help        # List available options
```

### Vulnerability reports

Reports require a Homebrew version with `brew vulns` support and an internet
connection. They cover known vulnerabilities in formulae; coverage is not
guaranteed for casks or retained older versions. Review the coverage summary
when interpreting results, including reports with no findings.

## Contributing

Suggestions, bug reports, and pull requests are welcome. Open an
[issue](https://github.com/hformaggi/brew-sentinel/issues) with steps to reproduce
and your macOS, Homebrew, and Brew Sentinel versions, or submit a
[pull request](https://github.com/hformaggi/brew-sentinel/pulls).

Run the checks before submitting code:

```sh
go test -race ./...
go vet ./...
```

## License

[BSD 3-Clause](LICENSE).
