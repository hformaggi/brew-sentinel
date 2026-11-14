package i18n

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed pt.json
var portugueseJSON []byte
var portuguese = func() map[string]string {
	var m map[string]string
	if e := json.Unmarshal(portugueseJSON, &m); e != nil {
		panic(e)
	}
	return m
}()

type Translator struct{ Language string }

func New(language string) Translator {
	if language != "pt" {
		language = "en"
	}
	return Translator{Language: language}
}

// English source text is also the fallback for missing translations.
func (t Translator) T(source string) string {
	if t.Language == "pt" {
		if s, ok := portuguese[source]; ok {
			return s
		}
	}
	return source
}
func Preference(value string) (string, error) {
	switch strings.ToLower(value) {
	case "system", "":
		return "system", nil
	case "en", "en-us", "en-gb":
		return "en", nil
	case "pt", "pt-br", "pt-pt":
		return "pt", nil
	default:
		return "", fmt.Errorf("unsupported language %q; use system, en or pt", value)
	}
}
func Resolve(preference, system string) string {
	if preference == "pt" || preference == "en" {
		return preference
	}
	if system == "pt" {
		return "pt"
	}
	return "en"
}
func language(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "pt" || strings.HasPrefix(locale, "pt_") || strings.HasPrefix(locale, "pt-") {
		return "pt"
	}
	return "en"
}
func Detect(getenv func(string) string, native func() string) string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		if value := getenv(key); value != "" {
			return language(strings.Split(value, ":")[0])
		}
	}
	return language(native())
}
func System() string {
	return Detect(os.Getenv, func() string {
		if runtime.GOOS != "darwin" {
			return ""
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		data, e := exec.CommandContext(ctx, "/usr/bin/defaults", "read", "-g", "AppleLanguages").Output()
		if e != nil {
			return ""
		}
		fields := strings.FieldsFunc(string(data), func(r rune) bool { return r == '(' || r == ')' || r == ',' || r == '"' || r == '\n' || r == ' ' })
		if len(fields) > 0 {
			return fields[0]
		}
		return ""
	})
}
func ConfigPath() (string, error) {
	dir, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(dir, "brew-sentinel", "settings.json"), nil
}
func Load(path string) (string, error) {
	data, e := os.ReadFile(path)
	if errors.Is(e, os.ErrNotExist) {
		return "system", nil
	}
	if e != nil {
		return "system", e
	}
	var cfg struct {
		Language string `json:"language"`
	}
	if e = json.Unmarshal(data, &cfg); e != nil {
		return "system", e
	}
	return Preference(cfg.Language)
}
func Save(path, preference string) error {
	p, e := Preference(preference)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	data, _ := json.MarshalIndent(struct {
		Language string `json:"language"`
	}{p}, "", "  ")
	tmp, e := os.CreateTemp(filepath.Dir(path), ".settings-*")
	if e != nil {
		return e
	}
	defer os.Remove(tmp.Name())
	if _, e = tmp.Write(append(data, '\n')); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Close(); e != nil {
		return e
	}
	return os.Rename(tmp.Name(), path)
}
