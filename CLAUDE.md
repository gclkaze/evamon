# Evamon — CLAUDE.md

## Project Overview

Evamon is a CLI real-time monitoring tool that connects to the **Evacron** job scheduler via WebSocket. It reads live metrics from running jobs and renders them as interactive charts (line chart, bar chart, boolean fill) inside a Fyne desktop GUI. Users define views and dashboards in JSON files stored in `~/.evamon/lib/`.

---

## Architecture: Port/Adapter Pattern

The UI layer is fully decoupled from the Fyne framework through a port/adapter split:
```
cmd/internal/ui/port/           ← abstract interfaces (no GUI dependency)
cmd/internal/ui/diagrams/port/  ← diagram-specific abstract interfaces
cmd/internal/ui/fyne/           ← Fyne adapter: Renderer, Layout, Controls, Actions
cmd/internal/ui/diagrams/fyne/  ← Fyne diagram widgets and drawers
```

A `Renderer` (port) holds `Layout`, `Controls`, `Actions`, and a `ChartRegistry`. The Fyne adapter implements all of these. The rest of the app only talks to ports.

---

## Package Map
```
cmd/
├── root.go                             # Cobra CLI root; wires services
├── job/                                # Job management subcommands
├── view/                               # View/dashboard render subcommand
└── internal/
    ├── app/evamon.go                   # Application orchestrator
    ├── models/                         # Core data models (Diagram, Filter, SetupItem, messages)
    ├── services/
    │   ├── widget_service.go           # Creates diagram widgets; dispatches incoming data
    │   └── view_service.go             # WebSocket listener; renders projects
    ├── viewproject/                    # ViewProject and DashboardProject JSON models
    ├── ui/
    │   ├── port/                       # Abstract UI interfaces
    │   │   ├── renderer.go             # Renderer (top-level factory)
    │   │   ├── ui.go                   # UIObject, Layout
    │   │   ├── controls.go             # Controls (buttons, checkboxes, menus)
    │   │   ├── actions.go              # DiagramActions (Maximize, Filters, Download)
    │   │   ├── chart_registry.go       # ChartRegistry: diagramID → DiagramUIRefs
    │   │   └── diagramuirefs.go        # DiagramUIRefs: all UI parts for one diagram
    │   ├── fyne/                       # Fyne adapter
    │   │   ├── renderer.go
    │   │   ├── layout.go
    │   │   ├── controls.go
    │   │   └── diagram_actions.go      # Maximize, Filters, Download implementations
    │   ├── diagrams/
    │   │   ├── port/                   # DiagramWidget, EvaWidget, Factory interfaces
    │   │   └── fyne/                   # Concrete widgets and drawers
    │   │       ├── factory.go
    │   │       ├── toolbar_factory.go  # Per-diagram-type toolbar factories (see TODO)
    │   │       ├── barchart_widget.go
    │   │       ├── barchart_drawer.go
    │   │       ├── linechart_widget.go
    │   │       ├── linechart_drawer.go
    │   │       ├── linechart_tooltip.go
    │   │       ├── maximize_adapter.go
    │   │       ├── bool_fill_widget.go
    │   │       └── boolean_fill_drawer.go
    │   ├── data/
    │   │   ├── imultiseriesdata.go     # IMultiSeriesData interface
    │   │   ├── multiseriesring.go      # Ring-buffer implementation with filter evaluation
    │   │   ├── aggregation_tree.go     # AggregationTree, Bucket, BucketIterator
    │   │   └── aggregation_func.go    # TimePeriod, AggregationFunction, Extract()
    │   ├── dashboard/builder.go        # Builds dashboard UI from DashboardProject
    │   └── factory/factory.go          # Renderer factory (selects Fyne or future Web)
    ├── config/                         # Endpoint resolution
    ├── auth/                           # Token loading
    ├── wsclient/                       # WebSocket client
    └── fs/                             # File I/O helpers
pkg/utils/                             # Expression evaluator, string/file utilities
```

---

## Data Flow
```
Evacron (WebSocket)
  ↓  MetricsMsg JSON
ViewService.listenForMessages()
  ↓  (jobID, variable, timestamp, value)
WidgetService.DispatchValue()
  ↓
JobRouter.Push()          [models/ui/jobrouter.go]
  ↓
VariableContainer.Push()  [models/ui/variablecontainer.go]
  ↓  (fans out to all widgets that care about this variable)
DiagramWidget.Push(at, val)
  ↓
MultiSeriesRing.Append()  [ui/data/multiseriesring.go]
  — stores time + values in ring buffers
  — evaluates filter expressions on the new point
  ↓
drawer.Push(at, val)      (BarChartDrawer / LineChartDrawer)
  — calls redrawWithAxis() or refreshes raster
  — calls refreshRoot() / refreshRaster()
    → normal mode: CanvasForObject(d.root).Refresh()
    → maximize mode: adapter.Refresh() → reliable canvas path
```

---

## Diagram Widget Architecture

Each chart type has two layers:

| Layer | Bar | Line | Bool |
|---|---|---|---|
| **Widget** (Fyne BaseWidget) | `BarChartWidget` | `LineChartWidget` | `BoolFillWidget` |
| **Drawer** (rendering logic) | `BarChartDrawer` | `LineChartDrawer` | `BooleanFillDrawer` |

- The **Widget** is what Fyne's container system manages. It owns hover events (`desktop.Hoverable`), tooltips, and zoom controls.
- The **Drawer** owns all canvas primitives (`d.root *fyne.Container`, `d.raster *canvas.Raster`, rectangles, labels). It performs all drawing and holds the data reference.
- The widget's `CreateRenderer()` returns `[drawer.Root(), tooltip.Object()]` as its renderer objects.

### Tooltip

`chartTooltip` (in `linechart_tooltip.go`) is shared by both chart types. It is an overlay container positioned with absolute coordinates inside the widget's renderer object list. It shows on `MouseMoved` via `HitTestX()` on the drawer.

### Maximize

`DiagramActionHandler.Maximize()` in `diagram_actions.go`:
1. Removes the chart widget from its slot container; shows a placeholder.
2. Calls `widget.MaximizeView(1200, 800)` which:
   - Pre-sizes `d.root` (so `readRootSize()` doesn't exit early on first push).
   - Creates a `maximizeAdapter` widget wrapping `d.root`.
   - Installs `drawer.refreshHook = func() { adapter.Refresh() }` so every drawer refresh goes through `CanvasForObject(adapter)` (reliable — adapter IS the window content) rather than `CanvasForObject(d.root)` (unreliable when `d.root` lives inside a widget renderer on a different canvas).
3. Sets `maximizeAdapter` as the new window's content.
4. On window close: clears the hook, restores the widget to the slot.

The `maximizeAdapter`'s `renderer.Refresh()` uses `CanvasForObject(adapter)` to get the canvas, then calls `c.Refresh(d.root)` directly — bypassing the stale `CanvasForObject(d.root)` lookup entirely.

---

## Key Types

| Type | File | Role |
|---|---|---|
| `DiagramUIRefs` | `ui/port/diagramuirefs.go` | Holds all UI components for one diagram: Chart, ChartSlot, Toolbar, Legend, Filter controls. `ChartSlot` is the `Max` container wrapping the chart widget (allows placeholder swap on maximize). `Maximized bool` guards against duplicate windows. |
| `ChartRegistry` | `ui/port/chart_registry.go` | Global map: `diagramID → *DiagramUIRefs`. Used by toolbar callbacks and diagram actions. |
| `MultiSeriesRing` | `ui/data/multiseriesring.go` | Fixed-capacity ring buffer storing `times []time.Time` and `values [][]int32` for N variables. Also stores `ConditionResults` per data point for filter overlays. |
| `maximizeAdapter` | `diagrams/fyne/maximize_adapter.go` | Thin `BaseWidget` wrapping `d.root`. Its `Layout(size)` delegates to the drawer's full resize logic. Its renderer's `Refresh()` uses the adapter's own canvas to refresh `d.root`. |
| `AggregationTree` | `ui/data/aggregation_tree.go` | Nested-map tree of pre-aggregated buckets built from historic data. Immutable after construction. Queried via `IterateLevel()` or level-specific iterators. |
| `BucketIterator` | `ui/data/aggregation_tree.go` | Typed iterator over a chronologically sorted `[]Bucket` slice. Not goroutine-safe. |
| `HistoricDataStore` | `ui/data/` | Replaces `MultiSeriesRing` in historic mode. Holds the raw `[]DataPoint` and the built `*AggregationTree`. Never mixed with `MultiSeriesRing` in the same `DiagramUIRefs`. |

---

## Filter System

- A `SetupItem` can have a `Filter` with a `FilterSetup` (list of `FilterComponent` expressions + `FilterMode AND/OR`).
- Expressions are evaluated by `pkg/utils.RunExpression()` using the `tafexpr` evaluator.
- `MultiSeriesRing.ApplyFilterChanges()` replays expressions over all stored data points (backfill).
- Filter results (`ConditionResults[i].Results[componentID] bool`) are used by drawers to colour/overlay bars or line segments.

---

## Historic Data Aggregation (Bar Chart)

### Overview

Aggregation is available **only in historic view mode** (no WebSocket/real-time). Data is fetched from Evacron via HTTP (full range or batched) and then aggregated into a tree structure before rendering. The bar chart drawer receives pre-aggregated buckets — it never aggregates on the fly.

### Data Flow
```
Evacron HTTP endpoint
  ↓  []DataPoint (time.Time + []float64 per series)
BuildTree(points) → *AggregationTree
  ↓
AggregationTree.IterateLevel(period) → *BucketIterator
  ↓
BarChartDrawer (renders one bar per bucket)
```

### Core Types
```go
// Bucket holds aggregated values for one time slot at any level.
// Avg is computed on demand via AggregationFunction.Extract().
type Bucket struct {
    Start  time.Time
    Period TimePeriod
    Count  int
    Sum    []float64 // per series
    Min    []float64
    Max    []float64
}

type AggregationTree struct {
    Years map[int]*YearBucket
}

type YearBucket struct {
    Bucket
    Months map[time.Month]*MonthBucket
}

type MonthBucket struct {
    Bucket
    Weeks map[int]*WeekBucket
}

type WeekBucket struct {
    Bucket
    Days map[int]*DayBucket
}

type DayBucket struct {
    Bucket
    Hours map[int]*HourBucket
}

type HourBucket struct {
    Bucket
    Minutes map[int]*MinuteBucket
}

type MinuteBucket struct {
    Bucket
    Seconds map[int]*Bucket
}
```

### Supported Periods
```go
type TimePeriod int

const (
    PeriodSecond  TimePeriod = iota
    PeriodMinute
    PeriodHour
    PeriodDay
    PeriodWeek
    PeriodMonth
    PeriodQuarter // 3-month
    PeriodYear
)

// Seconds() returns the nominal duration of the period in seconds.
// Nominal values: Minute=60, Hour=3600, Day=86400, Week=604800,
// Month=2_592_000 (30d), Quarter=7_776_000 (90d), Year=31_536_000 (365d).
func (p TimePeriod) Seconds() float64

// Minutes(), Hours(), Days(), Weeks(), Months(), Years() derive from Seconds().
```

### Aggregation Functions
```go
type AggregationFunction int

const (
    FuncSum AggregationFunction = iota
    FuncAvg
    FuncMin
    FuncMax
    FuncCount
    FuncPerSecond
    FuncPerMinute
    FuncPerHour
    FuncPerDay
    FuncPerWeek
    FuncPerMonth
    FuncPerYear
)

// Extract selects the correct value from a Bucket for a given series index.
func (f AggregationFunction) Extract(b Bucket, series int) float64 {
    switch f {
    case FuncSum:       return b.Sum[series]
    case FuncAvg:       return b.Sum[series] / float64(b.Count)
    case FuncMin:       return b.Min[series]
    case FuncMax:       return b.Max[series]
    case FuncCount:     return float64(b.Count)
    case FuncPerSecond: return b.Sum[series] / b.Period.Seconds()
    case FuncPerMinute: return b.Sum[series] / b.Period.Minutes()
    case FuncPerHour:   return b.Sum[series] / b.Period.Hours()
    case FuncPerDay:    return b.Sum[series] / b.Period.Days()
    case FuncPerWeek:   return b.Sum[series] / b.Period.Weeks()
    case FuncPerMonth:  return b.Sum[series] / b.Period.Months()
    case FuncPerYear:   return b.Sum[series] / b.Period.Years()
    }
    return 0
}
```

### Building the Tree

Single pass over raw points. Each point is accumulated into every ancestor node simultaneously:
```go
func BuildTree(points []DataPoint) *AggregationTree
```

Each level's `getOrCreate*` helper initialises the child map and `Bucket` fields lazily. `accumulate(values []float64)` updates `Count`, `Sum`, `Min`, `Max` for that node.

### Querying
```go
// BucketsAtLevel flattens the entire tree at the requested level into a
// chronologically sorted slice — ready for the drawer or for export.
func (t *AggregationTree) BucketsAtLevel(period TimePeriod) []Bucket
```

### Bucket Iterator

A typed iterator so callers never traverse the nested map structure directly.
```go
type BucketIterator struct {
    buckets []Bucket
    index   int
}

func (it *BucketIterator) HasNext() bool  { return it.index < len(it.buckets) }
func (it *BucketIterator) Next() Bucket   { b := it.buckets[it.index]; it.index++; return b }
func (it *BucketIterator) Reset()         { it.index = 0 }
```

Level-specific factory methods on `AggregationTree` (each collects and sorts by `Bucket.Start` before returning):
```go
func (t *AggregationTree) IterateYears() *BucketIterator
func (t *AggregationTree) IterateMonths(year int) *BucketIterator
func (t *AggregationTree) IterateWeeks(year int, month time.Month) *BucketIterator
func (t *AggregationTree) IterateDays(year int, month time.Month) *BucketIterator
func (t *AggregationTree) IterateHours(year int, month time.Month, day int) *BucketIterator
func (t *AggregationTree) IterateMinutes(year int, month time.Month, day int, hour int) *BucketIterator
func (t *AggregationTree) IterateSeconds(year int, month time.Month, day int, hour int, minute int) *BucketIterator

// IterateLevel flattens the entire tree at the requested period — the common case for the drawer.
func (t *AggregationTree) IterateLevel(period TimePeriod) *BucketIterator
```

#### Usage Pattern
```go
it := tree.IterateLevel(PeriodDay)
for it.HasNext() {
    b := it.Next()
    value := aggFunc.Extract(b, seriesIndex)
    // draw bar at b.Start with height proportional to value
}
```

### Toolbar Integration (Historic Mode)

The bar chart toolbar in historic mode exposes two additional selectors:

- **Period selector** — dropdown: Second / Minute / Hour / Day / Week / Month / Quarter / Year
- **Function selector** — dropdown: Sum / Avg / Min / Max / Count / Per Second / Per Minute / Per Hour / Per Day / Per Week / Per Month / Per Year

Both selectors trigger a `RebuildToolbar` on change and a drawer refresh via `DiagramUIRefs`. These controls are **absent in real-time mode**.

### Conventions

- `Bucket.Sum/Min/Max` are `[]float64` regardless of source `int32` values — promotes to float64 to avoid overflow on Sum and to support Avg cleanly.
- `PerX` rate functions use the **nominal bucket duration**, not the actual span of points within the bucket — consistent behaviour over sparse data.
- `AggregationTree` is built once after the HTTP fetch and is **immutable** after that. Changing period or function does not rebuild the tree — only `IterateLevel()` / `BucketsAtLevel()` is re-called.
- Historic mode diagrams carry a `HistoricDataStore` instead of a `MultiSeriesRing`. The two are never mixed in the same `DiagramUIRefs`.
- Iterators are **not goroutine-safe** — iterator state (`index`) is per-caller. Each render pass should obtain a fresh iterator via `IterateLevel()`, or call `Reset()` for a deliberate multi-pass render (e.g. first pass for axis scaling, second pass for drawing bars).
- Level-specific iterators return an **empty iterator** (not nil, not an error) when the requested parent key doesn't exist — callers need no nil checks.

---

## Build & Run
```bash
# Build
go build ./...
make build          # produces build/evamon.exe

# Run
evamon view <view-name>
evamon view --dashboard <dashboard-name>

# Config
~/.evamon/application.properties   # server host/port, token path
~/.evamon/lib/                      # saved view and dashboard JSON files
```

**Key dependencies:** `fyne.io/fyne/v2` v2.7.2, `github.com/coder/websocket`, `github.com/gclkaze/tafexpr`, `github.com/spf13/cobra`.

---

## Conventions

- All Fyne UI mutations must run on the main goroutine. Use `UI(fn)` (`fyne.Do`) inside drawers.
- `BaseWidget.Resize()` **must** call `w.BaseWidget.Resize(size)` — omitting it leaves `baseObject.size = {0,0}` and Fyne's hit-tester never dispatches hover/tap events.
- Use `container.RemoveAll()` + `container.Add()` (not direct `Objects` slice mutation) when the canvas needs to track object tree changes.
- `refreshHook` on drawers: nil in normal mode (uses `d.root.Refresh()` / `d.raster.Refresh()`); set to `adapter.Refresh()` during maximize to route through the adapter's reliable canvas reference.
- `ChartSlot` is always a `container.NewMax(chartWidget)` — never put the raw widget directly into the dashboard layout. This allows the slot's content to be swapped without rebuilding the layout.

---

## TODO

### Refactor: Per-Diagram-Type Toolbar Factory

**Current state**: `DefaultDiagramToolbarFactory` is a single factory holding only a `Renderer` field. It builds a toolbar for all diagram types uniformly.

**Problem**: Not all diagram types need a toolbar. `BoolFillWidget` in particular has no meaningful toolbar actions — period/function selectors, maximize, filters, and download are all irrelevant or inapplicable. Forcing a toolbar onto it is noise.

**Goal**: Replace `DefaultDiagramToolbarFactory` with a per-diagram-type factory:
```go
type DiagramToolbarFactory interface {
    BuildToolbar(refs *DiagramUIRefs) fyne.CanvasObject
}

// Implementations (in cmd/internal/ui/diagrams/fyne/toolbar_factory.go):
type BarChartToolbarFactory  struct{ Renderer Renderer } // period + function + maximize + filters + download
type LineChartToolbarFactory struct{ Renderer Renderer } // maximize + filters + download
type BoolFillToolbarFactory  struct{}                    // returns nil — no toolbar
```

`DiagramUIRefs.Toolbar` must tolerate nil gracefully — the dashboard layout must not allocate space for it when absent.

**Impact**: `ChartRegistry`, `DiagramUIRefs`, and the `RebuildToolbar` closure are unaffected — the factory is only invoked at construction time and on toolbar rebuild triggers.