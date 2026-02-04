package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gclkaze/evamon/cmd/internal/config"
)

// --- helpers ---

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustContain(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error containing %q, got: %v", substr, err)
	}
}

// --- tests ---

func TestLoad_OK_BooleanDiagram_RelativeScriptResolved(t *testing.T) {
	tmp := t.TempDir()

	// script file must exist for validation
	scriptPath := filepath.Join(tmp, "scripts", "check.eva")
	writeFile(t, scriptPath, "@Main(){}")

	// config references script relatively to config file location
	cfgPath := filepath.Join(tmp, "configs", "soilcrates.json")
	cfgJSON := `{
	  "script": "../scripts/check.eva",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "boolean",
	        "setup": [
	          {
	            "variable": "soilcrates-up",
	            "title": "Soilcrates is UP?",
	            "variableType": "boolean",
	            "diagramStyle": { "true": "green", "false": "red" }
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, cfgPath, cfgJSON)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if cfg.ConfigPath == "" {
		t.Fatalf("expected ConfigPath to be set")
	}
	if cfg.ScriptAbs != scriptPath {
		t.Fatalf("expected ScriptAbs=%q, got %q", scriptPath, cfg.ScriptAbs)
	}
	if cfg.View.MultiTab != false {
		t.Fatalf("expected MultiTab=false")
	}
	if len(cfg.View.Diagrams) != 1 {
		t.Fatalf("expected 1 diagram, got %d", len(cfg.View.Diagrams))
	}

	// typed union should decode into BooleanStyle
	setup := cfg.View.Diagrams[0].Setup[0]
	if setup.DiagramStyle == nil {
		t.Fatalf("expected DiagramStyle not nil")
	}
	if _, ok := setup.DiagramStyle.(config.BooleanStyle); !ok {
		t.Fatalf("expected BooleanStyle, got %T", setup.DiagramStyle)
	}
}

func TestLoad_OK_BarDiagram_WindowStyleInts(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	cfgJSON := `{
	  "script": "run.eva",
	  "view": {
	    "multiTab": true,
	    "windowStyle": { "width": 900, "height": 600 },
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "soilcrates-users",
	            "title": "Soilcrates Users",
	            "variableType": "integer",
	            "diagramStyle": { "axis": "green", "background": "black" }
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, cfgPath, cfgJSON)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if cfg.View.WindowStyle == nil || cfg.View.WindowStyle.Width == nil || cfg.View.WindowStyle.Height == nil {
		t.Fatalf("expected view.windowStyle with width/height set")
	}
	if *cfg.View.WindowStyle.Width != 900 || *cfg.View.WindowStyle.Height != 600 {
		t.Fatalf("unexpected view.windowStyle values: %+v", cfg.View.WindowStyle)
	}

	setup := cfg.View.Diagrams[0].Setup[0]
	if _, ok := setup.DiagramStyle.(config.BarStyle); !ok {
		t.Fatalf("expected BarStyle, got %T", setup.DiagramStyle)
	}
}

func TestLoad_Fails_MissingScript(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.json")

	writeFile(t, cfgPath, `{
	  "view": { "multiTab": false, "diagrams": [] }
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "script is required")
}

func TestLoad_Fails_ScriptDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.json")

	writeFile(t, cfgPath, `{
	  "script": "nope.eva",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	    ]}
	  ]}
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "script not found")
}

func TestLoad_Fails_ScriptIsDirectory(t *testing.T) {
	tmp := t.TempDir()

	// make a directory that will be used as "script"
	scriptDir := filepath.Join(tmp, "asdir.eva")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(tmp, "config.json")
	writeFile(t, cfgPath, `{
	  "script": "asdir.eva",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	    ]}
	  ]}
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "script path is a directory")
}

func TestLoad_Fails_UnknownField_StrictJSON(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	// "multiTabs" is a typo: strict decoder should fail
	writeFile(t, cfgPath, `{
	  "script": "run.eva",
	  "view": {
	    "multiTabs": false,
	    "diagrams": [
	      { "type": "boolean", "setup": [
	        { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	      ]}
	    ]
	  }
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "unknown field")
}

func TestLoad_Fails_WindowStyleNegative(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	writeFile(t, cfgPath, `{
	  "script": "run.eva",
	  "view": {
	    "multiTab": true,
	    "windowStyle": { "width": -1, "height": 600 },
	    "diagrams": [
	      { "type": "bar", "setup": [
	        { "variable": "u", "title": "Users", "variableType": "integer", "diagramStyle": {"axis":"g","background":"b"} }
	      ]}
	    ]
	  }
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "view.windowStyle.width must be > 0")
}

func TestLoad_Fails_BooleanDiagramStyleMissingKeys(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	// boolean diagramStyle missing "false"
	writeFile(t, cfgPath, `{
	  "script": "run.eva",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"green"} }
	    ]}
	  ]}
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "requires non-empty keys")
}

func TestLoad_Fails_TypeUnionMismatch_BooleanDiagramWithBarStyleShape(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	// diagram type boolean, but style object looks like bar style (axis/background)
	writeFile(t, cfgPath, `{
	  "script": "run.eva",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"axis":"green","background":"black"} }
	    ]}
	  ]}
	}`)

	_, err := config.Load(cfgPath)
	// Will fail either during decode (unknown fields for BooleanStyle) or validation
	mustContain(t, err, "diagramStyle")
}

func TestLoad_Fails_VariableTypeMismatch_BarRequiresInteger(t *testing.T) {
	tmp := t.TempDir()

	scriptPath := filepath.Join(tmp, "run.eva")
	writeFile(t, scriptPath, "@Main(){}")

	cfgPath := filepath.Join(tmp, "config.json")
	writeFile(t, cfgPath, `{
	  "script": "run.eva",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "bar", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"axis":"green","background":"black"} }
	    ]}
	  ]}
	}`)

	_, err := config.Load(cfgPath)
	mustContain(t, err, "variableType must be")
}
