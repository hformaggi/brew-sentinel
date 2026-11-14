package tui

import "brew-sentinel/internal/i18n"

type Options struct{ Language, ConfigPath string }

func (u *UI) tr(text string) string { return u.language.T(text) }
func (u *UI) languageMenu() {
	u.menu(u.tr("Language"), []menuItem{
		{u.tr("System default"), u.tr("Follow the system language; English is the fallback"), 0, func() { u.setLanguage("system") }},
		{"English", u.tr("Always use English"), 0, func() { u.setLanguage("en") }},
		{"Português", u.tr("Always use Portuguese"), 0, func() { u.setLanguage("pt") }},
	})
}
func (u *UI) setLanguage(preference string) {
	if e := i18n.Save(u.configPath, preference); e != nil {
		u.message(u.tr("Could not save language"), u.errorText(e))
		return
	}

	u.useLanguage(preference)
}
func (u *UI) useLanguage(preference string) {
	u.preference = preference
	system := "en"
	if preference == "system" {
		system = i18n.System()
	}
	u.language = i18n.New(i18n.Resolve(preference, system))
	u.nav.SetTitle(" " + u.tr("Navigation") + " ")
	u.inspector.SetTitle(" " + u.tr("Package overview") + " ")
	u.search.SetLabel(u.tr("Search  ")).SetPlaceholder(u.tr("Name or description…"))
	u.status = u.tr("Language saved")
	u.render()
}

func (u *UI) errorText(err error) string { return i18n.ErrorText(err, u.tr) }
