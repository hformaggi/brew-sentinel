package i18n

import (
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestSystemLocale(t *testing.T) {
	for _, tc := range []struct {
		env          map[string]string
		native, want string
	}{{map[string]string{"LANG": "pt_BR.UTF-8"}, "", "pt"}, {map[string]string{"LANG": "en_US.UTF-8"}, "pt-BR", "en"}, {map[string]string{"LANG": "fr_FR.UTF-8"}, "", "en"}, {map[string]string{"LANG": "pt_BR", "LC_ALL": "C"}, "", "en"}, {map[string]string{}, "pt-PT", "pt"}, {map[string]string{}, "", "en"}, {map[string]string{"LANG": "pt_BR", "LC_MESSAGES": "en_US"}, "", "en"}} {
		got := Detect(func(k string) string { return tc.env[k] }, func() string { return tc.native })
		if got != tc.want {
			t.Errorf("%v/%s = %s; want %s", tc.env, tc.native, got, tc.want)
		}
	}
}
func TestPreferenceOverride(t *testing.T) {
	if Resolve("en", "pt") != "en" || Resolve("pt", "en") != "pt" || Resolve("system", "pt") != "pt" {
		t.Fatal("preference ignored")
	}
}
func TestPreferencePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "settings.json")
	if e := Save(path, "pt"); e != nil {
		t.Fatal(e)
	}
	p, e := Load(path)
	if e != nil || p != "pt" {
		t.Fatalf("%s %v", p, e)
	}
	if e := Save(path, "xx"); e == nil {
		t.Fatal("invalid language saved")
	}
}
func TestEnglishFallback(t *testing.T) {
	if got := New("en").T("Untranslated text"); got != "Untranslated text" {
		t.Fatal(got)
	}
}

func TestTranslationsPreserveFormatArguments(t *testing.T) {
	pattern := regexp.MustCompile(`%[a-zA-Z]`)
	for english, portuguese := range portuguese {
		if !reflect.DeepEqual(pattern.FindAllString(english, -1), pattern.FindAllString(portuguese, -1)) {
			t.Errorf("format arguments differ: %q -> %q", english, portuguese)
		}
	}
}
