package models

type DiagramOwner interface {
	GetProjectBase() *ProjectBase
	GetOwnerKind() string
	Save() error
}
