# GEMINI.md - Ugeplanen Project Instructions

Welcome to **Ugeplanen**, a lightweight, locally-hosted week planner designed to support daily duties and planning. This project is optimized for "vibecoding"—focused on speed, simplicity, clean local execution, and high visual appeal without unnecessary complexity.

---

## 1. Overview & Core Philosophy
Ugeplanen is a self-contained, single-binary Go application meant to run on a local machine or home server, accessible to anyone on the local network. 

- **Local & Trust-Based:** No authentication or password protection is required. Both the dashboard and edit interfaces are open to anyone on the local network.
- **Zero-Dependency Storage:** All week plans, task data, and user settings are stored in a single, portable `plan.json` file.
- **Danish & English Support:** The dashboard supports on-the-fly language toggling (Danish/English). User-edited content is displayed exactly as entered, while system UI elements translate dynamically.
- **Monday-to-Sunday Focus:** The dashboard displays a clean, structured calendar-style week grid starting on Monday and ending on Sunday.

---

## 2. Tech Stack & Architecture
- **Backend:** Go (standard library `net/http`, `html/template`, and `encoding/json` are preferred for minimal external dependencies).
- **Frontend:** Vanilla HTML/CSS with light, responsive styling and zero/minimal frontend build steps. Modern CSS features (CSS Grid/Flexbox, CSS Variables, System Font Stacks) should be used to provide a polished, modern aesthetic.
- **Database:** A single `plan.json` file loaded into memory at startup, saved atomically on updates, and periodically persisted.

---

## 3. Data Schema (`plan.json`)
The entire application state is stored in a single, well-structured JSON file. 

```json
{
  "settings": {
    "language": "da" 
  },
  "week_plan": {
    "monday": {
      "day_name_da": "Mandag",
      "day_name_en": "Monday",
      "tasks": [
        { "id": "1", "time": "08:00", "title": "Morgenmad", "done": false },
        { "id": "2", "time": "18:00", "title": "Aftensmad", "done": false }
      ]
    },
    "tuesday": {
      "day_name_da": "Tirsdag",
      "day_name_en": "Tuesday",
      "tasks": []
    },
    "wednesday": {
      "day_name_da": "Onsdag",
      "day_name_en": "Wednesday",
      "tasks": []
    },
    "thursday": {
      "day_name_da": "Torsdag",
      "day_name_en": "Thursday",
      "tasks": []
    },
    "friday": {
      "day_name_da": "Fredag",
      "day_name_en": "Friday",
      "tasks": []
    },
    "saturday": {
      "day_name_da": "Lørdag",
      "day_name_en": "Saturday",
      "tasks": []
    },
    "sunday": {
      "day_name_da": "Søndag",
      "day_name_en": "Sunday",
      "tasks": []
    }
  }
}
```

### Constraints:
- User-edited content (like task titles and times) must be shown exactly as is, with proper HTML escaping to prevent XSS.
- The JSON file must auto-initialize with a default structure if it does not exist.

---

## 4. UI/UX & Visual Design Guidelines
Since users judge applications by their visual impact, Ugeplanen must feel modern, clean, and polished:

- **Theme & Aesthetics:** Use a modern, calming color scheme (e.g., slate/indigo or soft sage green) with dark mode/light mode awareness via CSS variables.
- **Layout:**
  - **Dashboard:** A card-based or grid-based layout representing the 7 days of the week, chronologically ordered from Monday to Sunday.
  - **Header:** Contains the application title, a toggle button to change language between "Danish" and "English", and an "Edit Week Plan" button.
  - **Interaction:** Checking or unchecking tasks should seamlessly update state (using simple AJAX/Fetch requests to avoid full-page reloads where possible).
- **Edit Interface:** A simple, intuitive form/table layout allowing users to add, modify, reorder, or delete tasks for each day. Changes must immediately save to `plan.json`.

---

## 5. Recommended Directory Structure
Keep the Go project neat, modular, and easy to run:

```
ugeplanen/
├── GEMINI.md            # This instruction file
├── main.go              # Application entry point, router, and server
├── plan.json            # Local JSON database (auto-created if missing)
├── templates/
│   ├── dashboard.html   # Main dashboard template
│   └── edit.html        # Plan editor template
└── static/
    ├── css/
    │   └── style.css    # Clean modern styles (Vanilla CSS)
    └── js/
        └── app.js       # Light interactivity (AJAX updates, toggles)
```

---

## 6. Development Workflow
To run and develop Ugeplanen locally:

1. **Run the server:**
   ```bash
   go run main.go
   ```
2. **Access the application:**
   - Open your browser to `http://localhost:8080` (or configured local port).
   - Ensure the server binds to `0.0.0.0` or local interfaces so other devices on your home network can access it (e.g., `http://<your-local-ip>:8080`).

---

## 7. Key Rules for AI Assistance
- **Security:** Do not bind to public WAN addresses; limit default hosting to local interfaces or clearly document local-network scope.
- **Simplicity Over Overengineering:** Stick to Go's robust standard library before pulling in heavy frameworks (like Gin or Fiber) unless absolute necessity arises.
- **Styling:** Rely entirely on Vanilla CSS. Avoid TailwindCSS or other CSS utility libraries to keep pages fast, self-contained, and easily customizable.
