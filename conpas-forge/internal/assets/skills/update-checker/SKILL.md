# Skill: update-checker

## Purpose

Automatically check for available updates to conpas-forge ecosystem modules at the start of every Claude Code session.

## Trigger

This skill activates **silently at the start of every session**. No user prompt needed.

## Instructions

At the start of every session, run the following command **silently** (do not announce you are doing it):

```bash
conpas-forge check --json
```

Parse the JSON output, which is an object with a `modules` array. Each element has the following fields:
- `name` — module name
- `installedVersion` — currently installed version
- `latestVersion` — latest available version
- `status` — one of: `up-to-date`, `outdated`, `not-installed`, `unknown`
- `downloadUrl` — URL to download the latest release

## Rules

- If `conpas-forge` is not found in PATH or the command fails, **say nothing** — do not mention the failure.
- If all modules have status `up-to-date` or `unknown` or `not-installed`, **say nothing** — zero noise.
- Only speak if at least one module has status `outdated`.

## When Updates Are Available

If any module has status `outdated`, display the following **once** at the start of the session:

---

**🔔 conpas-forge updates available**

| Module | Current | Available |
|--------|---------|-----------|
| {name} | {installedVersion} | {latestVersion} |

Download the latest release: https://github.com/conpas-ai/conpas-forge/releases/latest

Then run `conpas-forge install` to apply the updates.

---

Only include modules with status `outdated` in the table. Do not list up-to-date modules.

After displaying this message, proceed normally with the session. Do not repeat this message during the same session.
