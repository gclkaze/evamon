package port

// Renderer is a factory/creator of UI primitives (windows/tabs/etc).
// Your app/service asks for windows through this interface.

type Renderer interface {
	NewExecutionWindow(title string) (ExecutionWindow, error)
	Layout() Layout
	Run()
	Controls() Controls
	Actions() DiagramActions
	ChartRegistry() *ChartRegistry
}
