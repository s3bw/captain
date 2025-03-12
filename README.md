<div align='center'>
    <h1>🗺️ Captain ⛵</h1>
</div>

<div align="center">
    Log book, recording and documenting
</div>

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
```
