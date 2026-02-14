# eusurveymgr — EUSurvey Management CLI Tool

## Context

The user manages a private EUSurvey instance at `https://eusurvey.escoaladevalori.ro/eusurvey` and currently relies on manual web UI interactions and ad-hoc shell scripts (like `bin/generate-pdf.sh`) to manage surveys, export results, and generate PDFs. This tool consolidates all EUSurvey management operations into a single Go CLI, following the same conventions as the sibling `riasec.git` project.

Target directory: `/Users/sorins/Dev/Go.Code/eusurveymgr.git/`

## Project Structure

```
eusurveymgr.git/
  main.go                     # Entry point + subcommand dispatcher
  go.mod
  Makefile
  .gitignore
  bin/
    eusurveymgr.json.example  # Example config
  config/
    config.go                 # JSON config loading (same pattern as riasec)
  log/
    log.go                    # Logger (copied from riasec)
  client/
    client.go                 # Client struct, New(), shared HTTP state
    basic.go                  # HTTP Basic Auth helpers (WebService API)
    session.go                # Form login (CSRF + cookies) for PDF endpoints
    surveys.go                # Survey listing/metadata methods + types
    results.go                # Results export methods + types
    pdf.go                    # Answer PDF generation/download methods
    tokens.go                 # Token management methods + types
  db/
    db.go                     # ConnectToMySQL (same pattern as riasec)
    surveys.go                # List/get surveys from MySQL
    answers.go                # List answer sets, lookup UNIQUECODE by email
  cmd/
    surveys.go                # "surveys" subcommand handler
    results.go                # "results" subcommand handler
    pdf.go                    # "pdf" subcommand handler
    tokens.go                 # "tokens" subcommand handler
    db.go                     # "db" subcommand handler
```

## Config File Format

`eusurveymgr.json`:
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
  "timeout_seconds": 30,
  "insecure_tls": false
}
```

## Command Structure

```
eusurveymgr -config <file> [-v] <command> <action> [flags]

surveys list                          # List surveys (via WebService API, HTTP Basic Auth)
surveys info  -alias <name>           # Get survey metadata

results export -id <surveyID> [-o file]  # Export survey results (async: prepare + poll + download)

pdf survey  -alias <name> [-o file]   # Download survey form PDF
pdf answer  -email <addr> -survey <id> [-o dir]  # Generate + download answer PDF (login + DB lookup + trigger + download)
pdf answer  -code <uniquecode> [-o dir]           # Same but with known uniquecode

tokens list   -survey <name>          # List invitation tokens
tokens create -survey <name>          # Create new token

db surveys                            # List surveys directly from MySQL
db answers  -survey <id>              # List answer sets for a survey
db lookup   -email <addr> -survey <id> # Look up UNIQUECODE by email
```

## Key Design Decisions

- **No cobra** — hand-rolled two-level subcommand dispatcher using `flag.FlagSet` per action (matches `go` tool pattern and riasec.git style)
- **Single `client.Client` struct** with two auth modes: Basic (stateless header) for `/webservice/*` and session-based (CSRF + cookie jar) for `/worker/*` and `/pdf/*`
- **Only dependency**: `github.com/go-sql-driver/mysql` — everything else uses stdlib (`net/http`, `net/http/cookiejar`, `encoding/xml`, `encoding/json`, `flag`)
- **Copy `log/log.go`** directly from riasec.git
- **Config pattern** mirrors riasec's `config.LoadFromFile` with a flat struct
- **DB access is optional** — commands that only need the web API work without DB config; `pdf answer -email` needs DB for UNIQUECODE lookup

## Implementation Order

### Phase 1: Skeleton
1. `go mod init eusurveymgr`, create `go.mod` (go 1.24.0)
2. Copy `log/log.go` from riasec (verbatim)
3. Create `config/config.go` — `Configuration` struct + `LoadFromFile`
4. Create `main.go` — global flags (`-config`, `-v`), subcommand dispatcher skeleton
5. Create `Makefile` (same cross-compilation pattern as riasec)
6. Create `.gitignore`, `bin/eusurveymgr.json.example`
7. Verify: `go build` compiles, `./eusurveymgr -h` prints usage

### Phase 2: HTTP Client + Surveys (Basic Auth)
1. `client/client.go` — `Client` struct with `*http.Client` + `CookieJar`, `New()`
2. `client/basic.go` — `doBasicGet()` / `doBasicRequest()` helpers setting `Authorization` header
3. `client/surveys.go` — `GetSurveys()`, `GetSurveyMetadata()` parsing XML responses, response types
4. `cmd/surveys.go` — `RunSurveys(cfg, args)` dispatching `list`/`info`
5. Verify: `eusurveymgr surveys list` returns survey list from the live instance

### Phase 3: Results Export
1. `client/results.go` — `PrepareResults()`, `GetResults()` with polling loop (1s backoff, configurable timeout)
2. `cmd/results.go` — `RunResults(cfg, args)` for `export` action
3. Verify: `eusurveymgr results export -id 4578 -o results.xml`

### Phase 4: Session Auth + Answer PDF
1. `client/session.go` — `Login()` (GET `/auth/login` → extract CSRF → POST `/login`), `CreateAnswerPDF()`, `DownloadAnswerPDF()`, `IsAnswerPDFReady()`
2. `client/pdf.go` — `GetSurveyPDF()` (Basic Auth), answer PDF orchestration
3. `cmd/pdf.go` — `RunPDF(cfg, args)` for `survey`/`answer` actions
4. Verify: `eusurveymgr pdf answer -code ae8d5fec-daaf-4aba-b860-544d1f717d8a -o ./`

### Phase 5: Database Access
1. `db/db.go` — `ConnectToMySQL()` (copy pattern from riasec)
2. `db/surveys.go` — `ListSurveys()`, `db/answers.go` — `ListAnswerSets()`, `LookupUniqueCode()`
3. `cmd/db.go` — `RunDB(cfg, args)` for `surveys`/`answers`/`lookup`
4. Wire `pdf answer -email` to do DB lookup → then trigger PDF
5. Verify: `eusurveymgr pdf answer -email user@example.com -survey 4609`

### Phase 6: Tokens + Polish
1. `client/tokens.go`, `cmd/tokens.go`
2. Add `version` subcommand (ldflags in Makefile)
3. Add `--json` output option for machine-readable output

## Reference Files to Reuse
- `riasec.git/log/log.go` → copy verbatim as `log/log.go`
- `riasec.git/config/config.go` → adapt pattern for `config/config.go`
- `riasec.git/db/db.go:14-26` → reuse `ConnectToMySQL` function
- `riasec.git/Makefile` → adapt for eusurveymgr binary name
- `riasec.git/bin/generate-pdf.sh` → reference implementation for PDF flow (CSRF, login, trigger, download)

## Verification
After each phase, run against the live instance:
```bash
go build -o bin/eusurveymgr && ./bin/eusurveymgr -config bin/eusurveymgr.json surveys list
```
End-to-end test: generate a PDF for a known respondent email + survey ID and verify the downloaded file.