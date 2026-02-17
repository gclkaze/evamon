package viewproject

// ProjectBase is shared metadata for any on-disk UI project file.
// Embed it to simulate "super class" behavior in Go.
type ProjectBase struct {
	ID string `json:"ID"`
	// Absolute path of the project file (filled by loader)
	ProjectPath string `json:"-"`
}
