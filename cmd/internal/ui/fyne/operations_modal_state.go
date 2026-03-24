package fynerenderer

import (
	"fyne.io/fyne/v2"
	"github.com/gclkaze/evamon/cmd/internal/models"
)

type OperationsModalState struct {
	ParentWindow        fyne.Window
	AvailableComponents []models.FilterComponent
	SelectedItems       []*models.FilterItem
	Files               []*models.ActionFile
	ActiveFile          *models.ActionFile
}

func NewOperationsModalState(
	parent fyne.Window,
	components []models.FilterComponent,
	files []*models.ActionFile,
) *OperationsModalState {
	return &OperationsModalState{
		ParentWindow:        parent,
		AvailableComponents: components,
		SelectedItems:       make([]*models.FilterItem, 0),
		Files:               files,
		ActiveFile:          nil,
	}
}

func (ms *OperationsModalState) pruneAssignments() {
	validPaths := make(map[string]bool)
	for _, f := range ms.Files {
		validPaths[f.Path] = true
	}
	for _, item := range ms.SelectedItems {
		filtered := item.Files[:0]
		for _, f := range item.Files {
			if validPaths[f.Path] {
				filtered = append(filtered, f)
			}
		}
		item.Files = filtered
	}
}

func (ms *OperationsModalState) isFileAssigned(file *models.ActionFile, item *models.FilterItem) bool {
	return item.HasFile(file.Path)
}
