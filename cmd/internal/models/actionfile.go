package models

type ActionFile struct {
	Path string
}

func NewActionFile(path string) *ActionFile {
	return &ActionFile{Path: path}
}
