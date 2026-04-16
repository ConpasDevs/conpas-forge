# PRD: conpas-forge

> **One binary. conpas AI ecosystem — configured and ready.**

| Version | Date | Author | Status |
|---------|------|--------|--------|
| 0.1.0 | 2026-04-16 | conpas team | Draft |

> **Note**: This document is a fork/adaptation of `docs/PRD.md` (Gentleman AI Installer). Where they diverge, this document takes precedence for conpas-forge.

---

## 1. Problem Statement

The conpas team uses Claude Code and OpenCode with a specific, opinionated setup: SDD workflow, Engram persistent memory, Zoho Deluge coding standards, a team persona, and hardcoded security settings. Getting a new developer (or a new machine) to the same state requires copying files, registering MCP servers, and tracking down the right skill versions — a fragile, manual, error-prone process.

**conpas-forge eliminates that gap.** Run the binary, pick your modules from a TUI, and the full conpas AI stack is configured and ready in under a minute.

---

## 2. Vision

**An opinionated, team-internal CLI that installs the conpas AI ecosystem into Claude Code and OpenCode in one run.**

This is not a general-purpose installer for the public. It is a tool built for conpas developers that:

- Embeds all skills, configs, and assets directly in the binary
- Configures agents idempotently — safe to re-run at any time
- Tracks installed artifacts via a manifest so stale files are cleaned up automatically
- Reports its own version and surfaces update notifications at session start

---

## 3. Target Users

**Primary**: Developers on the conpas team onboarding to a new machine or setting up a fresh environment.

**Secondary**: Conpas team members upgrading to a new version of the skill set or switching agents.

This is a **team-internal tool**, not a public product. Users are technical — no hand-holding, no dependency detection, no guided tutorials.

---

## 4. Supported Platforms

| Platform | Priority |
|----------|----------|
| macOS (Apple Silicon) | P0 |
| macOS (Intel) | P0 |
| Linux (Ubuntu/Debian) | P0 |
| WSL2 (Windows) | P0 |
| Windows (native) | P1 |

---

## 5. Architecture (High-Level)

```
conpas-forge (Go binary)
├── cmd/
│   ├── install     — TUI module selector → runs selected installers
│   └── check       — version check (--json flag for machine output)
├── internal/
│   ├── installers/
│   │   ├── GentleAIInstaller   — SDD skills + CLAUDE.md + output styles + shared assets
│   │   ├── ConpasAIInstaller   — Zoho Deluge skill
│   │   ├── EngramInstaller     — downloads Engram binary + registers MCP server
│   │   └── ClaudeCodeInstaller — writes bypassPermissions Claude Code settings
│   ├── manifest/   — .forge-manifest.json tracking + stale file cleanup
│   └── config/     — ~/.config/conpas-forge/config.yaml
└── assets/         — ALL skills, CLAUDE.md, config templates EMBEDDED in binary
```

### Key architectural decisions

- **Skills are embedded** in the binary via Go `embed`. No git clones, no external downloads (except Engram binary). This makes installs fast, reproducible, and fully offline-capable.
- **Idempotent by design**: every installer checks current state before writing. Re-running is safe and produces the same result.
- **Atomic writes with backup**: files are written atomically; existing files are backed up before overwrite.
- **Manifest-driven cleanup**: `.forge-manifest.json` in the config dir tracks every installed file. On re-run, obsolete files from a prior version are deleted.
- **Config is opinionated**: no user-customizable config knobs. The tool installs the conpas standard. Deviations belong in the repo, not in user config.

---

## 6. Current Modules

### 6.1 Gentle AI Skills
**Installer**: `GentleAIInstaller`

Installs 21 SDD skills + CLAUDE.md persona + output styles + shared assets into `~/.config/opencode/skills/` (and equivalent paths for Claude Code).

Skills include the full SDD pipeline: `sdd-orchestrator`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-qa`, `sdd-archive`, `sdd-init`, `sdd-onboard`, `sdd-clarify`, `branch-pr`, `issue-creation`, `judgment-day`, `go-testing`, `skill-creator`, `skill-registry`, `sdd-update-checker`, and `zoho-deluge` (via ConpasAIInstaller).

Also writes `CLAUDE.md` to `~/.config/opencode/` with the team persona.

### 6.2 Zoho Deluge
**Installer**: `ConpasAIInstaller`

Installs the `zoho-deluge` skill — mandatory coding standard for all Zoho Deluge / Creator / CRM automation work. Focused on Extreme Statement Optimization, security, and maintainability.

### 6.3 Engram
**Installer**: `EngramInstaller`

Downloads the Engram binary for the detected platform/arch and registers it as an MCP server in the agent's config. Engram provides persistent cross-session memory so the AI agent remembers decisions, bugs, and conventions across sessions.

> **Note**: Engram auto-start on tool invocation is on the backlog — currently requires manual start.

### 6.4 Claude Code Settings
**Installer**: `ClaudeCodeInstaller`

Writes Claude Code settings with `bypassPermissions: true`. This is the team-standard security posture — hardcoded, not user-configurable.

---

## 7. `conpas-forge check`

```
conpas-forge check           # human-readable version status
conpas-forge check --json    # machine-readable JSON for CI / update-checker skill
```

Reports: installed version, latest available version, whether an update is available.

The `update-checker` skill (installed as part of module 6.1) calls this at session start and notifies the developer if a new version of conpas-forge is available.

---

## 8. Backlog (Prioritized)

| Priority | Feature | Notes |
|----------|---------|-------|
| 🔴 Highest | **Multi-agent support** | Support OpenCode, Cursor, etc. beyond current state. Absolute priority. |
| 🟠 High | **Agent interface abstraction** | Prerequisite for multi-agent — unified interface per agent type |
| 🟠 High | **Health verification** (`check --health`) | Verify Engram is running, skills deployed, MCP registered correctly |
| 🟠 High | **Engram auto-start** | Start Engram daemon if not running; depends on health check |
| 🟡 Medium | **Config backup & restore** | Snapshot + restore full conpas-forge config state |
| 🟡 Medium | **Config sync across machines** | Sync installed state via shared config; tied to multi-agent |
| 🟡 Medium | **OS/arch detection improvements** | Better detection for multi-agent and Engram binary selection |
| 🟢 Low | **GGA / auto code reviewer** | Needs dedicated design spike before implementation |
| 🟢 Low | **Theme/statusline customization** | Warp theme, Starship prompt, terminal styling |
| ⚪ Super low | **`curl` one-liner installer** | `curl -sL ... \| sh` bootstrap |
| ⚪ Super low | **Homebrew tap** | `brew install conpas/tap/conpas-forge` |

---

## 9. Explicitly Out of Scope

These items will NOT be built. Raising them again should come with a strong justification.

| Item | Reason |
|------|--------|
| Dependency detection / installation | Users are technical; no external deps except the binary |
| Non-interactive / headless CLI install mode | Not needed; TUI is fine for the team |
| Repair command | Idempotent reinstall covers this entirely |
| Review/confirm screen before install | Users are technical; not needed |
| Team profiles / preset system | Too narrow; TUI selection is sufficient |
| MCP selection UI | Context7 is hardcoded; no need for selection |
| Skills cloned from external repos | Embedded in binary — deliberate, non-negotiable decision |
| Granular self-update | Re-running the new binary is sufficient |
| Security permissions UI | `bypassPermissions` is hardcoded team standard |

---

## 10. Success Metrics

| Metric | Target |
|--------|--------|
| Time from zero to full conpas setup | < 2 minutes |
| Re-run safety (idempotency) | Zero data loss, zero unexpected side effects |
| Skills deployed correctly | 100% — verifiable via `check --health` (backlog) |
| Onboarding friction for new team members | Zero — binary + one command |
| Update adoption | Surfaced automatically via `update-checker` skill at session start |

---

## 11. File Layout (Installed Artifacts)

```
~/.config/opencode/
├── skills/
│   ├── sdd-orchestrator/SKILL.md
│   ├── sdd-explore/SKILL.md
│   ├── ... (21 skills total)
│   └── zoho-deluge/SKILL.md
├── CLAUDE.md
└── output-styles/

~/.config/conpas-forge/
├── config.yaml
└── .forge-manifest.json

~/.local/bin/engram   (or OS equivalent)
```

---

## 12. Non-Goals (Philosophy)

conpas-forge is a **tool for the team, by the team**. It is not:

- A general-purpose AI ecosystem installer for the public
- A plugin marketplace
- A configuration management system
- A replacement for individual agent documentation

When in doubt, prefer **less complexity** over more features. The goal is a fast, reliable, boring tool that does exactly what it says and nothing else.
