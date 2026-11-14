package brew

import (
	"context"
	"encoding/json"
	"strings"

	"brew-sentinel/internal/i18n"
)

type Candidate struct {
	Kind, Name, FullName, Description, Version, Homepage string
	Installed                                            bool
}

func ParseSearch(text, kind string) []Candidate {
	var out []Candidate
	seen := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "==>") {
			continue
		}
		for _, name := range strings.Fields(line) {
			if packageName.MatchString(name) && !seen[name] {
				out = append(out, Candidate{Kind: kind, Name: name, FullName: name})
				seen[name] = true
			}
		}
	}
	return out
}
func (c *Client) Search(ctx context.Context, kind, query string) ([]Candidate, error) {
	if kind != "formula" && kind != "cask" {
		return nil, i18n.Errorf("invalid package type")
	}
	if !packageName.MatchString(query) {
		return nil, i18n.Errorf("use part of a package name, without spaces or options")
	}
	out, e := c.Read(ctx, "search", "--"+kind, query)
	if e != nil {
		return nil, e
	}
	return ParseSearch(string(out), kind), nil
}
func ParseCandidate(data []byte, kind, name string) (Candidate, error) {
	var doc struct {
		Formulae []formula `json:"formulae"`
		Casks    []cask    `json:"casks"`
	}
	if e := json.Unmarshal(data, &doc); e != nil {
		return Candidate{}, e
	}
	if kind == "formula" && len(doc.Formulae) == 1 {
		f := doc.Formulae[0]
		full := f.FullName
		if full == "" {
			full = f.Name
		}
		return Candidate{Kind: kind, Name: f.Name, FullName: full, Description: f.Desc, Version: f.Versions.Stable, Homepage: f.Homepage, Installed: len(f.Installed) > 0}, nil
	}
	if kind == "cask" && len(doc.Casks) == 1 {
		c := doc.Casks[0]
		full := c.FullToken
		if full == "" {
			full = c.Token
		}
		return Candidate{Kind: kind, Name: c.Token, FullName: full, Description: c.Desc, Version: c.Version, Homepage: c.Homepage, Installed: len(c.Installed) > 0 && string(c.Installed) != "null" && string(c.Installed) != `""` && string(c.Installed) != "[]"}, nil
	}
	return Candidate{}, i18n.Errorf("could not resolve %s as %s", name, kind)
}
func (c *Client) Describe(ctx context.Context, kind, name string) (Candidate, error) {
	if _, e := ActionArgs("install", kind, name); e != nil {
		return Candidate{}, e
	}
	data, e := c.Read(ctx, "info", "--json=v2", "--"+kind, name)
	if e != nil {
		return Candidate{}, e
	}
	return ParseCandidate(data, kind, name)
}
