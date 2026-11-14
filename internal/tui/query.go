package tui

import (
	"brew-sentinel/internal/inventory"
	"strings"
)

func Filter(packages []inventory.Package, query string) []inventory.Package {
	var out []inventory.Package
	query = strings.ToLower(query)
	for _, p := range packages {
		matches := true
		for _, token := range strings.Fields(query) {
			switch {
			case strings.HasPrefix(token, "type:"):
				matches = matches && p.Kind == strings.TrimPrefix(token, "type:")
			case token == "outdated:true":
				matches = matches && p.Outdated
			case strings.HasPrefix(token, "reason:"):
				matches = matches && p.Reason == strings.TrimPrefix(token, "reason:")
			default:
				matches = matches && strings.Contains(strings.ToLower(p.Name+" "+p.Description+" "+p.FullName), token)
			}
		}
		if matches {
			out = append(out, p)
		}
	}
	return out
}
func Versions(p inventory.Package) string {
	var v []string
	for _, k := range p.Versions {
		v = append(v, k.Version)
	}
	return strings.Join(v, ", ")
}
