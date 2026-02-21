package viewproject_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gclkaze/evamon/cmd/internal/viewproject"
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

func mustContainAny(t *testing.T, err error, substrs ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing one of %v, got nil", substrs)
	}
	for _, s := range substrs {
		if strings.Contains(err.Error(), s) {
			return
		}
	}
	t.Fatalf("expected error containing one of %v, got: %v", substrs, err)
}

// --- tests: LoadViewProject (jobID + view) ---

func TestLoadViewProject_OK_BooleanDiagram(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "configs", "soilcrates.project.json")
	projectJSON := `{
	  "jobID": "this-is-an-id",
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
	writeFile(t, projectPath, projectJSON)

	vp, err := viewproject.LoadViewProject(projectPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if vp.ProjectPath == "" {
		t.Fatalf("expected ProjectPath to be set")
	}
	if vp.JobID != "this-is-an-id" {
		t.Fatalf("expected JobID=%q, got %q", "this-is-an-id", vp.JobID)
	}
	if vp.View.MultiTab != false {
		t.Fatalf("expected MultiTab=false")
	}
	if len(vp.View.Diagrams) != 1 {
		t.Fatalf("expected 1 diagram, got %d", len(vp.View.Diagrams))
	}

	// typed union should decode into BooleanStyle
	setup := vp.View.Diagrams[0].Setup[0]
	if setup.DiagramStyle == nil {
		t.Fatalf("expected DiagramStyle not nil")
	}
	if _, ok := setup.DiagramStyle.(viewproject.BooleanStyle); !ok {
		t.Fatalf("expected BooleanStyle, got %T", setup.DiagramStyle)
	}
}

func TestLoadViewProject_OK_BarDiagram_WindowStyleInts(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
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
	writeFile(t, projectPath, projectJSON)

	vp, err := viewproject.LoadViewProject(projectPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if vp.View.WindowStyle == nil || vp.View.WindowStyle.Width == nil || vp.View.WindowStyle.Height == nil {
		t.Fatalf("expected view.windowStyle with width/height set")
	}
	if *vp.View.WindowStyle.Width != 900 || *vp.View.WindowStyle.Height != 600 {
		t.Fatalf("unexpected view.windowStyle values: %+v", vp.View.WindowStyle)
	}

	setup := vp.View.Diagrams[0].Setup[0]
	if _, ok := setup.DiagramStyle.(viewproject.BarStyle); !ok {
		t.Fatalf("expected BarStyle, got %T", setup.DiagramStyle)
	}
}

func TestLoadViewProject_Fails_MissingJobID(t *testing.T) {
	tmp := t.TempDir()
	projectPath := filepath.Join(tmp, "project.json")

	writeFile(t, projectPath, `{
	  "view": { "multiTab": false, "diagrams": [] }
	}`)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "jobID is required")
}

func TestLoadViewProject_Fails_UnknownField_StrictJSON(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	// "multiTabs" is a typo: strict decoder should fail
	writeFile(t, projectPath, `{
	  "jobID": "job-123",
	  "view": {
	    "multiTabs": false,
	    "diagrams": [
	      { "type": "boolean", "setup": [
	        { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	      ]}
	    ]
	  }
	}`)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "unknown field")
}

func TestLoadViewProject_Fails_WindowStyleNegative(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	writeFile(t, projectPath, `{
	  "jobID": "job-123",
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

	_, err := viewproject.LoadViewProject(projectPath)
	// note: error path comes from view validation; keep it fuzzy enough
	mustContain(t, err, "windowStyle.width must be > 0")
}

func TestLoadViewProject_Fails_BooleanDiagramStyleMissingKeys(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	// boolean diagramStyle missing "false"
	writeFile(t, projectPath, `{
	  "jobID": "job-123",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"green"} }
	    ]}
	  ]}
	}`)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "requires non-empty keys")
}

func TestLoadViewProject_Fails_TypeUnionMismatch_BooleanDiagramWithBarStyleShape(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	// diagram type boolean, but style object looks like bar style (axis/background)
	writeFile(t, projectPath, `{
	  "jobID": "job-123",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"axis":"green","background":"black"} }
	    ]}
	  ]}
	}`)

	_, err := viewproject.LoadViewProject(projectPath)
	// Will fail either during decode (unknown fields for BooleanStyle) or validation
	mustContain(t, err, "diagramStyle")
}

func TestLoadViewProject_Fails_VariableTypeMismatch_BarRequiresInteger(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	writeFile(t, projectPath, `{
	  "jobID": "job-123",
	  "view": { "multiTab": false, "diagrams": [
	    { "type": "bar", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"axis":"green","background":"black"} }
	    ]}
	  ]}
	}`)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "variableType must be")
}

// --- tests: LoadViewWindow (view-only) ---

func TestLoadViewWindow_OK_BooleanDiagram(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	viewJSON := `{
	  "multiTab": false,
	  "diagrams": [
	    {
	      "type": "boolean",
	      "setup": [
	        {
	          "variable": "soilcrates-up",
	          "title": "Soilcrates is UP?",
	          "variableType": "boolean",
	          "diagramStyle": { "true": "green", "false": "red" },
	          "windowStyle": { "width": 300, "height": 300 }
	        }
	      ]
	    }
	  ]
	}`
	writeFile(t, viewPath, viewJSON)

	vw, err := viewproject.LoadViewWindow(viewPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if vw.MultiTab != false {
		t.Fatalf("expected MultiTab=false")
	}
	if len(vw.Diagrams) != 1 {
		t.Fatalf("expected 1 diagram, got %d", len(vw.Diagrams))
	}

	setup := vw.Diagrams[0].Setup[0]
	if setup.WindowStyle == nil || setup.WindowStyle.Width == nil || setup.WindowStyle.Height == nil {
		t.Fatalf("expected setup.windowStyle with width/height set")
	}
	if *setup.WindowStyle.Width != 300 || *setup.WindowStyle.Height != 300 {
		t.Fatalf("unexpected setup.windowStyle values: %+v", setup.WindowStyle)
	}
	if _, ok := setup.DiagramStyle.(viewproject.BooleanStyle); !ok {
		t.Fatalf("expected BooleanStyle, got %T", setup.DiagramStyle)
	}
}

func TestLoadViewWindow_Fails_UnknownField_StrictJSON(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	// unknown top-level field "jobID" should fail for view-only document
	writeFile(t, viewPath, `{
	  "jobID": "job-123",
	  "multiTab": false,
	  "diagrams": [
	    { "type": "boolean", "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	    ]}
	  ]
	}`)

	_, err := viewproject.LoadViewWindow(viewPath)
	mustContain(t, err, "unknown field")
}

func TestLoadViewWindow_Fails_EmptyDiagrams(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	writeFile(t, viewPath, `{
	  "multiTab": false,
	  "diagrams": []
	}`)

	_, err := viewproject.LoadViewWindow(viewPath)
	// view-only validation error
	mustContain(t, err, "diagrams must not be empty")
}

func TestLoadViewWindow_Fails_WindowStyleNegative(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	writeFile(t, viewPath, `{
	  "multiTab": true,
	  "windowStyle": { "width": -10, "height": 100 },
	  "diagrams": [
	    { "type": "bar", "setup": [
	      { "variable": "u", "title": "Users", "variableType": "integer", "diagramStyle": {"axis":"g","background":"b"} }
	    ]}
	  ]
	}`)

	_, err := viewproject.LoadViewWindow(viewPath)
	mustContain(t, err, "windowStyle.width must be > 0")
}

func TestLoadViewWindow_Fails_MissingDiagramType(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	writeFile(t, viewPath, `{
	  "multiTab": false,
	  "diagrams": [
	    { "setup": [
	      { "variable": "x", "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	    ]}
	  ]
	}`)

	_, err := viewproject.LoadViewWindow(viewPath)
	mustContain(t, err, "type is required")
}

func TestLoadViewWindow_Fails_MissingSetupVariable(t *testing.T) {
	tmp := t.TempDir()

	viewPath := filepath.Join(tmp, "view.json")
	writeFile(t, viewPath, `{
	  "multiTab": false,
	  "diagrams": [
	    { "type": "boolean", "setup": [
	      { "title": "X", "variableType": "boolean", "diagramStyle": {"true":"g","false":"r"} }
	    ]}
	  ]
	}`)

	_, err := viewproject.LoadViewWindow(viewPath)
	mustContain(t, err, "variable is required")
}
func TestLoadViewProject_OK_BarDiagram_MultiVariableSetup(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "My super important Value!",
	            "variableType": "integer",
	            "diagramStyle": { "background": "black" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU Metrics",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" }
	              },
	              {
	                "variable": "mem",
	                "title": "Memory Metrics",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	vp, err := viewproject.LoadViewProject(projectPath)
	if err != nil {
		t.Fatalf("expected ok, got err: %v", err)
	}

	if len(vp.View.Diagrams) != 1 {
		t.Fatalf("expected 1 diagram, got %d", len(vp.View.Diagrams))
	}
	if len(vp.View.Diagrams[0].Setup) != 1 {
		t.Fatalf("expected 1 setup item, got %d", len(vp.View.Diagrams[0].Setup))
	}

	setup := vp.View.Diagrams[0].Setup[0]
	if got := len(setup.MultiVariableSetup); got != 2 {
		t.Fatalf("expected 2 children, got %d", got)
	}

	// Parent style should decode into BarStyle
	if setup.DiagramStyle == nil {
		t.Fatalf("expected parent DiagramStyle not nil")
	}
	if _, ok := setup.DiagramStyle.(viewproject.BarStyle); !ok {
		t.Fatalf("expected parent BarStyle, got %T", setup.DiagramStyle)
	}

	// Child styles should decode into BarStyle as well
	for i, child := range setup.MultiVariableSetup {
		if child.DiagramStyle == nil {
			t.Fatalf("expected child[%d] DiagramStyle not nil", i)
		}
		if _, ok := child.DiagramStyle.(viewproject.BarStyle); !ok {
			t.Fatalf("expected child[%d] BarStyle, got %T", i, child.DiagramStyle)
		}
	}
}

func TestLoadViewProject_Fails_MultiVariableSetup_TooShort(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "integer",
	            "diagramStyle": { "background": "black" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	// Keep assertion fuzzy because your error message might be slightly different.
	mustContain(t, err, "multiVariableSetup")
	mustContain(t, err, "at least 2")
}

func TestLoadViewProject_Fails_MultiVariableSetup_ParentBackgroundMissing(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "integer",
	            "diagramStyle": { "axis": "green" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" }
	              },
	              {
	                "variable": "mem",
	                "title": "MEM",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "diagramStyle")
	mustContain(t, err, "background")
}

func TestLoadViewProject_Fails_MultiVariableSetup_ChildBarMissing(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "integer",
	            "diagramStyle": { "background": "black" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU",
	                "variableType": "integer",
	                "diagramStyle": { }
	              },
	              {
	                "variable": "mem",
	                "title": "MEM",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "multiVariableSetup")
	mustContain(t, err, "diagramStyle")
	mustContain(t, err, "bar")
}

func TestLoadViewProject_Fails_MultiVariableSetup_DuplicateChildVariable(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "integer",
	            "diagramStyle": { "background": "black" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU #1",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" }
	              },
	              {
	                "variable": "cpu",
	                "title": "CPU #2",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "duplic")
}

func TestLoadViewProject_Fails_MultiVariableSetup_NotAllowedForBooleanDiagram(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "boolean",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "boolean",
	            "diagramStyle": { "true": "green", "false": "red" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" }
	              },
	              {
	                "variable": "mem",
	                "title": "MEM",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "multiVariableSetup")
	mustContainAny(t, err, "bar/line", "unknown field", "diagramStyle")
}

func TestLoadViewProject_Fails_MultiVariableSetup_UnknownField_StrictJSON(t *testing.T) {
	tmp := t.TempDir()

	projectPath := filepath.Join(tmp, "project.json")
	projectJSON := `{
	  "jobID": "job-123",
	  "view": {
	    "multiTab": false,
	    "diagrams": [
	      {
	        "type": "bar",
	        "setup": [
	          {
	            "variable": "theValue",
	            "title": "Bundle",
	            "variableType": "integer",
	            "diagramStyle": { "background": "black" },
	            "multiVariableSetup": [
	              {
	                "variable": "cpu",
	                "title": "CPU",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "dodgerblue" },
	                "unknownField": 123
	              },
	              {
	                "variable": "mem",
	                "title": "MEM",
	                "variableType": "integer",
	                "diagramStyle": { "bar": "violet" }
	              }
	            ]
	          }
	        ]
	      }
	    ]
	  }
	}`
	writeFile(t, projectPath, projectJSON)

	_, err := viewproject.LoadViewProject(projectPath)
	mustContain(t, err, "unknown field")
}
