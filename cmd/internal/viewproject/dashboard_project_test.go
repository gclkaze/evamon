// dashboard_project_test.go
package viewproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempJSON(t *testing.T, dir, name, json string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(json), 0o600); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	return p
}

func TestLoadDashboardProject_OK_WidgetColumns(t *testing.T) {
	dir := t.TempDir()

	// This matches the "dashboard JSON" shape: column has jobId + widget:{type,setup}
	json := `{
	  "ID": "dash-001",
	  "title": "Evamon Monitoring Dashboard",
	  "views": [
		{
		  "id": "view-overview",
		  "title": "Overview",
		  "rows": [
			{
			  "columns": [
				{
				  "jobId": "job-soilcrates-health",
				  "widget": {
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
				},
				{
				  "jobId": "job-system-metrics",
				  "widget": {
					"type": "bar",
					"setup": [
					  {
						"variable": "theValue",
						"title": "My super important Value!",
						"variableType": "integer",
						"diagramStyle": { "axis": "dodgerblue", "background": "black" }
					  }
					]
				  }
				}
			  ]
			}
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "dashboard.json", json)

	dp, err := LoadDashboardProject(path)
	if err != nil {
		t.Fatalf("LoadDashboardProject: %v", err)
	}

	if dp.ID != "dash-001" {
		t.Fatalf("expected ID dash-001, got %q", dp.ID)
	}
	if dp.Title != "Evamon Monitoring Dashboard" {
		t.Fatalf("expected title, got %q", dp.Title)
	}
	if dp.ProjectPath == "" || !strings.HasSuffix(dp.ProjectPath, "dashboard.json") {
		t.Fatalf("expected ProjectPath to be filled, got %q", dp.ProjectPath)
	}

	if len(dp.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(dp.Views))
	}
	v := dp.Views[0]
	if v.Title != "Overview" {
		t.Fatalf("expected view title Overview, got %q", v.Title)
	}
	if len(v.Rows) != 1 || len(v.Rows[0].Columns) != 2 {
		t.Fatalf("expected 1 row with 2 cols, got rows=%d cols=%d", len(v.Rows), len(v.Rows[0].Columns))
	}

	// Verify widget->ViewWindow normalization: View.Diagrams should contain exactly the widget diagram.
	c0 := v.Rows[0].Columns[0]
	if c0.JobID != "job-soilcrates-health" {
		t.Fatalf("expected jobId, got %q", c0.JobID)
	}
	if c0.View.MultiTab != false {
		t.Fatalf("expected MultiTab=false, got %v", c0.View.MultiTab)
	}
	if len(c0.View.Diagrams) != 1 {
		t.Fatalf("expected 1 diagram in normalized ViewWindow, got %d", len(c0.View.Diagrams))
	}
	if c0.View.Diagrams[0].Type != DiagramTypeBoolean {
		t.Fatalf("expected boolean diagram, got %q", c0.View.Diagrams[0].Type)
	}

	c1 := v.Rows[0].Columns[1]
	if len(c1.View.Diagrams) != 1 || c1.View.Diagrams[0].Type != DiagramTypeBar {
		t.Fatalf("expected 1 bar diagram, got diagrams=%d type=%q", len(c1.View.Diagrams), c1.View.Diagrams[0].Type)
	}
}

func TestLoadDashboardProject_OK_ViewColumns(t *testing.T) {
	dir := t.TempDir()

	// Column can alternatively provide "view": ViewWindow (reuses your existing JSON format)
	json := `{
	  "ID": "dash-002",
	  "title": "Dashboard via view windows",
	  "views": [
		{
		  "title": "Tab A",
		  "rows": [
			{
			  "columns": [
				{
				  "jobId": "job-1",
				  "view": {
					"multiTab": false,
					"diagrams": [
					  {
						"type": "boolean",
						"setup": [
						  {
							"variable": "ok",
							"title": "OK?",
							"variableType": "boolean",
							"diagramStyle": { "true": "green", "false": "red" }
						  }
						]
					  }
					]
				  }
				}
			  ]
			}
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "dashboard_view.json", json)

	dp, err := LoadDashboardProject(path)
	if err != nil {
		t.Fatalf("LoadDashboardProject: %v", err)
	}
	if dp.ID != "dash-002" || dp.Title != "Dashboard via view windows" {
		t.Fatalf("unexpected dp fields: %+v", dp)
	}

	col := dp.Views[0].Rows[0].Columns[0]
	if col.JobID != "job-1" {
		t.Fatalf("expected job-1, got %q", col.JobID)
	}
	if len(col.View.Diagrams) != 1 || col.View.Diagrams[0].Type != DiagramTypeBoolean {
		t.Fatalf("expected boolean diagram in view, got %v", col.View.Diagrams)
	}
}

func TestLoadDashboardProject_StrictUnknownField_Fails(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-003",
	  "title": "Strict",
	  "views": [],
	  "unknownField": 123
	}`

	path := writeTempJSON(t, dir, "strict.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected error for unknown field, got nil")
	}
}

func TestLoadDashboardProject_ValidationFails_MissingTitle(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-004",
	  "views": []
	}`

	path := writeTempJSON(t, dir, "missing_title.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

func TestLoadDashboardProject_ValidationFails_ViewsEmpty(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-005",
	  "title": "No views",
	  "views": []
	}`

	path := writeTempJSON(t, dir, "views_empty.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

func TestLoadDashboardProject_ValidationFails_RowColumnsEmpty(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-006",
	  "title": "Bad grid",
	  "views": [
		{
		  "title": "Tab",
		  "rows": [
			{ "columns": [] }
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "bad_grid.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

func TestLoadDashboardProject_ColumnUnmarshalFails_MissingJobId(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-007",
	  "title": "Missing jobId",
	  "views": [
		{
		  "title": "Tab",
		  "rows": [
			{
			  "columns": [
				{
				  "widget": {
					"type": "bar",
					"setup": [
					  { "variable": "x", "title": "X", "variableType": "integer" }
					]
				  }
				}
			  ]
			}
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "missing_jobid.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoadDashboardProject_ColumnUnmarshalFails_BothWidgetAndView(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-008",
	  "title": "Both widget and view",
	  "views": [
		{
		  "title": "Tab",
		  "rows": [
			{
			  "columns": [
				{
				  "jobId": "job-x",
				  "widget": {
					"type": "bar",
					"setup": [
					  { "variable": "x", "title": "X", "variableType": "integer" }
					]
				  },
				  "view": {
					"multiTab": false,
					"diagrams": [
					  {
						"type": "boolean",
						"setup": [
						  { "variable": "ok", "title": "OK", "variableType": "boolean",
							"diagramStyle": { "true":"green", "false":"red" } }
						]
					  }
					]
				  }
				}
			  ]
			}
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "both.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLoadDashboardProject_ColumnUnmarshalFails_NeitherWidgetNorView(t *testing.T) {
	dir := t.TempDir()

	json := `{
	  "ID": "dash-009",
	  "title": "Neither widget nor view",
	  "views": [
		{
		  "title": "Tab",
		  "rows": [
			{
			  "columns": [
				{ "jobId": "job-x" }
			  ]
			}
		  ]
		}
	  ]
	}`

	path := writeTempJSON(t, dir, "neither.json", json)
	_, err := LoadDashboardProject(path)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
