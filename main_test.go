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
		"getDayDate": func(key string, lang string, offsetDays int) string {
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
		Weeks: []WeekRenderData{
			{
				TargetKey:  "current",
				Title:      "Denne uge",
				Plan:       testPlan,
				IsCurrent:  true,
				WeekNum:    28,
				OffsetDays: 0,
			},
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
		"getDayDate": func(key string, lang string, offsetDays int) string {
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
		NextWeekPlan: WeekPlan{
			Monday: Day{
				DayNameDa: "Mandag",
				DayNameEn: "Monday",
				Tasks: []Task{
					{ID: "planned_task", Title: "Planned Task", Done: true},
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

	// 2. Transition when AutoResetWeek is enabled (wipes/resets both weeks)
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
	if len(currentState.NextWeekPlan.Monday.Tasks) != 1 || currentState.NextWeekPlan.Monday.Tasks[0].Title != "Template Task" {
		t.Error("Expected NextWeekPlan to automatically reset to TemplatePlan when week changes and AutoResetWeek is true")
	}
	stateMu.RUnlock()

	// 3. Transition when AutoResetWeek is disabled (shifts NextWeekPlan into WeekPlan)
	stateMu.Lock()
	currentState.LastWeekNum = currentWeek - 1
	if currentState.LastWeekNum <= 0 {
		currentState.LastWeekNum = 52
	}
	currentState.Settings.AutoResetWeek = false
	// Set mock NextWeekPlan with planned tasks
	currentState.NextWeekPlan.Monday.Tasks = []Task{{ID: "planned_task_from_previous_week", Title: "Planned Task", Done: true}}
	// Set mock WeekPlan with active tasks that should be shifted out
	currentState.WeekPlan.Monday.Tasks = []Task{{ID: "expired_task", Title: "Expired Task", Done: true}}
	stateMu.Unlock()

	checkWeekTransition()

	stateMu.RLock()
	if currentState.LastWeekNum != currentWeek {
		t.Errorf("Expected LastWeekNum to update to %d, got %d", currentWeek, currentState.LastWeekNum)
	}
	if len(currentState.WeekPlan.Monday.Tasks) != 1 || currentState.WeekPlan.Monday.Tasks[0].ID != "planned_task_from_previous_week" {
		t.Errorf("Expected WeekPlan to shift from NextWeekPlan when AutoResetWeek is false, got length %d", len(currentState.WeekPlan.Monday.Tasks))
	}
	if len(currentState.NextWeekPlan.Monday.Tasks) != 1 || currentState.NextWeekPlan.Monday.Tasks[0].Title != "Template Task" {
		t.Error("Expected NextWeekPlan to be populated with fresh template tasks after shifting")
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

func TestSaveTemplateUpdatesBothWeeks(t *testing.T) {
	stateMu.Lock()
	currentState = AppState{
		WeekPlan: WeekPlan{
			Sunday: Day{
				Tasks: []Task{},
			},
		},
		NextWeekPlan: WeekPlan{
			Sunday: Day{
				Tasks: []Task{},
			},
		},
		TemplatePlan: WeekPlan{
			Sunday: Day{
				Tasks: []Task{},
			},
		},
	}
	stateMu.Unlock()

	// Simulate receiving a POST request on /api/save-template with a task on Sunday
	newTemplate := WeekPlan{
		Sunday: Day{
			Tasks: []Task{
				{Time: "12:00", Title: "Sunday Template Task", Color: "default"},
			},
		},
	}

	stateMu.Lock()
	currentState.TemplatePlan = newTemplate
	resetWeekPlanToTemplate()
	resetNextWeekPlanToTemplate()
	stateMu.Unlock()

	// Verify that both current week and next week plans have the Sunday task!
	stateMu.RLock()
	defer stateMu.RUnlock()

	if len(currentState.WeekPlan.Sunday.Tasks) != 1 {
		t.Errorf("Expected current week Sunday to have 1 task, got %d", len(currentState.WeekPlan.Sunday.Tasks))
	} else if currentState.WeekPlan.Sunday.Tasks[0].Title != "Sunday Template Task" {
		t.Errorf("Expected current week Sunday task title to be 'Sunday Template Task', got '%s'", currentState.WeekPlan.Sunday.Tasks[0].Title)
	}

	if len(currentState.NextWeekPlan.Sunday.Tasks) != 1 {
		t.Errorf("Expected next week Sunday to have 1 task, got %d", len(currentState.NextWeekPlan.Sunday.Tasks))
	} else if currentState.NextWeekPlan.Sunday.Tasks[0].Title != "Sunday Template Task" {
		t.Errorf("Expected next week Sunday task title to be 'Sunday Template Task', got '%s'", currentState.NextWeekPlan.Sunday.Tasks[0].Title)
	}
}

func TestResetWeekPreservesAdHoc(t *testing.T) {
	stateMu.Lock()
	currentState = AppState{
		WeekPlan: WeekPlan{
			Monday: Day{
				Tasks: []Task{
					{ID: "template_task_id", Title: "Template Task", Done: true, AdHoc: false},
					{ID: "adhoc_task_id", Title: "Adhoc Task", Done: false, AdHoc: true},
				},
			},
		},
		TemplatePlan: WeekPlan{
			Monday: Day{
				Tasks: []Task{
					{Title: "Template Task Blueprint", AdHoc: false},
				},
			},
		},
	}
	stateMu.Unlock()

	resetWeekPlanToTemplate()

	stateMu.RLock()
	defer stateMu.RUnlock()

	mondayTasks := currentState.WeekPlan.Monday.Tasks
	// We expect 2 tasks: 1 template task from TemplatePlan, and 1 adhoc task preserved!
	if len(mondayTasks) != 2 {
		t.Fatalf("Expected 2 tasks on Monday, got %d", len(mondayTasks))
	}

	// Verify adhoc task is preserved
	var hasAdhoc bool
	var hasTemplate bool
	for _, task := range mondayTasks {
		if task.AdHoc {
			if task.ID == "adhoc_task_id" && task.Title == "Adhoc Task" {
				hasAdhoc = true
			}
		} else {
			if task.Title == "Template Task Blueprint" && !task.Done {
				hasTemplate = true
			}
		}
	}

	if !hasAdhoc {
		t.Error("Expected existing ad-hoc task to be preserved during reset")
	}
	if !hasTemplate {
		t.Error("Expected standard template tasks to be reset to template blueprint")
	}
}

func TestSettingsMigration(t *testing.T) {
	// Create a test file representing a plan.json without layout settings (older schema)
	testPath := "test_old_settings.json"
	content := `{
		"settings": {
			"language": "en"
		}
	}`
	err := os.WriteFile(testPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create old settings mock: %v", err)
	}
	defer os.Remove(testPath)

	// Route loadPlan to mock file
	oldPath := planPath
	planPath = testPath
	defer func() {
		planPath = oldPath
	}()

	err = loadPlan()
	if err != nil {
		t.Fatalf("Expected loadPlan to succeed on older config layout, got %v", err)
	}

	stateMu.RLock()
	defer stateMu.RUnlock()

	if currentState.Settings.Language != "en" {
		t.Errorf("Expected language 'en' to be preserved, got '%s'", currentState.Settings.Language)
	}
	if currentState.Settings.DesktopLayout != "horizontal" {
		t.Errorf("Expected desktop layout to migrate/fallback to 'horizontal', got '%s'", currentState.Settings.DesktopLayout)
	}
	if currentState.Settings.MobileLayout != "vertical" {
		t.Errorf("Expected mobile layout to migrate/fallback to 'vertical', got '%s'", currentState.Settings.MobileLayout)
	}
}
