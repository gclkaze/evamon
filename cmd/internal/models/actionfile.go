package models

type ActionFile struct {
	Path string `json:"path"`
}

func NewActionFile(path string) *ActionFile {
	return &ActionFile{Path: path}
}
