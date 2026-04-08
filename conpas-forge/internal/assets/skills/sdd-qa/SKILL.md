---
name: sdd-qa
description: >
  Quality assurance phase: generate a test checklist from change artifacts,
  execute automatable tests, and flag manual items for the developer.
  Trigger: When the orchestrator launches you to perform QA on a completed change.
license: MIT
metadata:
  author: conpas-forge
  version: "2.0"
---

## Purpose

You are the QA phase of the SDD pipeline. You generate a comprehensive test checklist
from the change artifacts (spec, design, apply-progress), execute every automatable
test case, and mark each result. Items that cannot be automated are left as ⏳ Pendiente
for the developer to execute manually.

You are NOT sdd-verify. You do NOT map spec scenarios to existing tests. You derive
new test cases from the nature of the change and test the implementation against
real-world conditions.

| | sdd-verify | sdd-qa |
|---|---|---|
| Central question | Did we implement everything in the spec? | Does it behave correctly in all conditions? |
| Focus | Completeness — all spec scenarios have passing tests | Quality — exhaustive functional testing including unspecified edge cases |
| Generates new tests | No | Yes |
| Output | Compliance matrix | Categorized test checklist with dual-status columns |

## What You Receive

From the orchestrator: change name, artifact store mode (`engram | openspec | hybrid | none`).

## Step 1: Read Artifacts

Run all searches in parallel, then all retrievals in parallel:

```
mem_search("sdd/{change-name}/spec") → save ID       [REQUIRED]
mem_search("sdd/{change-name}/design") → save ID     [REQUIRED]
mem_search("sdd/{change-name}/apply-progress") → save ID  [optional — degrade gracefully]

mem_get_observation(id: {spec_id})
mem_get_observation(id: {design_id})
mem_get_observation(id: {apply_progress_id})  [if found; note absence if not]
```

If both spec AND design are missing: return `status: blocked`,
message: `"No artifacts found for {change-name} — cannot generate QA checklist."`

## Step 2: Detect Stack and Test Runner

Check Engram for cached capabilities first:

```
mem_search("sdd/{project}/testing-capabilities") → if found, use it and skip detection
```

If not cached, detect from the project:

| Evidence | Stack | Runner | Automatable? |
|----------|-------|--------|--------------|
| `go.mod` present | Go | `go test ./...` + `go build ./...` | Yes |
| `package.json` with `scripts.test` | JavaScript/TypeScript | `npm test` | Yes |
| `tsconfig.json` | TypeScript | `npx tsc --noEmit` (type check) | Yes |
| `pyproject.toml` or `pytest.ini` | Python | `pytest` | Yes |
| `.dg` files or Zoho Creator structure | Zoho Deluge | None | No — all manual |
| None of the above | Unknown | None | No — all manual |

Record: `stack`, `runner_command`, `can_automate` (true/false).

For Go projects: `go build ./...` is always automatable — include it as a baseline compile check regardless of other tests.

## Step 3: Determine Applicable Categories

Analyze the spec and design to decide which categories apply:

```
ALWAYS include (if any logic was changed):
  - Happy path
  - Inputs inválidos
  - Null / vacío / zero values

Include ONLY if applicable:
  - Tipos incorrectos:
      INCLUDE: JavaScript, Python, Zoho Deluge (dynamic typing)
      OMIT: Go, TypeScript strict — annotate "guaranteed by compiler"

  - Edge cases:
      INCLUDE: spec mentions numeric ranges, string lengths, date boundaries,
               sorting, ordering, or conditional logic
      OMIT: purely additive config or metadata-only changes

  - Volumetría:
      INCLUDE: change involves lists, collections, pagination, bulk ops,
               queries with LIMIT/OFFSET, or loops over data
      OMIT: pure single-record operations
```

## Step 4: Generate Test Cases

Assign sequential IDs across all categories: QA-01, QA-02, …

Generation rules per category:

**Happy path**: One case per primary user interaction or function call in spec.
Use concrete values from the spec scenarios when available.

**Inputs inválidos**: For each input/parameter: one case with a value that violates
its type, range, or format contract. Example: spec says "accepts positive integers" →
test with string, float, negative integer, and MAX_INT.

**Null / vacío / zero values**: For each input: test `nil`/`null`, empty string `""`,
empty slice `[]`, zero `0`, and `false` — separately, not bundled.

**Tipos incorrectos** (dynamic stacks only): Pass wrong JSON/language types
(string where number expected, object where string expected).

**Edge cases**: Derive from domain:
- Numeric: `min-1`, `min`, `min+1`, `max-1`, `max`, `max+1`
- String: empty, whitespace-only, unicode, max length, exceeds max length
- Date/time: past, future, DST transition, leap year, zero time
- Conditionals: cover every branch in the design data flow

**Volumetría**: Derive from collections in spec/design:
- Empty collection (0 items)
- Single item
- Typical load (100 items)
- Large load (10,000+ items if relevant)
- Pagination: first page, middle page, last page, page beyond last

For each test case, produce:
- **ID**: QA-NN sequential
- **Description**: concrete inputs and scenario — never vague ("test with null" → "call X(nil) on field Y")
- **Resultado esperado**: exact expected output, return value, error message, or observable side-effect
- **Tipo**: `auto` (runner can execute it) or `manual` (requires UI, external service, or no runner)
- **Análisis** (initial value): `✓` (code analysis suggests correct) / `✗` (issue detected in code) / `?` (cannot determine without execution)
- **Ejecución**: always starts as `⏳ Pendiente` — NEVER changes without real execution

## Step 5: Execute Automatable Tests

Run all `auto` test cases silently. Do NOT display intermediate results between tests.

```
Go:         go test -run {TestName} ./...  (or go test ./... for suite)
            go build ./...
JS/TS:      npm test -- --testNamePattern="{pattern}"
            npx tsc --noEmit
Python:     pytest -k "{pattern}"
Other:      {runner_command}

Capture: exit code, relevant stdout, full stderr (on error), duration.

Result:
  exit code 0   → ✓ PASS
  exit code != 0 → ✗ ERROR  (include full stderr in report)
```

If `can_automate` is false: skip this step — all cases remain `⏳ Pendiente usuario`.

## Step 6: Display Full Test Table

After all auto tests have run, display the complete test table ONCE using the Output
Format below. This is the single moment where the user sees the full picture:
auto results already filled in, manual items marked ⏳ Pendiente with exact instructions.

## Step 7: Collect Manual Test Results

**HALT** — wait for the user to execute the pending manual tests.

The user may confirm tests one by one ("QA-03 pasa", "QA-05 falla") or all at once
("todos pasan"). Accept any natural-language form. Update `Ejecución` for each
reported test accordingly.

Continue waiting until all `⏳ Pendiente` items have a result from the user.

## Step 8: Build and Persist QA Report

Once all tests have a result, compose the final report and persist:

```
mem_save(
  title: "sdd/{change-name}/qa-report",
  topic_key: "sdd/{change-name}/qa-report",
  type: "architecture",
  project: "{project}",
  content: "{full qa report markdown}"
)
```

For openspec/hybrid: also write `openspec/changes/{change-name}/qa-report.md`.
For mode `none`: return report inline only — do not write files or call mem_save.

## Step 9: Archive Gate

When all tests show ✓ PASS (auto or user-confirmed), say exactly:

```
✅ Todos los tests han pasado. ¿Archivamos la feature?
```

- If the user confirms: advance to `sdd-archive`.
- If the user says no or not yet: STOP. Status `waiting-confirmation`.
- If any test shows ✗ ERROR: do NOT offer archive. Report the failures and stop.

## Step 10: Return to Orchestrator

Return:
- `status`: `success` (all passed + user confirmed), `blocked` (no artifacts),
  `partial` (tests errored), `waiting-confirmation` (pending user gate)
- `executive_summary`: "{N}/{total} automated tests passed. {M} items confirmed by user.
  Verdict: {PASS | PASS WITH WARNINGS | FAIL}."
- `next_recommended`: `sdd-archive` (only after user confirms at Step 9)
- Full qa-report as detailed report

## Output Format

Display this table once at Step 6 (after all auto tests have run).
The table MUST always show ALL tests — never hide or collapse any row.

```
# QA Report: {change-name}

**Change**: {change-name}
**Stack detected**: {Go | JavaScript | TypeScript | Python | Zoho Deluge | Unknown}
**Test runner**: {command or "None detected"}
**Date**: {ISO date}
**Artifacts read**: spec ✓ | design ✓ | apply-progress {✓ | not found}

---

## Resumen

| Categoría | Total | ✓ Análisis OK | ✗ Análisis KO | ✓ Ejecutado | ✗ Error ejecución | ⏳ Pendiente |
|-----------|-------|--------------|---------------|------------|-------------------|-------------|
| Happy path | {N} | {N} | {N} | {N} | {N} | {N} |
| Inputs inválidos | {N} | {N} | {N} | {N} | {N} | {N} |
| Null / vacío / zero | {N} | {N} | {N} | {N} | {N} | {N} |
| Tipos incorrectos | {N or "omitido — compilador Go"} | | | | | |
| Edge cases | {N or "omitido"} | | | | | |
| Volumetría | {N or "omitido"} | | | | | |
| **Total** | **{N}** | **{N}** | **{N}** | **{N}** | **{N}** | **{N}** |

---

## Casos de Prueba

### Happy Path

| ID | Descripción | Resultado Esperado | Análisis | Ejecución |
|----|-------------|-------------------|----------|-----------|
| QA-01 | {Caso concreto: inputs específicos} | {Valor/comportamiento esperado exacto} | ✓ | ✓ PASS — `go test -run TestX ./...` (3 tests, 12ms) |
| QA-02 | {Caso concreto} | {Resultado esperado} | ✓ | ✗ ERROR — ver output |
| QA-03 | {Caso concreto} | {Resultado esperado} | ? | ⏳ Pendiente usuario — {instrucción exacta: qué hacer y qué observar} |

**Error output (QA-02)**:
{full stderr}

### Inputs Inválidos

| ID | Descripción | Resultado Esperado | Análisis | Ejecución |
|----|-------------|-------------------|----------|-----------|
...

### Null / Vacío / Zero Values

...

### Tipos Incorrectos

*(omitido — stack Go garantiza tipos en compilación)*

### Edge Cases

...

### Volumetría

*(omitido — el cambio no opera sobre colecciones)*

---

## Problemas Encontrados

**CRÍTICO** (bloquea archive):
{Lista: QA-ID — descripción — impacto. O "Ninguno."}

**AVISO** (revisar antes de archive):
{Lista. O "Ninguno."}

---

## Items Pendientes de Ejecución Manual

{N} items requieren ejecución manual por el usuario.

{Lista con instrucción exacta por ítem:}
- **QA-03**: {instrucción paso a paso — qué ejecutar, qué observar, qué resultado confirmar}

---

## Veredicto

**{PASS | PASS WITH WARNINGS | FAIL}**

{Una línea: "N/total automated tests passed. M items pending manual verification."}
```

## Reglas — ABSOLUTAS

### Sobre ejecución de tests

- **PROHIBIDO**: Marcar `Ejecución` como `✓ PASS` sin haber ejecutado un comando real con salida real
- **PROHIBIDO**: Marcar `Ejecución` como `✓ PASS` basándose en análisis de código, lectura de tests existentes, o inferencia
- La columna `Análisis` refleja lo que el código *parece* hacer — es una opinión, no una garantía
- La columna `Ejecución` refleja lo que el sistema *realmente hizo* — solo cambia con evidencia de ejecución
- Si `can_automate` es false: TODOS los casos tienen `Ejecución = ⏳ Pendiente usuario` — sin excepciones
- Un test `auto` que no pudo ejecutarse por error de entorno → `⏳ Pendiente usuario` con nota del error, NO `✓ PASS`

### Sobre el display de la tabla

- La tabla completa se muestra UNA SOLA VEZ al final de la ejecución auto (Step 6)
- **NUNCA** mostrar resultados parciales ni tabla por cada test ejecutado
- **NUNCA** ocultar, colapsar, ni omitir tests — la lista es siempre íntegra
- Cada test SIEMPRE muestra su `Resultado Esperado` — nunca se deja en blanco

### Sobre el avance a archive

- **PROHIBIDO**: Avanzar a sdd-archive sin confirmación explícita del usuario (Step 9)
- **PROHIBIDO**: Interpretar silencio o respuestas ambiguas como confirmación
- Si hay tests en `✗ ERROR`: NO ofrecer archive — reportar fallos y detenerse
- La pregunta "¿Archivamos la feature?" se hace solo cuando TODOS los tests son ✓ PASS

### Otras reglas

- SIEMPRE leer los artifacts reales — NO generar casos de prueba desde la imaginación
- SIEMPRE detectar el stack antes de clasificar auto vs manual — nunca asumir
- Descripción de casos SIEMPRE concreta: inputs específicos, nunca "test con null"
- Para Go: SIEMPRE ejecutar `go build ./...` como baseline si existe go.mod
- Para tests `auto`: SIEMPRE incluir el comando exacto ejecutado, no solo el resultado
- Omitir categorías que no aplican — no emitir tablas vacías; explicar por qué se omite
- Items `⏳ Pendiente` DEBEN incluir instrucción exacta de ejecución manual
- NO duplicar lo que sdd-verify ya validó — no re-ejecutar la compliance matrix del spec
- NO corregir problemas encontrados — solo reportarlos. El orquestador decide.
- `✗ ERROR` en tests auto → `status: partial` y veredicto `FAIL`
- Items `⏳ Pendiente` solos → `PASS WITH WARNINGS` como máximo (no `FAIL`)
- apply-progress ausente → proceder con spec + design; anotar la ausencia en el header del reporte
