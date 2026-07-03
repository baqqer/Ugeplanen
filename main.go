package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Task represents a single duty/task in a day.
type Task struct {
	ID    string `json:"id"`
	Time  string `json:"time"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Color string `json:"color"`           // "default", "red", "green", "blue", "yellow", "purple"
	AdHoc bool   `json:"adhoc,omitempty"` // Is this an ad-hoc task added on the fly?
}

// Day represents a single day of the week containing its tasks.
type Day struct {
	DayNameDa string `json:"day_name_da"`
	DayNameEn string `json:"day_name_en"`
	Tasks     []Task `json:"tasks"`
}

// GetName returns the translated day name.
func (d Day) GetName(lang string) string {
	if lang == "en" {
		return d.DayNameEn
	}
	return d.DayNameDa
}

// WeekPlan holds the seven days of the week.
type WeekPlan struct {
	Monday    Day `json:"monday"`
	Tuesday   Day `json:"tuesday"`
	Wednesday Day `json:"wednesday"`
	Thursday  Day `json:"thursday"`
	Friday    Day `json:"friday"`
	Saturday  Day `json:"saturday"`
	Sunday    Day `json:"sunday"`
}

// DayWithKey pairs a JSON key with Day data to maintain order in templates.
type DayWithKey struct {
	Key string
	Day Day
}

// Days returns a slice of days ordered from Monday to Sunday.
func (wp WeekPlan) Days() []DayWithKey {
	return []DayWithKey{
		{Key: "monday", Day: wp.Monday},
		{Key: "tuesday", Day: wp.Tuesday},
		{Key: "wednesday", Day: wp.Wednesday},
		{Key: "thursday", Day: wp.Thursday},
		{Key: "friday", Day: wp.Friday},
		{Key: "saturday", Day: wp.Saturday},
		{Key: "sunday", Day: wp.Sunday},
	}
}

// Settings stores user configuration like language and layouts.
type Settings struct {
	Language          string `json:"language"`            // "da" or "en"
	DesktopLayout     string `json:"desktop_layout"`      // "horizontal" or "vertical"
	MobileLayout      string `json:"mobile_layout"`       // "horizontal" or "vertical"
	ShowPassedDays    bool   `json:"show_passed_days"`    // true or false
	HighlightToday    bool   `json:"highlight_today"`     // true or false
	ShowDates         bool   `json:"show_dates"`          // true or false
	ShowWeekNumber    bool   `json:"show_week_number"`    // true or false
	TouchFriendlyMode bool   `json:"touch_friendly_mode"` // true or false (larger touch targets)
	RowTapToggle      bool   `json:"row_tap_toggle"`      // true or false (tap entire task row to toggle done)
	AutoResetWeek     bool   `json:"auto_reset_week"`     // true or false (automatically reset week plan when new week starts)
}

// AppState represents the complete plan.json data structure.
type AppState struct {
	Settings     Settings `json:"settings"`
	WeekPlan     WeekPlan `json:"week_plan"`
	TemplatePlan WeekPlan `json:"template_plan"`
	LastWeekNum  int      `json:"last_week_num"` // Stores the last ISO week number rendered/processed
}

var (
	currentState AppState
	stateMu      sync.RWMutex
	planPath     = "plan.json"
)

// UI Translations
var translations = map[string]map[string]string{
	"da": {
		"title":                  "Ugeplanen",
		"edit_plan":              "Rediger Ugeplan",
		"back_to_dashboard":      "Tilbage til Dashboard",
		"save":                  "Gem ugeplan",
		"save_settings":         "Gem indstillinger",
		"add_task":              "Tilføj Opgave",
		"delete":                "Slet",
		"time":                  "Tid",
		"task":                  "Opgave",
		"done":                  "Udført",
		"no_tasks":              "Ingen opgaver for denne dag.",
		"toggle_lang":           "English",
		"language_label":        "Sprog",
		"other_lang_code":       "en",
		"day":                   "Dag",
		"actions":               "Handlinger",
		"save_success":          "Ugeplanen blev gemt succesfuldt!",
		"confirm_discard":       "Er du sikker på, at du vil forlade siden? Ulæste ændringer vil gå tabt.",
		"today":                 "I dag",
		"apply_template":        "Hent fra skabelon",
		"save_template_success": "Skabelonen blev gemt succesfuldt!",
		"apply_template_success": "Skabelon hentet! Husk at gemme din ugeplan.",
		"confirm_apply":         "Dette vil overskrive dine ændringer i editoren. Vil du fortsætte?",
		"copy":                  "Kopier",
		"copy_to":               "Kopier til...",
		"copy_target_title":     "Vælg dage at kopiere til",
		"select_all":            "Vælg alle",
		"deselect_all":          "Fravælg alle",
		"cancel":                "Annuller",
		"copy_success":          "Opgave kopieret!",
		"color":                 "Farve",
		"color_default":         "Standard",
		"color_red":             "Rød",
		"color_green":           "Grøn",
		"color_blue":            "Blå",
		"color_yellow":          "Gul",
		"color_purple":          "Lilla",
		"settings":              "Indstillinger",
		"layout_settings_header": "Layoutindstillinger",
		"additional_settings_header": "Flere indstillinger",
		"supported_languages_helper": "Understøttede sprog: Dansk (da) og Engelsk (en)",
		"desktop_layout_label":  "Layout på computer (Desktop)",
		"mobile_layout_label":   "Layout på mobil (Mobile)",
		"layout_horizontal":     "Vandret (Kolonner)",
		"layout_vertical":       "Lodret (Liste)",
		"show_passed_days_label": "Vis ugedage der er passeret",
		"highlight_today_label":  "Fremhæv nuværende ugedag",
		"show_dates_label":      "Vis datoer på ugedage",
		"show_week_number_label": "Vis ugenummer på dashboard",
		"touch_friendly_mode_label": "Touchvenlig tilstand (Større knapper & felter)",
		"row_tap_toggle_label":      "Tryk på hele opgaven for at markere afsluttet",
		"auto_reset_week_label":     "Nulstil automatisk opgaver når en ny uge starter",
		"week":                  "Uge",
		"manage_templates":      "Rediger skabelon",
		"save_template":         "Gem skabelon",
		"reset_week":            "Nulstil uge til skabelon",
		"confirm_reset":         "Er du sikker på, at du vil nulstille denne uges plan til skabelon-skabelonen? Dine uge-specifikke tilføjelser vil gå tabt.",
		"reset_success":         "Ugen blev nulstillet succesfuldt!",
		"adhoc":                 "Ad-hoc",
	},
	"en": {
		"title":                  "Ugeplanen",
		"edit_plan":              "Edit Week Plan",
		"back_to_dashboard":      "Back to Dashboard",
		"save":                  "Save Week Plan",
		"save_settings":         "Save settings",
		"add_task":              "Add Task",
		"delete":                "Delete",
		"time":                  "Time",
		"task":                  "Task",
		"done":                  "Done",
		"no_tasks":              "No tasks for this day.",
		"toggle_lang":           "Dansk",
		"language_label":        "Language",
		"other_lang_code":       "da",
		"day":                   "Day",
		"actions":               "Actions",
		"save_success":          "Week plan saved successfully!",
		"confirm_discard":       "Are you sure you want to leave? Unsaved changes will be lost.",
		"today":                 "Today",
		"apply_template":        "Apply Template",
		"save_template_success": "Template saved successfully!",
		"apply_template_success": "Template applied! Remember to save your week plan.",
		"confirm_apply":         "This will overwrite your changes in the editor. Do you want to continue?",
		"copy":                  "Copy",
		"copy_to":               "Copy to...",
		"copy_target_title":     "Select days to copy to",
		"select_all":            "Select all",
		"deselect_all":          "Deselect all",
		"cancel":                "Cancel",
		"copy_success":          "Task copied!",
		"color":                 "Color",
		"color_default":         "Default",
		"color_red":             "Red",
		"color_green":           "Green",
		"color_blue":            "Blue",
		"color_yellow":          "Yellow",
		"color_purple":          "Purple",
		"settings":              "Settings",
		"layout_settings_header": "Layout Settings",
		"additional_settings_header": "Additional Settings",
		"supported_languages_helper": "Supported languages: Danish (da) and English (en)",
		"desktop_layout_label":  "Desktop Layout",
		"mobile_layout_label":   "Mobile Layout",
		"layout_horizontal":     "Horizontal (Columns)",
		"layout_vertical":       "Vertical (List)",
		"show_passed_days_label": "Show weekdays already passed",
		"highlight_today_label":  "Highlight current weekday",
		"show_dates_label":      "Show dates on weekdays",
		"show_week_number_label": "Show week number on dashboard",
		"touch_friendly_mode_label": "Touch Friendly Mode (Larger buttons & targets for mobile)",
		"row_tap_toggle_label":      "Tap entire task row to toggle done status",
		"auto_reset_week_label":     "Automatically reset week plan tasks when a new week starts",
		"week":                  "Week",
		"manage_templates":      "Edit Template",
		"save_template":         "Save Template",
		"reset_week":            "Reset Week to Template",
		"confirm_reset":         "Are you sure you want to reset this week's plan to the template blueprint? Your week-specific additions will be lost.",
		"reset_success":         "Week reset successfully!",
		"adhoc":                 "Ad-hoc",
	},
}

// loadPlan reads the state from the JSON file or initializes a default one.
func loadPlan() error {
	stateMu.Lock()
	defer stateMu.Unlock()

	file, err := os.ReadFile(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize default plan
			_, currentWeek := time.Now().ISOWeek()
			currentState = AppState{
				Settings: Settings{
					Language:          "da",
					DesktopLayout:     "horizontal",
					MobileLayout:      "vertical",
					ShowPassedDays:    true,
					HighlightToday:    true,
					ShowDates:         true,
					ShowWeekNumber:    true,
					TouchFriendlyMode: false,
					RowTapToggle:      false,
					AutoResetWeek:     false,
				},
				WeekPlan: WeekPlan{
					Monday:    Day{DayNameDa: "Mandag", DayNameEn: "Monday", Tasks: []Task{}},
					Tuesday:   Day{DayNameDa: "Tirsdag", DayNameEn: "Tuesday", Tasks: []Task{}},
					Wednesday: Day{DayNameDa: "Onsdag", DayNameEn: "Wednesday", Tasks: []Task{}},
					Thursday:  Day{DayNameDa: "Torsdag", DayNameEn: "Thursday", Tasks: []Task{}},
					Friday:    Day{DayNameDa: "Fredag", DayNameEn: "Friday", Tasks: []Task{}},
					Saturday:  Day{DayNameDa: "Lørdag", DayNameEn: "Saturday", Tasks: []Task{}},
					Sunday:    Day{DayNameDa: "Søndag", DayNameEn: "Sunday", Tasks: []Task{}},
				},
				TemplatePlan: WeekPlan{
					Monday:    Day{DayNameDa: "Mandag", DayNameEn: "Monday", Tasks: []Task{}},
					Tuesday:   Day{DayNameDa: "Tirsdag", DayNameEn: "Tuesday", Tasks: []Task{}},
					Wednesday: Day{DayNameDa: "Onsdag", DayNameEn: "Wednesday", Tasks: []Task{}},
					Thursday:  Day{DayNameDa: "Torsdag", DayNameEn: "Thursday", Tasks: []Task{}},
					Friday:    Day{DayNameDa: "Fredag", DayNameEn: "Friday", Tasks: []Task{}},
					Saturday:  Day{DayNameDa: "Lørdag", DayNameEn: "Saturday", Tasks: []Task{}},
					Sunday:    Day{DayNameDa: "Søndag", DayNameEn: "Sunday", Tasks: []Task{}},
				},
				LastWeekNum: currentWeek,
			}
			stateMu.Unlock()
			errSave := savePlanAtomic()
			stateMu.Lock()
			return errSave
		}
		return err
	}

	return json.Unmarshal(file, &currentState)
}

// savePlanAtomic writes current state atomically to plan.json
func savePlanAtomic() error {
	stateMu.RLock()
	data, err := json.MarshalIndent(currentState, "", "  ")
	stateMu.RUnlock()
	if err != nil {
		return err
	}

	// Create a temp file in the same directory to guarantee atomic rename
	tmpFile, err := os.CreateTemp(".", "plan.*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()

	if _, err = tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	if err = tmpFile.Close(); err != nil {
		return err
	}

	renameErr := os.Rename(tmpName, planPath)
	if renameErr != nil {
		// If rename fails (e.g. because of Docker single-file bind mount block),
		// fall back to direct file write (truncate and write).
		log.Printf("Warning: atomic rename failed (%v), falling back to direct write for %s", renameErr, planPath)
		
		f, errWrite := os.OpenFile(planPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if errWrite != nil {
			return fmt.Errorf("direct write fallback failed: %v (original rename error: %v)", errWrite, renameErr)
		}
		defer f.Close()
		
		if _, errWrite = f.Write(data); errWrite != nil {
			return fmt.Errorf("direct write fallback write failed: %v (original rename error: %v)", errWrite, renameErr)
		}
		
		// Remove the temp file now that we succeeded via fallback
		_ = os.Remove(tmpName)
		return nil
	}
	return nil
}

// resetWeekPlanToTemplate resets currentState.WeekPlan to currentState.TemplatePlan tasks (with fresh IDs)
func resetWeekPlanToTemplate() {
	cloneTasks := func(src Day, dayKey string) Day {
		cloned := []Task{}
		for i, t := range src.Tasks {
			cloned = append(cloned, Task{
				ID:    fmt.Sprintf("task_%s_%d_%d", dayKey, time.Now().UnixNano(), i),
				Time:  t.Time,
				Title: t.Title,
				Done:  false,
				Color: t.Color,
			})
		}
		return Day{
			DayNameDa: src.DayNameDa,
			DayNameEn: src.DayNameEn,
			Tasks:     cloned,
		}
	}

	currentState.WeekPlan.Monday = cloneTasks(currentState.TemplatePlan.Monday, "monday")
	currentState.WeekPlan.Tuesday = cloneTasks(currentState.TemplatePlan.Tuesday, "tuesday")
	currentState.WeekPlan.Wednesday = cloneTasks(currentState.TemplatePlan.Wednesday, "wednesday")
	currentState.WeekPlan.Thursday = cloneTasks(currentState.TemplatePlan.Thursday, "thursday")
	currentState.WeekPlan.Friday = cloneTasks(currentState.TemplatePlan.Friday, "friday")
	currentState.WeekPlan.Saturday = cloneTasks(currentState.TemplatePlan.Saturday, "saturday")
	currentState.WeekPlan.Sunday = cloneTasks(currentState.TemplatePlan.Sunday, "sunday")
}

// checkWeekTransition checks if the calendar week has changed. If so, it updates the saved week number.
// If the AutoResetWeek setting is enabled, it automatically resets the week's plan to the standard template.
func checkWeekTransition() {
	_, currentWeek := time.Now().ISOWeek()

	stateMu.Lock()
	if currentState.LastWeekNum == 0 {
		currentState.LastWeekNum = currentWeek
		stateMu.Unlock()
		savePlanAtomic()
		return
	}

	if currentState.LastWeekNum != currentWeek {
		log.Printf("New week transition detected! From week %d to week %d.", currentState.LastWeekNum, currentWeek)
		currentState.LastWeekNum = currentWeek
		
		if currentState.Settings.AutoResetWeek {
			log.Println("AutoResetWeek is enabled. Automatically resetting tasks to standard template...")
			resetWeekPlanToTemplate()
		}
		stateMu.Unlock()
		savePlanAtomic()
		return
	}
	stateMu.Unlock()
}

// TemplateData passed to html templates
type TemplateData struct {
	Language       string
	State          AppState
	Trans          map[string]string
	DayKeys        []string
	CurrentDay     string
	CurrentWeekNum int
}

func getTemplateData() TemplateData {
	stateMu.RLock()
	defer stateMu.RUnlock()

	lang := currentState.Settings.Language
	if lang != "da" && lang != "en" {
		lang = "da"
	}

	// Calculate current day
	weekday := time.Now().Weekday()
	var currentDay string
	switch weekday {
	case time.Monday:
		currentDay = "monday"
	case time.Tuesday:
		currentDay = "tuesday"
	case time.Wednesday:
		currentDay = "wednesday"
	case time.Thursday:
		currentDay = "thursday"
	case time.Friday:
		currentDay = "friday"
	case time.Saturday:
		currentDay = "saturday"
	case time.Sunday:
		currentDay = "sunday"
	}

	_, weekNum := time.Now().ISOWeek()

	return TemplateData{
		Language:       lang,
		State:          currentState,
		Trans:          translations[lang],
		DayKeys:        []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"},
		CurrentDay:     currentDay,
		CurrentWeekNum: weekNum,
	}
}

var daMonths = []string{"jan", "feb", "mar", "apr", "maj", "jun", "jul", "aug", "sep", "okt", "nov", "dec"}

func getWeekDateString(dayIndex int, lang string) string {
	now := time.Now()
	weekday := int(now.Weekday())
	daysFromMonday := weekday - 1
	if weekday == 0 {
		daysFromMonday = 6
	}
	mondayDate := now.AddDate(0, 0, -daysFromMonday)
	targetDate := mondayDate.AddDate(0, 0, dayIndex)

	if lang == "en" {
		return targetDate.Format("Jan 2")
	}
	monthIdx := int(targetDate.Month()) - 1
	return fmt.Sprintf("%d. %s", targetDate.Day(), daMonths[monthIdx])
}

func getDayIndex(key string) int {
	switch key {
	case "monday":
		return 1
	case "tuesday":
		return 2
	case "wednesday":
		return 3
	case "thursday":
		return 4
	case "friday":
		return 5
	case "saturday":
		return 6
	case "sunday":
		return 7
	}
	return 0
}

func sortTasks(tasks []Task) {
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].Time > tasks[j].Time {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
}

// getLocalIP returns the first non-loopback local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	return lrw.ResponseWriter.Write(b)
}

// loggingMiddleware logs incoming HTTP requests to stdout
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default status
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		durationMs := float64(duration.Nanoseconds()) / 1e6

		log.Printf("[%s] %s %s %s - %d %s (%.3fms)",
			start.Format("2006-01-02 15:04:05"),
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			lrw.statusCode,
			http.StatusText(lrw.statusCode),
			durationMs,
		)
	})
}

func main() {
	if err := loadPlan(); err != nil {
		log.Fatalf("Error loading plan database: %v", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Template mapping with a dynamic Day Name Translation func
	funcMap := template.FuncMap{
		"getDayName": func(day Day, lang string) string {
			return day.GetName(lang)
		},
		"getDayIndex": func(key string) int {
			return getDayIndex(key)
		},
		"getDayDate": func(key string, lang string) string {
			idx := getDayIndex(key) - 1
			if idx < 0 || idx > 6 {
				return ""
			}
			return getWeekDateString(idx, lang)
		},
	}

	// Dashboard Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Check and trigger calendar week transition resets if applicable
		checkWeekTransition()

		tmpl, err := template.New("dashboard.html").Funcs(funcMap).ParseFiles(
			filepath.Join("templates", "dashboard.html"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}

		data := getTemplateData()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error rendering dashboard: %v", err)
		}
	})

	// Application Settings Page Handler
	http.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("settings.html").Funcs(funcMap).ParseFiles(
			filepath.Join("templates", "settings.html"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}

		data := getTemplateData()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error rendering settings: %v", err)
		}
	})

	// Edit Standard Template Page Handler
	http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("manage_templates.html").Funcs(funcMap).ParseFiles(
			filepath.Join("templates", "manage_templates.html"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}

		data := getTemplateData()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error rendering template editor: %v", err)
		}
	})

	// API: Toggle task checkbox status
	http.HandleFunc("/api/toggle-task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Day    string `json:"day"`
			TaskID string `json:"task_id"`
			Done   bool   `json:"done"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		stateMu.Lock()
		var found bool

		updateTasks := func(day *Day) {
			for i, t := range day.Tasks {
				if t.ID == req.TaskID {
					day.Tasks[i].Done = req.Done
					found = true
					break
				}
			}
		}

		switch req.Day {
		case "monday":
			updateTasks(&currentState.WeekPlan.Monday)
		case "tuesday":
			updateTasks(&currentState.WeekPlan.Tuesday)
		case "wednesday":
			updateTasks(&currentState.WeekPlan.Wednesday)
		case "thursday":
			updateTasks(&currentState.WeekPlan.Thursday)
		case "friday":
			updateTasks(&currentState.WeekPlan.Friday)
		case "saturday":
			updateTasks(&currentState.WeekPlan.Saturday)
		case "sunday":
			updateTasks(&currentState.WeekPlan.Sunday)
		}
		stateMu.Unlock()

		if !found {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving plan: %v", err)
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// API: Save layout and application settings
	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Settings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		stateMu.Lock()
		currentState.Settings = req
		stateMu.Unlock()

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving settings: %v", err)
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// API: Quick-add task from dashboard
	http.HandleFunc("/api/add-task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Day   string `json:"day"`
			Time  string `json:"time"`
			Title string `json:"title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		// Clean time input
		timeStr := req.Time
		if timeStr == "" {
			timeStr = "12:00"
		}

		stateMu.Lock()
		
		// Create a brand new independent task with a unique timestamp ID
		newTask := Task{
			ID:    fmt.Sprintf("task_%d", time.Now().UnixNano()),
			Time:  timeStr,
			Title: req.Title,
			Done:  false,
			Color: "default",
			AdHoc: true,
		}

		var day *Day
		switch req.Day {
		case "monday":
			day = &currentState.WeekPlan.Monday
		case "tuesday":
			day = &currentState.WeekPlan.Tuesday
		case "wednesday":
			day = &currentState.WeekPlan.Wednesday
		case "thursday":
			day = &currentState.WeekPlan.Thursday
		case "friday":
			day = &currentState.WeekPlan.Friday
		case "saturday":
			day = &currentState.WeekPlan.Saturday
		case "sunday":
			day = &currentState.WeekPlan.Sunday
		}

		if day == nil {
			stateMu.Unlock()
			http.Error(w, "Invalid day", http.StatusBadRequest)
			return
		}

		day.Tasks = append(day.Tasks, newTask)
		sortTasks(day.Tasks)
		stateMu.Unlock()

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving plan: %v", err)
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "task": newTask})
	})

	// API: Delete task from dashboard
	http.HandleFunc("/api/delete-task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Day    string `json:"day"`
			TaskID string `json:"task_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		stateMu.Lock()
		var found bool

		deleteTask := func(day *Day) {
			newTasks := []Task{}
			for _, t := range day.Tasks {
				if t.ID == req.TaskID {
					found = true
				} else {
					newTasks = append(newTasks, t)
				}
			}
			day.Tasks = newTasks
		}

		switch req.Day {
		case "monday":
			deleteTask(&currentState.WeekPlan.Monday)
		case "tuesday":
			deleteTask(&currentState.WeekPlan.Tuesday)
		case "wednesday":
			deleteTask(&currentState.WeekPlan.Wednesday)
		case "thursday":
			deleteTask(&currentState.WeekPlan.Thursday)
		case "friday":
			deleteTask(&currentState.WeekPlan.Friday)
		case "saturday":
			deleteTask(&currentState.WeekPlan.Saturday)
		case "sunday":
			deleteTask(&currentState.WeekPlan.Sunday)
		}
		stateMu.Unlock()

		if !found {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving plan: %v", err)
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// GET: Reset week plan back to standard template plan
	http.HandleFunc("/reset-week", func(w http.ResponseWriter, r *http.Request) {
		stateMu.Lock()
		resetWeekPlanToTemplate()
		stateMu.Unlock()

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving plan: %v", err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// API: Get standard template week plan
	http.HandleFunc("/api/get-template", func(w http.ResponseWriter, r *http.Request) {
		stateMu.RLock()
		tmpl := currentState.TemplatePlan
		stateMu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tmpl)
	})

	// API: Save standard template week plan
	http.HandleFunc("/api/save-template", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req WeekPlan
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		stateMu.Lock()
		currentState.TemplatePlan = req
		stateMu.Unlock()

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving template plan: %v", err)
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// POST/GET: Language switch toggle
	http.HandleFunc("/set-language", func(w http.ResponseWriter, r *http.Request) {
		lang := r.FormValue("lang")
		if lang != "da" && lang != "en" {
			lang = "da"
		}

		stateMu.Lock()
		currentState.Settings.Language = lang
		stateMu.Unlock()

		if err := savePlanAtomic(); err != nil {
			log.Printf("Error saving language: %v", err)
		}

		// Redirect back to referring page or homepage
		referer := r.Header.Get("Referer")
		if referer == "" {
			referer = "/"
		}
		http.Redirect(w, r, referer, http.StatusSeeOther)
	})

	port := "9000"
	localIP := getLocalIP()

	log.Printf("Ugeplanen server is running!")
	log.Printf("-> Local access: http://localhost:%s", port)
	if localIP != "localhost" {
		log.Printf("-> Network access: http://%s:%s", localIP, port)
	}
	log.Printf("Serving on 0.0.0.0:%s to support local network devices.", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, loggingMiddleware(http.DefaultServeMux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
