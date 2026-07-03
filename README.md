# Ugeplanen

Ugeplanen is a lightweight, self-contained weekly calendar and task planner designed for local networks, family dashboards, or home servers. It is built as a single-binary Go application with no external runtime dependencies. All configuration, settings, and task data are stored locally in a single, portable JSON file.

## Features

* **Master Template Blueprint**: Define a standard weekly schedule of recurring tasks (such as chores, lessons, or meals). The active week can be manually or automatically reset back to this blueprint at any time.
* **Ad-hoc Tasks**: Add quick, one-off tasks directly to any day on the dashboard. These tasks can be completed or deleted on the fly and are cleared out when a new week is started.
* **On-the-fly Localization**: Full English and Danish translations are supported. Language preference is saved in the user settings and dynamically updates the interface on reload.
* **Mobile Touch Support**: Customizable options for Touch-Friendly Mode (enlarging checkboxes, buttons, and input targets to exceed 44x44px requirements) and Row-Tap Toggle (allowing you to check off tasks by tapping anywhere on their card row).
* **Automatic Weekly Reset**: An optional setting that automatically detects a calendar ISO week transition on render and refreshes the active week with a fresh, uncompleted clone of the master template plan.
* **Zero-Configuration Database**: Uses a plain-text `plan.json` file. If the file does not exist, the server automatically initializes it with a default structure.

## Technical Stack

* **Backend**: Go (standard library `net/http`, `html/template`, and `encoding/json` are used exclusively).
* **Frontend**: Responsive Vanilla HTML and CSS (no build steps, bundlers, or CSS frameworks) and Vanilla JS for asynchronous updates (AJAX) and layout controls.
* **Database**: Atomic, single-file JSON storage (`plan.json`) with periodic state saving and memory caching.

## Project Structure

```text
ugeplanen/
├── main.go              # Core server, routing, endpoints, and persistence logic
├── main_test.go         # Unit and template rendering tests
├── plan.json            # JSON database (automatically generated on first run)
├── templates/
│   ├── dashboard.html   # Main calendar dashboard grid
│   ├── settings.html    # Configuration editor (layout, language, touch modes)
│   └── manage_templates.html # Master week template schedule editor
└── static/
    ├── css/
    │   └── style.css    # Layout, variables, and responsive styling rules
    └── js/
        └── app.js       # Core frontend scripts, settings storage, and AJAX handlers
```

## Getting Started

### Prerequisites

* Go 1.26 or higher

### Running the Application

1. Clone or download the repository to your local machine:
   ```bash
   cd ugeplanen
   ```

2. Run the application:
   ```bash
   go run main.go
   ```

3. Open your web browser and navigate to:
   ```text
   http://localhost:9000
   ```

The application binds to all available network interfaces (`0.0.0.0`), allowing other devices on your home or office local area network (LAN) to access the planner via your machine's local IP address (e.g., `http://192.168.1.15:9000`).

### Building a Single Binary

To build a compiled, production-ready binary with zero dependencies:

```bash
go build -o ugeplanen main.go
```

You can then copy and run the `ugeplanen` binary on your target machine.

## Running with Docker or Podman

The application can be packaged into a lightweight OCI-compliant container image using the provided multi-stage `Dockerfile`.

### Building the Image

To build the image locally, navigate to the project root directory:

**Using Docker:**
```bash
docker build -t ugeplanen .
```

**Using Podman:**
```bash
podman build -t ugeplanen .
```

### Running the Container with Persistent Storage

Because Ugeplanen is stateless outside of `plan.json`, you must mount the data file to the host system to ensure your plans and settings survive container updates or restarts.

#### Setup Data File
First, create an empty `plan.json` on your host if it does not already exist:
```bash
touch plan.json
```

#### Running the Container

**Using Docker:**
```bash
docker run -d \
  --name ugeplanen \
  -p 9000:9000 \
  -v $(pwd)/plan.json:/app/plan.json \
  --restart unless-stopped \
  ugeplanen
```

**Using Podman:**
On systems with SELinux (like Fedora, CentOS, or RHEL), append the `:Z` flag to your volume mount to label the file correctly for container write access:
```bash
podman run -d \
  --name ugeplanen \
  -p 9000:9000 \
  -v $(pwd)/plan.json:/app/plan.json:Z \
  --restart unless-stopped \
  ugeplanen
```

## Data Schema

The entire state is maintained in `plan.json` in the working directory:

```json
{
  "settings": {
    "language": "da",
    "desktop_layout": "horizontal",
    "mobile_layout": "vertical",
    "show_passed_days": true,
    "highlight_today": true,
    "show_dates": true,
    "show_week_number": true,
    "touch_friendly_mode": false,
    "row_tap_toggle": false,
    "auto_reset_week": false
  },
  "week_plan": {
    "monday": {
      "day_name_da": "Mandag",
      "day_name_en": "Monday",
      "tasks": []
    },
    ...
  },
  "template_plan": {
    "monday": {
      "day_name_da": "Mandag",
      "day_name_en": "Monday",
      "tasks": []
    },
    ...
  },
  "last_week_num": 27
}
```

## Testing

The application includes unit and template rendering test suites. To run tests:

```bash
go test -v ./...
```
