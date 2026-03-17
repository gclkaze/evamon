package fynerenderer

import (
	"os"
	"path/filepath"
	"sync"

	fyne "fyne.io/fyne/v2"
)

// IconLoader loads image files from the filesystem and returns them as Fyne
// static resources. Loaded resources are cached by absolute path.
type IconLoader struct {
	mu    sync.RWMutex
	cache map[string]fyne.Resource
}

func NewIconLoader() *IconLoader {
	return &IconLoader{cache: make(map[string]fyne.Resource)}
}

// Load reads the file at path and returns a Fyne resource.
// Supported formats are anything Fyne accepts (PNG, SVG, JPEG, …).
// Results are cached; subsequent calls with the same path return the cached resource.
func (l *IconLoader) Load(path string) (fyne.Resource, error) {
	l.mu.RLock()
	if r, ok := l.cache[path]; ok {
		l.mu.RUnlock()
		return r, nil
	}
	l.mu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	res := fyne.NewStaticResource(name, data)

	l.mu.Lock()
	l.cache[path] = res
	l.mu.Unlock()

	return res, nil
}

// MustLoad is like Load but panics on error. Useful for embedded/bundled assets
// whose absence is a programming error.
func (l *IconLoader) MustLoad(path string) fyne.Resource {
	r, err := l.Load(path)
	if err != nil {
		panic("IconLoader: " + err.Error())
	}
	return r
}

// Invalidate removes a cached resource so the file is re-read on the next Load.
func (l *IconLoader) Invalidate(path string) {
	l.mu.Lock()
	delete(l.cache, path)
	l.mu.Unlock()
}
