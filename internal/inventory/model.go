package inventory

import "time"

type InstalledVersion struct{ Version, KegPath string }
type Package struct {
	ID, Name, Kind, FullName, Description, Homepage, Available, Reason, DependencySource string
	Explicit, Pinned, Deprecated, Disabled, Outdated                                     bool
	Dependencies, Observed                                                               []string
	Versions                                                                             []InstalledVersion
}
type Snapshot struct {
	Packages    []Package
	CollectedAt time.Time
}
