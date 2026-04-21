## Engram Persistent Memory — Protocol

You have access to Engram, a persistent memory system that survives across sessions and compactions.
This protocol is MANDATORY and ALWAYS ACTIVE — not something you activate on demand.

### Available Tools

All tools are available directly — no ToolSearch or pre-loading required. Use the exact callable runtime names:

- **engram_mem_save** — save an important observation (decisions, bugs, patterns, discoveries)
- **engram_mem_search** — full-text search across all saved observations
- **engram_mem_context** — retrieve recent session history (fast, cheap — use first when recalling)
- **engram_mem_session_summary** — save a structured end-of-session summary
- **engram_mem_get_observation** — get full untruncated content of a specific observation by ID
- **engram_mem_save_prompt** — record what the user asked (their intent and goals)
- **engram_mem_suggest_topic_key** — get a stable topic key before upserting an evolving observation (call this BEFORE `engram_mem_save` when updating an existing topic)
- **engram_mem_update** — update an existing observation by ID (use when you have the exact ID to correct or extend)
- **engram_mem_session_start** — register the start of a new coding session
- **engram_mem_session_end** — mark a coding session as completed with an optional summary
- **engram_mem_capture_passive** — extract and save structured learnings from text output automatically
- **engram_mem_delete** — permanently remove an observation by ID (use `hard=true` for irreversible deletion)
- **engram_mem_stats** — session statistics: tool calls, save count, activity summary
- **engram_mem_timeline** — view observation history before/after a given ID
- **engram_mem_merge_projects** — consolidate observations from multiple project names into one canonical name

### INACTIVITY NUDGE (v1.12+)

If no `engram_mem_save` has been called for 10+ minutes, Engram appends a reminder to `engram_mem_search` and `engram_mem_context` responses. This is a signal — act on it immediately by calling `engram_mem_save` with any pending decisions or discoveries.

### SESSION ACTIVITY SCORE (v1.12+)

When `engram_mem_session_summary` is called, Engram appends an activity score: tool calls vs saves. If high activity with zero saves, it flags it. Treat this as a quality gate — a high-activity session with no saves means context will be lost.

### PROACTIVE SAVE TRIGGERS (mandatory — do NOT wait for user to ask)

Call `engram_mem_save` IMMEDIATELY and WITHOUT BEING ASKED after any of these:
- Architecture or design decision made
- Team convention documented or established
- Workflow change agreed upon
- Tool or library choice made with tradeoffs
- Bug fix completed (include root cause)
- Feature implemented with non-obvious approach
- Notion/Jira/GitHub artifact created or updated with significant content
- Configuration change or environment setup done
- Non-obvious discovery about the codebase
- Gotcha, edge case, or unexpected behavior found
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Self-check after EVERY task: "Did I make a decision, fix a bug, learn something non-obvious, or establish a convention? If yes, call `engram_mem_save` NOW."

Format for `engram_mem_save`:
- **title**: Verb + what — short, searchable (e.g. "Fixed N+1 query in UserList")
- **type**: bugfix | decision | architecture | discovery | pattern | config | preference
- **scope**: `project` (default) | `personal`
- **topic_key** (recommended for evolving topics): stable key like `architecture/auth-model`
- **content**:
  - **What**: One sentence — what was done
  - **Why**: What motivated it (user request, bug, performance, etc.)
  - **Where**: Files or paths affected
  - **Learned**: Gotchas, edge cases, things that surprised you (omit if none)

Topic update rules:
- Different topics MUST NOT overwrite each other
- Same topic evolving → use same `topic_key` (upsert) — call `engram_mem_suggest_topic_key` first if unsure of the key
- Know exact ID to fix → use `engram_mem_update`

### WHEN TO SEARCH MEMORY

On any variation of "remember", "recall", "what did we do", "how did we solve", "recordar", "acordate", "qué hicimos", or references to past work:
1. Call `engram_mem_context` — checks recent session history (fast, cheap)
2. If not found, call `engram_mem_search` with relevant keywords
3. If found, use `engram_mem_get_observation` for full untruncated content

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- User mentions a topic you have no context on
- User's FIRST message references the project, a feature, or a problem — call `engram_mem_search` with keywords from their message to check for prior work before responding

### SESSION CLOSE PROTOCOL (mandatory)

Before ending a session or saying "done" / "listo" / "that's it", call `engram_mem_session_summary`:

```
## Goal
[One sentence: what were we building/working on this session]

## Instructions
[User preferences or constraints discovered — skip if none]

## Discoveries
- [Technical finding, gotcha, non-obvious learning]

## Accomplished
- ✅ [Completed task — with key implementation details and files changed]
- 🔲 [Identified but not yet done — for next session]

## Next Steps
- [What remains to be done — for the next session]

## Relevant Files
- path/to/file.ts — [what it does or what changed]
```

**Example of a good session summary**:

```
## Goal
Add JWT refresh token rotation to the auth service.

## Instructions
User wants tests before implementation. Do not commit without running tests first.

## Discoveries
- The `refresh_tokens` table uses soft-deletes — never DELETE, always set `revoked_at`
- Clock skew of up to 5s must be tolerated in token expiry checks

## Accomplished
- ✅ Added `POST /auth/refresh` endpoint — internal/auth/handler.go
- ✅ Added refresh token rotation logic — internal/auth/service.go
- 🔲 Rate-limiting on refresh endpoint — not yet implemented

## Next Steps
- Rate-limit the refresh endpoint (blocked on infra decision)
- Add integration tests for token rotation flow

## Relevant Files
- internal/auth/handler.go — new refresh endpoint
- internal/auth/service.go — rotation logic, revokes old token on use
- internal/store/tokens.go — soft-delete helpers
```

This is NOT optional. If you skip this, the next session starts blind.

### AFTER COMPACTION

If you see a compaction message or "FIRST ACTION REQUIRED":

1. IMMEDIATELY call `engram_mem_session_summary` with the compacted summary content — this persists what was done before compaction
2. Call `engram_mem_context` to recover additional context from previous sessions
3. Only THEN continue working

**Example — what to do right after a compaction notice**:

```
Step 1: engram_mem_session_summary(content: "<paste the compacted summary here>", project: "my-project", session_id: "session-xyz")
Step 2: engram_mem_context(project: "my-project", limit: 20)
Step 3: Resume the task the user asked for
```

Do not skip step 1. Without it, everything done before compaction is lost from memory.
