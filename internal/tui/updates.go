package tui

import (
	"context"
	"strings"

	"brew-sentinel/internal/brew"
)

func (u *UI) updatesMenu() {
	u.menu(u.tr("Updates"), []menuItem{
		{label: u.tr("Upgrade all packages"), description: u.tr("Preview updates for formulae and applications"), run: func() { u.upgradePackages("all") }},
		{label: u.tr("Upgrade formulae"), description: u.tr("Preview updates for command-line tools and libraries"), run: func() { u.upgradePackages("formula") }},
		{label: u.tr("Upgrade applications"), description: u.tr("Preview updates for installed casks"), run: func() { u.upgradePackages("cask") }},
		{label: u.tr("Update Homebrew"), description: u.tr("Refresh Homebrew and the package catalog"), run: u.updateHomebrew},
	})
}

func (u *UI) upgradePackages(scope string) {
	args, err := brew.UpgradeArgs(scope)
	if err != nil {
		u.message(u.tr("Invalid operation"), u.errorText(err))
		return
	}
	title := map[string]string{
		"all":     "Upgrade all packages",
		"formula": "Upgrade formulae",
		"cask":    "Upgrade applications",
	}[scope]
	// All-package upgrades can include casks that require terminal interaction.
	kind := "cask"
	if scope == "formula" {
		kind = "formula"
	}
	u.startTask("Preparing upgrade preview…", func(ctx context.Context) (func(), error) {
		previewArgs := append(append([]string(nil), args...), "--dry-run")
		out, err := u.client.Read(ctx, previewArgs...)
		if err != nil {
			return nil, err
		}
		text := strings.TrimSpace(string(out))
		if text == "" {
			return func() { u.message(u.tr(title), u.tr("Homebrew did not list any packages to upgrade.")) }, nil
		}
		return func() {
			u.confirm(u.tr(title), text+u.tr("\n\nReview the preview before confirming. External changes may affect the result."), func() { u.perform(args, kind) })
		}, nil
	})
}

func (u *UI) updateHomebrew() {
	if !u.ready() {
		return
	}
	u.confirm(u.tr("Update Homebrew"), u.tr("Refresh Homebrew and the package catalog. Installed packages will not be upgraded."), func() {
		u.perform([]string{"update"}, "formula")
	})
}
