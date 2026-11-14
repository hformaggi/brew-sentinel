package brew

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"brew-sentinel/internal/i18n"
	"brew-sentinel/internal/inventory"
)

type dependency struct {
	FullName string `json:"full_name"`
	Direct   *bool  `json:"declared_directly"`
}
type receipt struct {
	Requested    *bool        `json:"installed_on_request"`
	Dependencies []dependency `json:"runtime_dependencies"`
}
type formula struct {
	Outdated bool   `json:"outdated"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Desc     string `json:"desc"`
	Homepage string `json:"homepage"`
	Versions struct {
		Stable string `json:"stable"`
	} `json:"versions"`
	Installed []struct {
		Version string       `json:"version"`
		Runtime []dependency `json:"runtime_dependencies"`
	} `json:"installed"`
	Dependencies []string `json:"dependencies"`
	Pinned       bool     `json:"pinned"`
	Deprecated   bool     `json:"deprecated"`
	Disabled     bool     `json:"disabled"`
}
type cask struct {
	Outdated  bool            `json:"outdated"`
	Token     string          `json:"token"`
	FullToken string          `json:"full_token"`
	Desc      string          `json:"desc"`
	Homepage  string          `json:"homepage"`
	Version   string          `json:"version"`
	Installed json.RawMessage `json:"installed"`
	Depends   struct {
		Formula []string `json:"formula"`
		Cask    []string `json:"cask"`
	} `json:"depends_on"`
}

func ParseInventory(data []byte, cellar string) (inventory.Snapshot, error) {
	var doc struct {
		Formulae []formula `json:"formulae"`
		Casks    []cask    `json:"casks"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return inventory.Snapshot{}, err
	}
	if doc.Formulae == nil && doc.Casks == nil {
		return inventory.Snapshot{}, i18n.Errorf("missing inventory arrays")
	}
	s := inventory.Snapshot{CollectedAt: time.Now()}
	for _, f := range doc.Formulae {
		name := f.FullName
		if name == "" {
			name = f.Name
		}
		p := inventory.Package{ID: "formula:" + name, Name: f.Name, Kind: "formula", FullName: name, Description: f.Desc, Homepage: f.Homepage, Available: f.Versions.Stable, Reason: "unknown", Outdated: f.Outdated, Pinned: f.Pinned, Deprecated: f.Deprecated, Disabled: f.Disabled, DependencySource: "installation receipt"}
		known := false
		directKnown := true
		for _, v := range f.Installed {
			for _, d := range v.Runtime {
				if d.FullName != "" {
					p.Observed = unique(p.Observed, "formula:"+d.FullName)
				}
			}
			keg := filepath.Join(cellar, filepath.Base(name), v.Version)
			p.Versions = append(p.Versions, inventory.InstalledVersion{Version: v.Version, KegPath: keg})
			raw, e := os.ReadFile(filepath.Join(keg, "INSTALL_RECEIPT.json"))
			var r receipt
			if e != nil || json.Unmarshal(raw, &r) != nil {
				directKnown = false
				continue
			}
			if r.Requested != nil {
				known = true
				if *r.Requested {
					p.Explicit = true
				}
			}
			if r.Dependencies == nil {
				directKnown = false
			}
			for _, d := range r.Dependencies {
				p.Observed = unique(p.Observed, "formula:"+d.FullName)
				if d.Direct == nil {
					directKnown = false
				}
				if d.Direct != nil && *d.Direct {
					p.Dependencies = unique(p.Dependencies, "formula:"+d.FullName)
				}
			}
		}
		if known {
			p.Reason = "dependency"
			if p.Explicit {
				p.Reason = "explicit"
			}
		}
		if !directKnown {
			p.DependencySource = "current recipe"
			if len(p.Dependencies) > 0 {
				p.DependencySource = "installation receipt (partial; current recipe separate)"
			} else {
				for _, d := range f.Dependencies {
					p.Dependencies = unique(p.Dependencies, "formula:"+d)
				}
			}
		}
		s.Packages = append(s.Packages, p)
	}
	for _, c := range doc.Casks {
		name := c.FullToken
		if name == "" {
			name = c.Token
		}
		p := inventory.Package{ID: "cask:" + name, Name: c.Token, Kind: "cask", FullName: name, Description: c.Desc, Homepage: c.Homepage, Available: c.Version, Reason: "unknown", Outdated: c.Outdated, DependencySource: "current recipe"}
		var installed string
		var versions []string
		if json.Unmarshal(c.Installed, &installed) == nil {
			if installed != "" {
				versions = []string{installed}
			}
		} else {
			_ = json.Unmarshal(c.Installed, &versions)
		}
		for _, v := range versions {
			p.Versions = append(p.Versions, inventory.InstalledVersion{Version: v})
		}
		for _, d := range c.Depends.Formula {
			p.Dependencies = append(p.Dependencies, "formula:"+d)
		}
		for _, d := range c.Depends.Cask {
			p.Dependencies = append(p.Dependencies, "cask:"+d)
		}
		s.Packages = append(s.Packages, p)
	}
	return s, nil
}
func unique(a []string, s string) []string {
	for _, v := range a {
		if v == s {
			return a
		}
	}
	return append(a, s)
}
func (c *Client) Inventory(ctx context.Context) (inventory.Snapshot, error) {
	path, e := c.Read(ctx, "--cellar")
	if e != nil {
		return inventory.Snapshot{}, e
	}
	data, e := c.Read(ctx, "info", "--json=v2", "--installed")
	if e != nil {
		return inventory.Snapshot{}, e
	}
	return ParseInventory(data, strings.TrimSpace(string(path)))
}
