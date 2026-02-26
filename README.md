<div align='center'>
    <h1>🗺️ Captain ⛵</h1>
</div>

<div align="center">
    Log book, recording and documenting
</div>

## Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Quick Install

```bash
# Clone the repository
git clone https://github.com/yourusername/captain.git
cd captain

# Build the binary
go build -o captain

# Move to your PATH (optional)
sudo mv captain /usr/local/bin/

# Or add to your PATH
export PATH="$PATH:$(pwd)"
```

### First Run

```bash
# Initialize captain (creates config and database)
captain log
```

Captain will create a `~/.captain` directory with:
- `config.ini` - Configuration file
- `testdo.db` - SQLite database

Do Types:

- `task`: Things to do
- `tell`: Things to tell to <someone>
- `brag`: Things to be proud of
- `ask`: Things to ask <someone>
- `learn`: Things to learn or reference

```
       do                            at               doc  type   prio    for
☐  4   ⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿  22-Jan-25 22:36  +    tell   high
☐  16  ⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿                   03-Mar-25 20:34       tell   medium  alice
☐  13  How to deal with stress       03-Mar-25 12:26       learn  medium
☐  10  I need to complete this       27-Feb-25 21:28  +    task   low
▣  15  How do we do this?            03-Mar-25 20:34       tell   medium  alice
▣  14  How do we tag                 03-Mar-25 20:34       ask    medium  john
▣  12  Dealing with outages          03-Mar-25 12:15  +    learn  medium
```

## Usage:

Create a task

```
$ captain do 'I should do this'
```

Complete a task

```
$ captain did <do.id>
```

Delete a task

```
$ captain scratch <do.id>
```

Revert deletion

```
$ captain unscratch <do.id>
```

Set priority

```
$ captain set prio <high/med/low> <do.id>
```

Set type

```
$ captain set type <task/ask/tell/learn/brag/PR/meta> <do.id>
```

### Attributes

Pin a do

```
$ captain pin <do.id>
```

Unpin a do
```
$ captain unpin <do.id>
```

Mark a do

```
$ captain mark <do.id> <sensitive>
```

Unmark a do

```
$ captain unmark <do.id> <sensitive>
```

### Views

Log all do, we can also pass arguements for filters. By default sensitive do's are hidden.

```
$ captain log
$ captain log --for standup
$ captain log --for alice --type tell
$ captain log --all
$ captain log --order asc/desc --sort priority
$ captain log --unhide
```

View pinned do

```
$ captain pinned
```

### Crew

Add a mate to the crew

```
$ captain recruit <name>
```

List the crew

```
$ captain crew
```

Rename a crew member

```
$ captain rename <oldname> <newname>
```

### Assignment

Ask something to someone

```
$ captain ask <name> 'Question'
```

Tell something to someone

```
$ captain tell <name> 'Something'
```

### Personal Development

Note an achievement

```
$ captain brag 'Overcame the odds'
```

Revisit or revise something

```
$ captain learn 'Something new'
```

### Going Deep

Write a document on the do

```
$ captain doc <do.id>
```

View the do's document

```
$ captain view <do.id>
```

### Templates

Templates allow you to create reusable task structures with placeholders that get filled in when creating a new task. Template output is saved as task documentation while the command-line description remains as the task title.

List all templates

```
$ captain templates
$ captain template list
```

Create a new template

```
$ captain template create <name>
```

This opens your `$EDITOR` where you can write a template with mostxt placeholders:

```markdown
Meeting with {{ person:string example('John Doe') }}

Date: {{ meeting_date:datetime 'YYYY-MM-DD' }}

Agenda:
{{ agenda:list describe('Enter agenda items, one per line. Empty line to finish.') }}

Notes:
{{ notes }}
```

Edit an existing template

```
$ captain template edit <name>
```

Delete a template

```
$ captain template delete <name>
```

Use a template to create a task

```
$ captain do 'Set up 121 meeting' --template 121
$ captain do 'Set up 121 meeting' -t 121
```

When using a template:
1. You'll be prompted to fill in each placeholder
2. The task description will be your command-line message
3. The filled template will be saved as task documentation (viewable with `captain view <id>`)

### Config

```
profile = main

[main]
dbname        = do.db
lookback_days = 14
CaptainDir    = ~/.captain
log_length    = 20
```

- `profile`: can be used to setup different config groups
- `dbname`: change db
- `lookback_days`: default number of days `captain log` shows
- `CaptainDir`: location to save config and db
- `log_length`: default max number of items to show on `captain log`


### SQLite

```
sqlite3 ~/.captain/testdo.db

.mode column
.headers on
```
