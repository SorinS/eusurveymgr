# EUSurvey API Reference

## Instance Details

- **URL**: https://eusurvey.escoaladevalori.ro/eusurvey
- **Platform**: EUSurvey (Java/Spring/Tomcat 8.0.15, Java 8u281)
- **Source**: https://github.com/EUSurvey/EUSURVEY

## Authentication

### HTTP Basic Auth (WebService API)
Used for all `/webservice/*` endpoints.
- **User**: root
- **Password**: Jc9CA76b3Y5TQ7fS

```bash
curl -u root:Jc9CA76b3Y5TQ7fS "$URL/webservice/getSurveys"
```

### Form-based Session Auth (Web UI + PDF endpoints)
Used for `/worker/*`, `/pdf/*`, and other web endpoints. Requires CSRF token.

```bash
# 1. Get CSRF token from login page
CSRF=$(curl -s -c cookies.txt "$URL/auth/login" \
  | grep '_csrf.*content=' | head -1 \
  | sed 's/.*content="\([^"]*\)".*/\1/')

# 2. Login (returns 302 on success)
curl -s -b cookies.txt -c cookies.txt -o /dev/null -w "%{http_code}" \
  "$URL/login" \
  -d "username=root&password=Jc9CA76b3Y5TQ7fS&_csrf=$CSRF"

# 3. Use session cookies for subsequent requests
curl -s -b cookies.txt "$URL/worker/createanswerpdf/{UNIQUECODE}"
```

**Key fields**: `username`, `password`, `_csrf` (POST to `/login`)
**CSRF source**: `<meta name="_csrf" content="TOKEN"/>` in the HTML of `/auth/login`
**Success indicator**: HTTP 302 redirect

## WebService API Endpoints (HTTP Basic Auth)

All under `/webservice/*`. Require HTTP Basic Auth.

### Surveys

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webservice/getSurveys` | List all surveys (supports query filters) |
| GET | `/webservice/getSurveyMetadata/{alias}` | Get survey metadata by alias/shortname |
| GET | `/webservice/getSurveyPDF/{alias}` | Download survey form as PDF |

### Results Export

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webservice/prepareResults/{formid}/{showids}` | Start async results export. Returns task ID |
| GET | `/webservice/getResults/{taskid}` | Download completed results export (XML/CSV) |
| GET | `/webservice/prepareResultsPDF/{formid}` | Start async PDF results export. Returns task ID |
| GET | `/webservice/getResultsPDF/{taskid}` | Download completed PDF results export |

**Note**: `prepareResults` and `prepareResultsPDF` are async. Call them, get a task ID, then poll `getResults`/`getResultsPDF` until the export is ready.

### Token Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webservice/getTokens/{surveyname}` | List invitation tokens for a survey |
| GET | `/webservice/createToken/{surveyname}` | Create a new invitation token |
| GET | `/webservice/deleteToken/{surveyname}/{token}` | Delete a token |
| GET | `/webservice/updateToken/{surveyname}/{token}` | Update a token |

## Web Endpoints (Session Auth)

### Answer PDF Generation (3-step flow)

```
Step 1: Login (see above)
Step 2: GET /worker/createanswerpdf/{UNIQUECODE}  → returns "OK"
Step 3: GET /pdf/answer/{UNIQUECODE}               → returns PDF file
```

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/worker/createanswerpdf/{uniquecode}` | **Trigger** answer PDF generation (returns "OK") |
| GET | `/pdf/answer/{uniquecode}` | **Download** generated answer PDF |
| GET | `/pdf/answerready/{uniquecode}` | **Check** if PDF exists ("OK" or "wait") |
| GET | `/preparecontribution/{uniquecode}` | Get answer HTML (the rendered survey response) |

**Important distinctions**:
- `/pdf/answer/` only SERVES existing PDFs — does NOT generate them
- `/pdf/answerready/` only CHECKS existence — does NOT generate them
- `/worker/createanswerpdf/` is what actually TRIGGERS generation

### Other Useful Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pdf/survey/{id}` | Download survey form PDF (by numeric ID) |
| GET | `/management/list` | List surveys in management console |
| GET | `/forms/{shortname}` | Access survey by shortname |

## PDF File Storage

Generated PDFs are stored on the server at:
```
/home/eusurvey/eusurveytemp/surveys/{first-char-of-uid}/{survey-uid}/EXPORTS/answer{uniquecode}.pdf
```

Example:
```
/home/eusurvey/eusurveytemp/surveys/f/f1234.../EXPORTS/answerae8d5fec-daaf-4aba-b860-544d1f717d8a.pdf
```

## UNIQUECODE

The `UNIQUECODE` is a UUID stored in `ANSWERS_SET.UNIQUECODE` that uniquely identifies a survey response. It's needed for all answer PDF operations.

To find a UNIQUECODE by email and survey ID:
```sql
SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE
FROM ANSWERS_SET a_set
JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID
WHERE a_set.SURVEY_ID = {SURVEY_ID}
  AND a.PA_ID = 0
  AND a.VALUE = '{EMAIL}'
ORDER BY a_set.ANSWER_SET_DATE DESC
LIMIT 1;
```

## JAXB Fix (Java 8)

The EUSurvey PDF generation uses Flying Saucer + iText which depends on JAXB. On this instance (Java 8u281), the webapp had `jaxb-core-2.2.11.jar` and `jaxb-impl-2.2.11.jar` but was missing `jaxb-api-2.2.11.jar`, causing `NoClassDefFoundError: javax/xml/bind/JAXBException`.

**Fix applied**: Added `jaxb-api-2.2.11.jar` to `WEB-INF/lib` (must match the 2.2.11 version, NOT 2.3.x which requires Java 9+).