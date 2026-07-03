package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDayGetName(t *testing.T) {
	day := Day{
		DayNameDa: "Mandag",
		DayNameEn: "Monday",
	}

	if name := day.GetName("da"); name != "Mandag" {
		t.Errorf("Expected Mandag, got %s", name)
	}

	if name := day.GetName("en"); name != "Monday" {
		t.Errorf("Expected Monday, got %s", name)
	}

	// Default fallback to Danish
	if name := day.GetName("fr"); name != "Mandag" {
		t.Errorf("Expected Mandag, got %s", name)
	}
}

func TestWeekPlanDaysOrder(t *testing.T) {
	wp := WeekPlan{
		Monday:    Day{DayNameDa: "Mandag", DayNameEn: "Monday"},
		Tuesday:   Day{DayNameDa: "Tirsdag", DayNameEn: "Tuesday"},
		Wednesday: Day{DayNameDa: "Onsdag", DayNameEn: "Wednesday"},
		Thursday:  Day{DayNameDa: "Torsdag", DayNameEn: "Thursday"},
		Friday:    Day{DayNameDa: "Fredag", DayNameEn: "Friday"},
		Saturday:  Day{DayNameDa: "Lørdag", DayNameEn: "Saturday"},
		Sunday:    Day{DayNameDa: "Søndag", DayNameEn: "Sunday"},
	}

	days := wp.Days()
	if len(days) != 7 {
		t.Errorf("Expected 7 days, got %d", len(days))
	}

	expectedOrder := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for i, d := range days {
		if d.Key != expectedOrder[i] {
			t.Errorf("Expected day at index %d to be %s, got %s", i, expectedOrder[i], d.Key)
		}
	}
}

func TestTemplates(t *testing.T) {
	funcMap := template.FuncMap{
		"getDayName": func(day Day, lang string) string {
			return day.GetName(lang)
		},
		"getDayIndex": func(key string) int {
			return getDayIndex(key)
		},
		"getDayDate": func(key string, lang string) string {
			return "Jun 29"
		},
	}

	testPlan := WeekPlan{
		Monday:    Day{DayNameDa: "Mandag", DayNameEn: "Monday", Tasks: []Task{{ID: "1", Title: "Breakfast"}}},
		Tuesday:   Day{DayNameDa: "Tirsdag", DayNameEn: "Tuesday", Tasks: []Task{}},
		Wednesday: Day{DayNameDa: "Onsdag", DayNameEn: "Wednesday", Tasks: []Task{}},
		Thursday:  Day{DayNameDa: "Torsdag", DayNameEn: "Thursday", Tasks: []Task{}},
		Friday:    Day{DayNameDa: "Fredag", DayNameEn: "Friday", Tasks: []Task{}},
		Saturday:  Day{DayNameDa: "Lørdag", DayNameEn: "Saturday", Tasks: []Task{}},
		Sunday:    Day{DayNameDa: "Søndag", DayNameEn: "Sunday", Tasks: []Task{}},
	}

	data := TemplateData{
		Language: "da",
		Trans:    translations["da"],
		State: AppState{
			Settings: Settings{Language: "da"},
			WeekPlan: testPlan,
		},
	}

	// Test dashboard.html
	tmplDashboard, err := template.New("dashboard.html").Funcs(funcMap).ParseFiles("templates/dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	var buf bytes.Buffer
	if err := tmplDashboard.Execute(&buf, data); err != nil {
		t.Errorf("Failed to execute dashboard.html: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Tirsdag") {
		t.Errorf("Expected dashboard HTML to contain 'Tirsdag'")
	}
	if !strings.Contains(out, "Søndag") {
		t.Errorf("Expected dashboard HTML to contain 'Søndag'")
	}

	// Test settings.html
	tmplSettings, err := template.New("settings.html").Funcs(funcMap).ParseFiles("templates/settings.html")
	if err != nil {
		t.Fatalf("Failed to parse settings.html: %v", err)
	}

	buf.Reset()
	if err := tmplSettings.Execute(&buf, data); err != nil {
		t.Errorf("Failed to execute settings.html: %v", err)
	}

	outSettings := buf.String()
	if !strings.Contains(outSettings, "Indstillinger") {
		t.Errorf("Expected settings HTML to contain 'Indstillinger'")
	}
}

func TestMobileTouchSettings(t *testing.T) {
	// Verify that TouchFriendlyMode and RowTapToggle exist on Settings struct
	s := Settings{
		TouchFriendlyMode: true,
		RowTapToggle:      true,
	}
	if !s.TouchFriendlyMode || !s.RowTapToggle {
		t.Error("Settings struct fields TouchFriendlyMode and RowTapToggle not functioning as expected")
	}

	// Verify settings.html can render them
	funcMap := template.FuncMap{
		"getDayName": func(day Day, lang string) string {
			return day.GetName(lang)
		},
		"getDayIndex": func(key string) int {
			return 0
		},
		"getDayDate": func(key string, lang string) string {
			return "Jun 29"
		},
	}

	data := TemplateData{
		Language: "da",
		Trans:    translations["da"],
		State: AppState{
			Settings: Settings{
				Language:          "da",
				TouchFriendlyMode: true,
				RowTapToggle:      true,
			},
		},
	}

	tmplSettings, err := template.New("settings.html").Funcs(funcMap).ParseFiles("templates/settings.html")
	if err != nil {
		t.Fatalf("Failed to parse settings.html: %v", err)
	}

	var buf bytes.Buffer
	if err := tmplSettings.Execute(&buf, data); err != nil {
		t.Errorf("Failed to execute settings.html: %v", err)
	}

	html := buf.String()
	// Check if our touch settings checkboxes are rendered
	if !strings.Contains(html, "settings-touch-friendly-mode") {
		t.Error("Expected settings HTML to render touch-friendly-mode checkbox")
	}
	if !strings.Contains(html, "settings-row-tap-toggle") {
		t.Error("Expected settings HTML to render row-tap-toggle checkbox")
	}
}

func TestWeekTransition(t *testing.T) {
	// Initialize a test state
	stateMu.Lock()
	currentState = AppState{
		Settings: Settings{
			Language:      "da",
			AutoResetWeek: true,
		},
		WeekPlan: WeekPlan{
			Monday: Day{
				DayNameDa: "Mandag",
				DayNameEn: "Monday",
				Tasks: []Task{
					{ID: "current_adhoc", Title: "Current Task", Done: true},
				},
			},
		},
		TemplatePlan: WeekPlan{
			Monday: Day{
				DayNameDa: "Mandag",
				DayNameEn: "Monday",
				Tasks: []Task{
					{ID: "template_task", Title: "Template Task", Done: false},
				},
			},
		},
		LastWeekNum: 10, // Mock last week
	}
	stateMu.Unlock()

	// 1. If LastWeekNum is 0 (first load baseline), it shouldn't reset, but baseline LastWeekNum to current week.
	stateMu.Lock()
	currentState.LastWeekNum = 0
	stateMu.Unlock()

	_, currentWeek := time.Now().ISOWeek()
	checkWeekTransition()

	stateMu.RLock()
	if currentState.LastWeekNum != currentWeek {
		t.Errorf("Expected LastWeekNum to be updated to %d, got %d", currentWeek, currentState.LastWeekNum)
	}
	if len(currentState.WeekPlan.Monday.Tasks) != 1 || currentState.WeekPlan.Monday.Tasks[0].ID != "current_adhoc" {
		t.Error("Expected no reset when LastWeekNum is initialized from 0")
	}
	stateMu.RUnlock()

	// 2. Transition when AutoResetWeek is enabled
	stateMu.Lock()
	currentState.LastWeekNum = currentWeek - 1
	if currentState.LastWeekNum <= 0 {
		currentState.LastWeekNum = 52
	}
	currentState.Settings.AutoResetWeek = true
	// Ensure we have our current_adhoc back
	currentState.WeekPlan.Monday.Tasks = []Task{{ID: "current_adhoc", Title: "Current Task", Done: true}}
	stateMu.Unlock()

	checkWeekTransition()

	stateMu.RLock()
	if currentState.LastWeekNum != currentWeek {
		t.Errorf("Expected LastWeekNum to update to %d, got %d", currentWeek, currentState.LastWeekNum)
	}
	if len(currentState.WeekPlan.Monday.Tasks) != 1 || currentState.WeekPlan.Monday.Tasks[0].Title != "Template Task" {
		t.Error("Expected WeekPlan to automatically reset to TemplatePlan when week changes and AutoResetWeek is true")
	}
	stateMu.RUnlock()

	// 3. Transition when AutoResetWeek is disabled
	stateMu.Lock()
	currentState.LastWeekNum = currentWeek - 1
	if currentState.LastWeekNum <= 0 {
		currentState.LastWeekNum = 52
	}
	currentState.Settings.AutoResetWeek = false
	// Set active tasks to adhoc
	currentState.WeekPlan.Monday.Tasks = []Task{{ID: "current_adhoc2", Title: "Current Task 2", Done: true}}
	stateMu.Unlock()

	checkWeekTransition()

	stateMu.RLock()
	if currentState.LastWeekNum != currentWeek {
		t.Errorf("Expected LastWeekNum to update to %d, got %d", currentWeek, currentState.LastWeekNum)
	}
	if len(currentState.WeekPlan.Monday.Tasks) != 1 || currentState.WeekPlan.Monday.Tasks[0].ID != "current_adhoc2" {
		t.Error("Expected WeekPlan to NOT reset when AutoResetWeek is false")
	}
	stateMu.RUnlock()
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a dummy handler
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("Short, stout"))
	})

	// Wrap with our logging middleware
	loggedHandler := loggingMiddleware(dummy)

	// Create test request
	req := httptest.NewRequest("GET", "/test-log-endpoint", nil)
	rec := httptest.NewRecorder()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // restore original

	// Serve request
	loggedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("Expected status %d, got %d", http.StatusTeapot, rec.Code)
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "GET") {
		t.Error("Expected log to contain HTTP method GET")
	}
	if !strings.Contains(logStr, "/test-log-endpoint") {
		t.Error("Expected log to contain RequestURI /test-log-endpoint")
	}
	if !strings.Contains(logStr, "418 I'm a teapot") {
		t.Error("Expected log to contain status code and text '418 I'm a teapot'")
	}
	if !strings.Contains(logStr, "ms)") {
		t.Error("Expected log to contain duration formatted strictly in milliseconds with 'ms)' suffix")
	}
}

func TestLoadEmptyPlan(t *testing.T) {
	// Create an empty test file (representing touch plan.json)
	testPath := "test_empty_plan.json"
	f, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	f.Close()
	defer os.Remove(testPath)

	// Set planPath to empty test file
	oldPath := planPath
	planPath = testPath
	defer func() {
		planPath = oldPath
	}()

	// Execute loadPlan
	err = loadPlan()
	if err != nil {
		t.Fatalf("Expected loadPlan to succeed on empty 0-byte file, got %v", err)
	}

	stateMu.RLock()
	if currentState.Settings.Language != "da" {
		t.Errorf("Expected default language 'da', got %s", currentState.Settings.Language)
	}
	if currentState.WeekPlan.Monday.DayNameDa != "Mandag" {
		t.Errorf("Expected day name 'Mandag', got %s", currentState.WeekPlan.Monday.DayNameDa)
	}
	stateMu.RUnlock()
}
