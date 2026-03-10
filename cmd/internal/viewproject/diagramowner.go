package viewproject

type DiagramOwner interface {
	GetProjectBase() *ProjectBase
	GetOwnerKind() string
	Save() error
}
