# SpaceEdu: Non-Invasive Security Review

---

## Objective

Perform a non-invasive review of a demo web application and document its observable security headers, HTTPS behaviour, information exposure, and usability/security concerns, with evidence, severity ratings and actionable recommendations. No scanning, exploitation or login bypass is allowed.

No third-party target was supplied, so I built a local demo application (SpaceEdu, a space-science course platform) and reviewed it. The app was built with an AI coding assistant (Antigravity) and is **deliberately left unhardened**. It is intended for local use only and must not be deployed.

## Scope and authorisation

- **Target:** my own application, running on my own machine. No third-party system is involved.
- **In scope:** viewing pages, reading response headers (browser DevTools, single `curl` requests), inspecting the TLS certificate, browser storage and cookies, reading page source, normal use of signup/login/checkout with a test card, replaying my own unmodified requests in Burp Suite Repeater.
- **Out of scope:** port scans, directory brute-forcing, fuzzing, exploit attempts, authentication bypass, brute-force or rate-limit testing, injection payloads, load generation, automated scanners (Burp Scanner and Intruder were not used), and any non-local host.


## Approach

1. Wrote the scope note before observing anything.
2. Collected HTTP evidence with single requests to the redirect, HTTPS site, API, direct backend, a missing path and one exposed file path.
3. Inspected the TLS certificate and the negotiated connection.
4. After a normal login and checkout, inspected browser storage and my own request/response traffic.
5. Read the source to explain or confirm observations. These findings are labelled **source-assisted** in the report.
6. Rated each finding by likelihood and impact (OWASP Risk Rating), assuming a public deployment.

## Findings summary

| ID | Finding | Severity |
|---|---|---|
| F-01 | Project root is web-served; `/backend/.env`, source and compose file reachable | Critical |
| F-02 | No security headers (HSTS, CSP, nosniff, frame protection, Referrer-Policy, Permissions-Policy) | Medium |
| F-03 | Session token and profile stored in localStorage; no cookies | Medium |
| F-04 | Card number and CVV sent to the app backend; order completed with no payment processor | High (real-world) |
| F-05 | Open redirect via `login.html?redirect=` (source-assisted) | Medium |
| F-06 | Weak authentication controls (source-assisted) | Medium |
| F-07 | JWT handling: HS256, 24 h lifetime, no revocation, hardcoded fallback secret | Medium |
| F-08 | Plain-HTTP backend on :8080 reachable directly | Low |
| F-09 | Misleading "256-bit SSL" trust claim on checkout | Low |
| F-10 | Silent fallback to hardcoded course data can show or enrol the wrong course | Medium |
| F-11 | PostgreSQL port published to the host with default credentials (source-assisted) | Low |
| F-12 | Informational: Server banner, default 404, third-party assets without SRI, no `security.txt` | Info |

Details, evidence references and remediation steps are in the [technical report](report/Technical-Report.pdf).

## Repository layout

```
.
├── README.md
├── report/                  Technical report (PDF)
├── docs/                    Scope and authorisation note
├── target-app/              The SpaceEdu demo application
│   ├── backend/             Go API (auth, courses, checkout)
│   ├── *.html               Frontend pages
│   ├── docker-compose.yml   PostgreSQL
│   └── Caddyfile            Local HTTPS reverse proxy (no security headers, by design)
└── evidence/                curl outputs and redacted screenshots
```

## Running the target locally (optional)

Prerequisites: Docker, Go, Caddy.

```bash
cd target-app

# 1. Database
docker compose up -d

# 2. Backend config: copy and fill in values (never commit the real .env)
cp backend/.env.example backend/.env

# 3. Backend
cd backend && go run .

# 4. HTTPS proxy (second terminal, from target-app/)
caddy trust
caddy run --config Caddyfile
```

- HTTPS site: `https://localhost:8443`
- HTTP redirect: `http://localhost:8081`
- Direct backend (bypasses the proxy): `http://localhost:8080`

Use a test card such as `4242 4242 4242 4242` for checkout. No real payment is processed.

Reproduce the main observations:

```bash
curl -sI  http://localhost:8081/index.html
curl -skI https://localhost:8443/
curl -skI https://localhost:8443/api/courses
curl -sI  http://localhost:8080/index.html
curl -skI https://localhost:8443/backend/.env    # status and headers only
```

## Limitations

- The target is self-built, so results demonstrate method rather than a real engagement.
- This is a passive review, not a penetration test. Absence of a finding is not proof of security.
- Authentication throttling, injection, cross-user access control and session attacks were deliberately not tested.
- Some findings rely on source reading and are labelled as such.
- Remediation was **not** applied or re-tested, so no before/after results are claimed.
- A course-detail lookup problem observed during testing was not root-caused (see F-10 in the report).

## Safety and data handling

- No real credentials, tokens or personal data are committed. `backend/.env` is git-ignored; `backend/.env.example` holds placeholders.
- Screenshots in `evidence/` are redacted.
- All testing was done on my own machine against my own application.


## References

OWASP Secure Headers Project, OWASP Cheat Sheets (HTTP Headers, HTML5 Security, Session Management, Authentication, Unvalidated Redirects), RFC 6797 (HSTS), RFC 9116 (security.txt), NIST SP 800-63B, PCI DSS, OWASP Risk Rating Methodology. Full list in the technical report.