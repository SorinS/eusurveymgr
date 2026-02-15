# eusurveymgr — EUSurvey Management CLI Tool

## Overview

Go CLI tool for managing a private EUSurvey instance at `https://eusurvey.escoaladevalori.ro/eusurvey`. Consolidates survey listing, results export, PDF generation, and database queries into a single binary using Cobra for command/subcommand dispatch.

## Project Structure

```
eusurveymgr.git/
  main.go                     # Entry point: cmd.SetVersion() + cmd.Execute()
  go.mod
  Makefile
  .gitignore
  bin/
    eusurveymgr.json.example  # Example config
  config/
    config.go                 # JSON config + env var overrides
  log/
    log.go                    # Logger (from riasec)
  client/
    client.go                 # Client struct, New(), shared HTTP state
    basic.go                  # HTTP Basic Auth + status-aware helpers
    session.go                # Form login (CSRF + cookies) for PDF endpoints
    surveys.go                # Survey listing/metadata + XML sanitization
    results.go                # Async results export with polling
    pdf.go                    # PDF generation/download/readiness check
    tokens.go                 # Token management (BROKEN — endpoints don't exist)
  db/
    db.go                     # ConnectToMySQL
    surveys.go                # List surveys (latest version per UID)
    answers.go                # List answer sets, lookup UNIQUECODE, get responses
  cmd/
    root.go                   # Cobra root command, persistent flags, init
    surveys.go                # surveys list/info commands
    results.go                # results export command
    pdf.go                    # pdf survey/answer commands
    tokens.go                 # tokens list/create commands (BROKEN)
    db.go                     # db surveys/answers/lookup/responses commands
  docs/
    PLAN.md                   # This file
    EUSURVEY-API.md           # API reference with verified endpoints
    EUSURVEY-DB.md            # Database schema notes
```

## Configuration

### Config file (`eusurveymgr.json`)

```json
{
  "base_url": "https://eusurvey.escoaladevalori.ro/eusurvey",
  "web_user": "root",
  "web_password": "...",
  "db_host": "127.0.0.1",
  "db_port": 3306,
  "db_name": "eusurveydb",
  "db_user": "reportr",
  "db_password": "...",
  "output_dir": ".",
  "timeout_seconds": 120,
  "insecure_tls": false
}
```

### Environment variable overrides

Env vars override config file values (avoids exposing credentials on the command line):

| Variable | Config field |
|----------|-------------|
| `EUSURVEYMGR_WEB_USER` | `web_user` |
| `EUSURVEYMGR_WEB_PASSWORD` | `web_password` |
| `EUSURVEYMGR_DB_HOST` | `db_host` |
| `EUSURVEYMGR_DB_NAME` | `db_name` |
| `EUSURVEYMGR_DB_USER` | `db_user` |
| `EUSURVEYMGR_DB_PASSWORD` | `db_password` |

## Command Reference

### Global flags

```
eusurveymgr [--config file] [-v] <command> <subcommand> [flags]

  --config string   Path to config file (default "eusurveymgr.json")
  -v, --verbose     Verbose (debug) output
```

### surveys — Manage surveys via WebService API

```
eusurveymgr surveys list [--json]
```
List all surveys for the authenticated user. Uses HTTP Basic Auth against `/webservice/getMySurveys`.

```
eusurveymgr surveys info --alias <name> [--json]
```
Get survey metadata (type, status, language, security, contact, etc.). Uses `/webservice/getSurveyMetadata/{alias}`.

### results — Export survey results

```
eusurveymgr results export --id <surveyID|alias> [--output file] [--showids]
```
Start an async results export and poll until complete. The `--id` flag accepts both numeric survey IDs and aliases.

**Note**: This triggers a server-side export job that can be slow for large surveys. The server returns HTTP 201 with a task ID, then HTTP 204 while processing, and finally HTTP 200 with the XML data. The `timeout_seconds` config controls how long to poll.

### pdf — Download PDF documents

```
eusurveymgr pdf survey --alias <name> [--output file]
```
Download the survey form as PDF. Uses HTTP Basic Auth.

```
eusurveymgr pdf answer --code <uniquecode> [--output dir]
eusurveymgr pdf answer --email <addr> --survey <id> [--output dir]
```
Generate and download an answer PDF. The flow:
1. Check if PDF already exists (`/pdf/answerready/`) — skip generation if so
2. If not, trigger generation (`/worker/createanswerpdf/`)
3. Poll readiness until server returns `"exists"`
4. Download the PDF (`/pdf/answer/`)

With `--email`, does a DB lookup first to find the UNIQUECODE.

Output filename: `<answerSetID>--<email>.pdf` (with `--email`) or `<uniquecode>.pdf` (with `--code`).

### tokens — Manage invitation tokens (BROKEN)

```
eusurveymgr tokens list --survey <name> [--json]
eusurveymgr tokens create --survey <name>
```

**These commands do not work.** The upstream EUSurvey source has no `getTokens` or `createToken` endpoints. Token management uses a group-based API (`createNewTokenList`, `createTokens/{groupid}/{number}`) that requires a different implementation. See EUSURVEY-API.md for details.

### db — Query the MySQL database directly

```
eusurveymgr db surveys [--json]
```
List all surveys from MySQL (latest version per SURVEY_UID, deduplicated). Shows ID, UID, alias, title, published status, answer count, and creation date.

```
eusurveymgr db answers --survey <id> [--json]
```
List all answer sets (respondents) for a survey. Shows answer set ID, UNIQUECODE, date, name, and email. Name and email are extracted from PA_ID=0 (identity section): MIN(ANSWER_ID) = name, MAX(ANSWER_ID) = email.

```
eusurveymgr db lookup --email <addr> --survey <id>
```
Look up the ANSWER_SET_ID and UNIQUECODE for a specific respondent by email address.

```
eusurveymgr db responses --email <addr> --survey <id> [--json]
```
Show all answer values for a respondent. Joins ANSWERS with ELEMENTS to display question titles alongside values.

### version

```
eusurveymgr version
```
Print version, commit, and build date (set via ldflags in Makefile).

## Typical Workflows

### List respondents and download a PDF

```bash
# 1. Find the survey ID
eusurveymgr db surveys

# 2. List respondents
eusurveymgr db answers --survey 4609

# 3. Download PDF for a respondent (skips generation if already exists)
eusurveymgr pdf answer --code <uniquecode>
# or by email:
eusurveymgr pdf answer --email user@example.com --survey 4609
```

### View a respondent's answers

```bash
eusurveymgr db responses --email user@example.com --survey 4609
# or as JSON:
eusurveymgr db responses --email user@example.com --survey 4609 --json
```

### Export survey results as XML

```bash
eusurveymgr results export --id Check4SkillsInRomana --output results.xml
```
Note: requires `timeout_seconds` to be high enough (120-300s) for large surveys.

## Dependencies

- `github.com/spf13/cobra` — CLI framework (adds `github.com/spf13/pflag`, `github.com/inconshreveable/mousetrap`)
- `github.com/go-sql-driver/mysql` — MySQL driver

## Known Issues

1. **Token commands broken** — `getTokens`/`createToken` endpoints don't exist in upstream EUSurvey. Needs redesign around group-based token API.
2. **XML contains invalid UTF-8** — EUSurvey stores Romanian diacritics as Latin-1 but declares UTF-8. Client sanitizes before parsing.
3. **Results export can be very slow** — Server generates PDFs as a side-effect of `prepareResults`, controlled by `export.poolSize` in `spring.properties` (default 2 threads).
4. **`answerready` returns `"exists"`** — Not `"OK"` as one might expect. Client checks for both.
