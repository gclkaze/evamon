# Evamon — CLAUDE.md

## Project Overview
Evamon is a CLI real-time monitoring tool connecting to the Evacron job
scheduler via WebSocket. It renders live metrics as interactive charts
(line, bar, boolean fill) inside a Fyne desktop GUI. Dashboards are
defined in JSON files stored in `~/.evamon/lib/`.

---

## Architecture: Port/Adapter Pattern
```
cmd/internal/ui/port/           ← abstract interfaces (no GUI dependency)
cmd/internal/ui/diagrams/port/  ← diagram-specific abstract interfaces
cmd/internal/ui/fyne/           ← Fyne adapter: Renderer, Layout, Controls, Actions
cmd/internal/ui/diagrams/fyne/  ← Fyne diagram widgets and drawers
```
A `Renderer` (port) holds `Layout`, `Controls`, `Actions`, and a
`ChartRegistry`. The rest of the app only talks to ports — never to
Fyne directly.

---

## Package Map
```
cmd/
├── root.go                             # Cobra CLI root; wires services
└── internal/
    ├── app/evamon.go                   # Application orchestrator
    ├── models/                         # Core data models
    │   ├── triggeroperationmsg.go      # TriggerOperationMsg + lifecycle + ToWSMessage()
    │   ├── execution_output.go         # ExecutionOutput (streamed log lines from Evacron)
    │   ├── execution_result.go         # ExecutionResult (final success/fail from Evacron)
    │   ├── execution_log.go            # ExecutionLogEntry, TriggerExecution
    │   ├── ring_buffer.go              # Generic RingBuffer[T]
    │   └── ws.go                       # WSMessage envelope
    ├── services/
    │   ├── widget_service.go           # Diagram widgets; trigger protocol; handleStreamLine
    │   ├── trigger_stream_tracker.go   # Tracks in-flight triggers; notifies coordinator
    │   ├── trigger_execution_coordinator.go  # OnStart/OnLine/OnDone → registry
    │   ├── trigger_execution_builder.go      # Accumulates lines into TriggerExecution
    │   ├── execution_log_registry.go         # map[DiagramID][RuleID]→RingBuffer
    │   ├── log_parser.go               # parseLogLine — extracts ParsedMsg from JSON lines
    │   └── view_service.go             # WebSocket listener; renders projects
    ├── ui/
    │   ├── port/                       # Abstract UI interfaces
    │   │   ├── renderer.go             # Renderer (top-level factory)
    │   │   ├── ui.go                   # UIObject, Layout
    │   │   ├── controls.go             # Controls, IconSnapshot
    │   │   ├── actions.go              # DiagramActions
    │   │   ├── chart_registry.go       # diagramID → DiagramUIRefs
    │   │   └── diagramuirefs.go        # All UI parts for one diagram
    │   ├── fyne/                       # Fyne adapter
    │   │   ├── renderer.go
    │   │   ├── layout.go
    │   │   ├── controls.go
    │   │   ├── diagram_actions.go      # Maximize, Filters, Download, Snapshot, Operations
    │   │   └── operations/             # Operations modal (see Operations section)
    │   ├── diagrams/
    │   │   ├── port/                   # DiagramWidget, EvaWidget, Factory interfaces
    │   │   └── fyne/                   # Concrete widgets and drawers
    │   │       ├── factory.go
    │   │       ├── barchart_widget.go / barchart_drawer.go
    │   │       ├── linechart_widget.go / linechart_drawer.go / linechart_tooltip.go
    │   │       ├── bool_fill_widget.go / boolean_fill_drawer.go
    │   │       └── maximize_adapter.go
    │   ├── data/
    │   │   ├── imultiseriesdata.go     # IMultiSeriesData interface + TriggerSendFunc
    │   │   ├── multiseriesring.go      # Ring buffer + filter + trigger evaluation
    │   │   ├── singularseriesdata.go   # Single-value IMultiSeriesData for BoolFill
    │   │   ├── aggregation_tree.go     # AggregationTree + BucketIterator
    │   │   └── aggregation_func.go     # TimePeriod, AggregationFunction, Extract()
    │   ├── execution_log/              # Execution Log Panel UI
    │   │   ├── execution_log_panel.go
    │   │   ├── execution_diagram_tab.go
    │   │   ├── execution_rule_tab.go
    │   │   ├── execution_list_view.go
    │   │   ├── execution_detail_view.go
    │   │   ├── execution_log_toolbar.go
    │   │   ├── execution_log_parser.go # ParseLogLine → ParsedLogLine + OperationType
    │   │   └── log_style_config.go     # Loads config/log_styles.json
    │   ├── dashboard/builder.go        # Builds dashboard UI from DashboardProject
    │   └── factory/
    │       ├── factory.go
    │       └── diagram_toolbar_factory.go
    ├── config/                         # Endpoint resolution
    ├── auth/                           # Token loading
    ├── wsclient/
    │   ├── wsclient.go                 # SendJSON, ReadJSON, SendText, ReadText
    │   └── wserrors.go
    └── fs/                             # File I/O helpers
pkg/utils/                             # Expression evaluator, string/file utilities
```

---

## Data Flow
```
Evacron WebSocket
  ↓ MetricsMsg JSON
ViewService.listenForMessages()
  ↓ (jobID, variable, timestamp, value)
WidgetService.DispatchValue()
  ↓
JobRouter → VariableContainer → DiagramWidget.Push(at, val)
  ↓
MultiSeriesRing.Append() / SingularSeriesData.Append()
  — evaluates filter expressions
  — evaluates trigger rules; rising-edge fires TriggerSendFunc goroutine
  ↓
drawer.Push(at, val) → redraw
```

---

## Key Types

| Type | File | Role |
|---|---|---|
| `DiagramUIRefs` | `ui/port/diagramuirefs.go` | All UI components for one diagram. `ChartSlot` is the `Max` container wrapping the chart (allows placeholder swap). `Maximized bool` guards duplicate windows. |
| `ChartRegistry` | `ui/port/chart_registry.go` | Global map: `diagramID → *DiagramUIRefs` |
| `MultiSeriesRing` | `ui/data/multiseriesring.go` | Ring buffer: `times []time.Time`, `values [][]int32` for N series. Evaluates trigger rules on every Append. |
| `SingularSeriesData` | `ui/data/singularseriesdata.go` | `IMultiSeriesData` for BoolFill. Single `(lastValue, lastTime)`. |
| `TriggerSendFunc` | `ui/data/imultiseriesdata.go` | `func(ruleID string, files []string)` injected from WidgetService. Closures capture `jobID` and `diagramID`. |
| `TriggerRule` | `models/triggerrule.go` | Condition + action pair. `MaintainLink` resolves live from source FilterComponent. `Edited` disables re-linking. |
| `TriggerOperationMsg` | `models/triggeroperationmsg.go` | Sent to Evacron on trigger fire. Fields: `ID`, `JobID`, `RuleID`, `DiagramID`, `Files []string`, `Status`. |
| `ExecutionOutput` | `models/execution_output.go` | Streamed log line: `JobID`, `FilePath`, `Stream`, `Line`. Stream field is ignored in UI. |
| `ExecutionResult` | `models/execution_result.go` | Final result: `JobID`, `Success bool`, `Error string`. Discriminated from ExecutionOutput by presence of `"success"` JSON key. |
| `TriggerExecution` | `models/execution_log.go` | One full execution run: `TriggerID`, `FilePath`, `StartedAt`, `FinishedAt`, `Success *bool`, `Lines []ExecutionLogEntry`. |
| `ExecutionLogEntry` | `models/execution_log.go` | Single log line: `Stream`, `Line`, `ParsedMsg`, `Timestamp`. Always render `ParsedMsg` — never `Line` or `Stream`. |
| `RingBuffer[T]` | `models/ring_buffer.go` | Generic ring buffer. `Push`, `All` (ordered), `Latest`. |
| `ExecutionLogRegistry` | `services/execution_log_registry.go` | `map[DiagramID][RuleID] → RingBuffer[TriggerExecution]` capacity 10. |
| `TriggerExecutionCoordinator` | `services/trigger_execution_coordinator.go` | `OnStart` / `OnLine` / `OnDone` — wires stream events to registry. |
| `triggerStreamTracker` | `services/trigger_stream_tracker.go` | `map[msgID → triggerStreamEntry]`. Notifies coordinator at each lifecycle stage. |
| `maximizeAdapter` | `diagrams/fyne/maximize_adapter.go` | Thin BaseWidget wrapping `d.root`. Provides reliable canvas reference during maximize. |
| `AggregationTree` | `ui/data/aggregation_tree.go` | Nested map (year→month→week→day→hour→minute→second). Built once via `BuildTree([]DataPoint)`. Immutable after construction. Queried via `IterateLevel(TimePeriod) → *BucketIterator`. |
| `OperationsStateDifferentiator` | `models/operationsstatedifferentiator.go` | Snapshots rule list; tracks per-category changes; `HasChanges()` / `GetDiff()`. |

---

## Diagram Widget Architecture

Each chart type has two layers:
- **Widget** (Fyne BaseWidget) — owns hover, tooltip, zoom
- **Drawer** — owns canvas primitives, `d.root`, `d.raster`, performs all drawing

`CreateRenderer()` returns `[drawer.Root(), tooltip.Object()]`.

**Maximize flow:**
1. Remove widget from slot; show placeholder
2. `widget.MaximizeView(1200, 800)` pre-sizes `d.root`, creates `maximizeAdapter`, installs `drawer.refreshHook = adapter.Refresh`
3. Set adapter as new window content
4. On close: clear hook, restore widget to slot

**refreshHook:** nil in normal mode; set to `adapter.Refresh()` during maximize — routes through adapter's reliable canvas instead of stale `CanvasForObject(d.root)`.

**Snapshot:** normal mode crops main window canvas to widget bounds; maximize mode captures `maximizeWindows[diagramID].Canvas().Capture()` directly.

---

## Trigger Protocol
```
Phase 1 — dispatch (main WS, 10s timeout)
  → Send TriggerOperationMsg  {type:"job.trigger", id, jobId, ruleId, diagramId, files}
  ← Receive WSMessage         {data: {port: N}}

Phase 2 — stream (new WS to ws://<host>:<port>)
  → Send  "ACK"
  ← Receive  <ExecutionOutput JSON>  (log lines, repeated)
  ← Receive  <ExecutionResult JSON>  (final result — has "success" key)
  ← Receive  "STREAM-END"
  → Send  "ACK"
  → Close stream connection
```

**Message discrimination in `handleStreamLine`:**
- Check for `"success"` key in raw JSON via `map[string]json.RawMessage`
- Present → `ExecutionResult` → call `completeWithResult`
- Absent → `ExecutionOutput` → call `recordLine`
- `error == "exit status 1"` is suppressed — not added to log lines

---

## Execution Log System

**Backend** (all already implemented):
- `ExecutionLogRegistry` — `map[DiagramID][RuleID] → RingBuffer[TriggerExecution]` cap 10
- `TriggerExecutionCoordinator` — `OnStart` / `OnLine` / `OnDone`
- `TriggerExecutionBuilder` — accumulates lines; `Finish(success)` → `TriggerExecution`
- `triggerStreamTracker` — `register` → `recordLine` → `completeWithResult` → `remove`
- `drainStream` → `handleStreamLine` — discriminates message types, feeds coordinator

**UI** (`ui/execution_log/`):
- `ExecutionLogPanel` — below diagram grid via `container.NewVSplit` (offset 0.7)
- Two-row toolbar: Row1 = Maximize / Detach / Download JSON; Row2 = string filter (fills width) + operation type dropdown
- Tabs: diagram → rule (label as tab title) → execution list → execution detail
- Only diagrams with triggerRules where `len(actions) > 0` get a tab
- Execution list: `#index · HH:MM:SS · filepath.Base · ✓/✗/…` — green ✓, red ✗
- Execution detail: back button + header + scrollable monospace ParsedMsg lines
- Filters (AND): string filter scoped to detail view lines; operation type (`All/Operation/Label/Program`)
- `LogLineRenderer` — left colored bar (4px) + badge per `OperationType`; free-form lines get bar only
- Style config: `config/log_styles.json` — hex colors per type; falls back to hardcoded defaults
- Auto-refresh every 500ms via goroutine + ticker; manual refresh button
- Detach opens new OS window; Maximize toggles VSplit offset to 0.0
- Download: `dialog.ShowFileSave` → JSON-encodes selected `TriggerExecution`
- Never render `entry.Line` or `entry.Stream` — always `entry.ParsedMsg`

**Log line format (EVA structured lines):**
```
Line:<N> <Name> <Type> <Phase>
```
`OperationType`: `Operation`, `Label`, `Program`. Non-matching = free-form.

---

## Filter System
- `SetupItem` has a `Filter` with `FilterSetup` (list of `FilterComponent` expressions + `AND/OR` mode)
- Evaluated by `pkg/utils.RunExpression()` via `tafexpr`
- `MultiSeriesRing.ApplyFilterChanges()` backfills expressions over stored data points
- Results stored in `ConditionResults[i].Results[componentID]` — used by drawers for overlays

---

## Historic Aggregation (Bar Chart)
- `AggregationTree` built once via `BuildTree([]DataPoint)` — immutable after construction
- Queried via `IterateLevel(TimePeriod) → *BucketIterator`
- `AggregationFunction.Extract(Bucket, series)` — Sum / Avg / Min / Max / Count / PerX rates
- `HistoricDataStore` replaces `MultiSeriesRing` in historic mode — never mixed
- Iterators are not goroutine-safe — obtain a fresh one per render pass
- Bar chart toolbar adds Period and Function selectors in historic mode only

---

## Operations System
See `ui/fyne/operations/` for full implementation.

Entry point: `ShowOperationsModal(parent, components, initialRules, onSave)`

Two tabs:
- **Conditions** — add/remove/link `TriggerRule` from `FilterComponent` pool
- **Actions** — assign ordered `.eva` action files per rule; add/edit/remove/reorder

Save is gated by `OperationsStateDifferentiator.HasChanges()`. On save each rule calls `BreakLink()` to materialise linked fields before handing `[]TriggerRule` to `onSave`.

**TODO — remaining work:**
1. Add `Operations []TriggerRule` to diagram JSON model (`omitempty`)
2. Implement save-to-file in `onSave` callback
3. Add `TriggerRules []TriggerRule` to `DiagramUIRefs`; populate at startup

---

## Conventions
- All Fyne UI mutations must run on the main goroutine — use `fyne.Do(fn)`
- `BaseWidget.Resize()` must call `w.BaseWidget.Resize(size)` — omitting leaves size `{0,0}` and breaks hit-testing
- Use `container.RemoveAll()` + `container.Add()` — not direct `Objects` slice mutation
- `ChartSlot` is always `container.NewMax(chartWidget)` — never put raw widget into dashboard layout
- No method longer than 20 lines
- All structs use private fields with public constructor functions
- Widget construction in constructors only — `Refresh()` mutates `.Objects` and calls `.Refresh()` on existing containers, never constructs new widgets
- `filepath.Base()` everywhere a file path is displayed in UI
- `LogLineRenderer` — free-form lines get left bar only; structured lines get bar + badge
- `config/log_styles.json` is the single source of truth for log line colors — never hardcode colors in renderer

---

## Build & Run
```bash
go build ./...
make build          # produces build/evamon.exe
evamon view <view-name>
evamon view --dashboard <dashboard-name>
# Config: ~/.evamon/application.properties
# Lib:    ~/.evamon/lib/
```

**Key dependencies:** `fyne.io/fyne/v2` v2.7.2, `github.com/coder/websocket`,
`github.com/gclkaze/tafexpr`, `github.com/spf13/cobra`