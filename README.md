# things

A fast CLI for Things 3 on macOS.

Reads directly from the Things 3 SQLite database for instant queries. Writes use the Things URL scheme with x-callback-url for reliable confirmation.

## Installation

```bash
# Build from source
go install github.com/codegangsta/things@latest

# Or clone and build
git clone https://github.com/codegangsta/things
cd things
go build -o things .
```

## Setup

For write operations (add, complete, update, delete), you need to set up an auth token:

1. Open Things 3
2. Go to **Settings → General → Enable Things URLs → Manage**
3. Copy your auth token
4. Run: `things auth <your-token>`

## Usage

### Reading Tasks

```bash
# View task lists
things today              # Today's tasks
things inbox              # Inbox items
things anytime            # Anytime tasks
things someday            # Someday tasks
things upcoming           # Scheduled tasks
things logbook            # Completed tasks
things trash              # Trashed tasks

# View by project/area
things projects           # List all projects
things areas              # List all areas
things list "Project"     # Tasks in a specific project or area

# Search and filter
things search "query"     # Search by title
things tagged "@phone"    # Filter by tag
things tags               # List all tags

# Get details
things get <id>           # Full details for a task, project, or area

# Statistics
things stats              # Task counts and statistics
```

### Output Options

```bash
things today --json       # JSON output
things today --brief      # Title and tags only
things today --count      # Just the count
things today --limit 5    # Limit results
```

### Writing Tasks

```bash
# Add tasks
things add "Buy groceries"
things add "Call mom" --when today
things add "Project deadline" --deadline 2024-12-31
things add "Work task" --list "Work" --tags "urgent,important"
things add "Shopping" --checklist "Milk" --checklist "Eggs"

# Add projects
things add-project "New Project" --area "Work"

# Bulk add from JSON
echo '[{"title": "Task 1"}, {"title": "Task 2"}]' | things add-json

# Complete tasks
things complete <id>
things complete <id1> <id2>   # Multiple tasks

# Delete (move to trash)
things delete <id>
things delete <id1> <id2>     # Multiple tasks

# Update tasks
things update <id> --title "New title"
things update <id> --when tomorrow
things update <id> --add-tags "@phone,5m"
things update <id> --notes "Added notes" --append
things update <id> --list "Project Name"
things update <id> --heading "Section"

# Update projects
things update-project <id> --title "New name"

# Manage checklists
things checklist <id> --append "New item"
things checklist <id> --prepend "First item"
```

### Write Options

```bash
things add "Task" --no-wait          # Don't wait for confirmation
things add "Task" --timeout 10s      # Custom timeout
```

## Examples

### Daily Review Workflow

```bash
# Check what's on today
things today

# Process inbox
things inbox
things update <id> --when today --tags "@computer,25m"

# Complete tasks as you work
things complete <id>
```

### Batch Operations

```bash
# Add multiple tasks from a file
cat tasks.json | things add-json

# Example tasks.json:
[
  {
    "title": "Review PR",
    "when": "today",
    "tags": ["@computer", "15m"],
    "list": "Work"
  },
  {
    "title": "Call dentist",
    "when": "tomorrow",
    "tags": ["@phone", "5m"]
  }
]
```

### Scripting

```bash
# Get task count for today
things today --count

# Export today's tasks as JSON
things today --json > today.json

# Find tasks by tag and output IDs
things tagged "@waiting" --json | jq -r '.[].id'
```

## Architecture

- **Reads**: Direct SQLite queries against the Things 3 database for speed
- **Writes**: Things URL scheme with x-callback-url for reliable confirmation
- **Database location**: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`

## License

MIT
