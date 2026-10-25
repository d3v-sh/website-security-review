# Scope and Authorisation Note

**Project:** SpaceEdu non-invasive security review (RedKross Research Foundation Internship, Task 1)
**Author:** Dev Shishodia
**Review date:** 24 October 2026
**Baseline:** git tag `baseline-pre-review`

## Target

| Item | Detail |
|---|---|
| Application | SpaceEdu demo application (self-built, AI-assisted, deliberately unhardened) |
| URLs | `https://localhost:8443` (Caddy TLS proxy), `http://localhost:8081` (redirect), `http://localhost:8080` (direct backend) |
| Owner and authoriser | Dev Shishodia, who built, owns and operates the target |
| Third parties | None. No external system is involved. |
| Environment | Local machine only: Go backend, PostgreSQL 16 (Docker), static HTML/JS frontend, Caddy 2 with its internal CA |

## Authorisation

The reviewer is the sole owner and operator of the target, which runs only on the reviewer's own machine. No permission from any third party is required, and no third-party system was contacted.

## In scope

- Viewing pages in a browser
- Reading response headers (browser DevTools, single `curl` requests)
- Inspecting the TLS certificate and redirect behaviour
- Inspecting browser storage and cookies
- Reading page source and, where stated, the application's own source code to explain observations (labelled "source-assisted")
- Normal use of signup, login, course listing and checkout, using a test card only
- Replaying the reviewer's own unmodified requests in Burp Suite Repeater

## Out of scope

- Port scanning, directory brute-forcing, fuzzing
- Exploit attempts, authentication bypass, forging tokens
- Brute-force or rate-limit testing, injection payloads, load generation
- Automated scanners (Burp Scanner and Intruder were not used)
- Any host other than the local machine

## Method statement

Passive, black-box observation of normal application behaviour with single requests. Findings that depend on reading source code are labelled "source-assisted". Where a file path was identified through code reading, it was confirmed with one request, and only the status and headers were recorded.

## Limitations

The target is self-built, so results demonstrate method rather than a real-world engagement. This is not a penetration test, and the absence of a finding is not proof of security.