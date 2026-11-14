# User guide

[Back to README](../README.md) · [Português](usage.pt-BR.md)

## Language

The interface follows the system locale by default: Portuguese for `pt` locales,
English for other or unsupported locales. English is the fallback language.

Choose **Tools → Language** with the arrow keys to select:

- **System default**: follow the system locale.
- **English**: always use English.
- **Português**: always use Portuguese.

The choice applies immediately and is saved in `brew-sentinel/settings.json` under
the OS user configuration directory (`~/Library/Application Support/` on macOS).
A command-line override applies to one launch without changing the saved preference:

```sh
./bin/brew-sentinel --lang en
./bin/brew-sentinel --lang pt
./bin/brew-sentinel --lang system
```

Locale detection checks `LC_ALL`, `LC_MESSAGES`, `LANGUAGE` and `LANG`, in that
order, then macOS preferred languages when these are unset. Package descriptions,
upstream advisories and external command output remain as provided by their sources.

## Navigation

| Key | Action |
|---|---|
| ↑ / ↓ | Browse items or scroll text |
| ← / → | Switch panes, tabs or buttons |
| Enter | Open or activate the selected item |
| Esc | Go back or cancel |

In the package list, **←** moves to the sidebar. In the sidebar, **→** returns to
the list; on a library category, it also applies that filter. **Enter** opens a tool.
Press **↑** on the first package to enter search. In text fields, ←/→ edit the cursor.

Package details use tabs for Overview, Dependencies, Dependents, Security and Actions.
←/→ switch tabs; ↑/↓ browse their contents. Enter opens a related package or action.
In a dependency tree, → expands/enters a branch and ← collapses/returns to its parent.

In the catalog, ↑/↓ move between the query, package type and action buttons.
←/→ select Formulae/Applications or Search/Back. Enter in the query starts a search.
In operation previews, ↑ moves to the scrollable text and ←/→ return to the buttons,
keeping the previous selection. Enter activates only the focused button.

Optional shortcuts: `/` search, `i` catalog, `m` maintenance, `s` security report,
`v` rescan, `r` refresh, `o` sort, `?` help and `q` quit. Ctrl+P opens the command list.
Ctrl+C cancels read-only queries; package mutations wait for completion.

Search supports names/descriptions and combined filters such as
`type:formula reason:explicit`, `type:cask chrome` and `outdated:true`.
The Filter list menu provides ready-made filters without typing their syntax.

## Interpreting results

- **Available versions and updates** use local Homebrew metadata, which may be stale.
  The app does not run `brew update` automatically.
- **Installation reason** comes from receipts when available. Unknown means missing
  evidence; it does not mean a package is unused.
- **Direct dependencies** prefer receipt entries marked `declared_directly`. Current
  recipe data is a labeled fallback; incomplete receipts retain proven relationships.
  Observed runtime dependencies may include transitive relationships.
- **Disk usage** uses `du -sk` without following symlinks. It measures formula kegs,
  including retained versions, without counting a dependency again for each dependent.
  Values are filesystem-reported allocated bytes, not exclusive physical APFS usage.
  Cask sizes are unavailable and cache is excluded from this total.
- **Security** uses `brew vulns --json` and its external service. Casks are not covered.
  Reports retain coverage gaps, skipped packages, errors and formula patches. No findings
  means no known matches in the assessed scope, not proof of vulnerability absence.
  Retained kegs may not all have been scanned. Findings identify upstream versions,
  which can differ from Homebrew revisions. Rescan after an upgrade to verify findings.

Opening the overview or navigating to a package never mutates the installation.
Install/upgrade show a dry run; uninstall shows known dependents and lets Homebrew
perform final validation. Cleanup and orphan removal are separate previewed operations.
No force flags, ignored dependencies or manual file deletion are used.

Formula operations stream up to 1 MiB of output in memory. Cask installers use the
suspended terminal to preserve interaction and password prompts, without capturing
that transcript. Returning to the app reloads inventory even after a partial failure.
External changes can occur between preview and execution.

Security and disk queries run concurrently; mutations and inventory refresh wait for
them to finish. App-owned interface text supports both languages; source data is not
machine-translated. Inventory and operation results remain in memory.
