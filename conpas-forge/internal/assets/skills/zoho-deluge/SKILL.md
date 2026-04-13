---
name: zoho-deluge
description: >
  Mandatory coding standard for Zoho Deluge development. Focused on Extreme Statement
  Optimization, Security, Scalability, and Maintainability.
  Trigger: When working with Zoho Deluge scripts, .dg files, Creator functions, CRM workflows, or any Zoho automation code.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
allowed-tools: Read, Edit, Write, Glob, Grep, Bash
---

# Skill: Zoho Deluge Architect

## Trigger

Invoke this skill when the user:
- Asks to write, review, or refactor Zoho Deluge code
- Mentions `.dg` or `.ds` files, Creator functions, CRM workflows, or Zoho automation
- Uses phrases like "Deluge", "Zoho Creator", "zoho.crm", "invokeurl", "for each record"
- Requests a Zoho-specific function, module, or integration script

**MANDATORY**: Apply every standard in this skill before writing a single line of Deluge code.

## Description
Mandatory coding standard for Zoho Deluge development. Focused on **Extreme Statement Optimization**, **Security**, **Scalability**, and **Maintainability**.

## Philosophy
1.  **Business Logic > Code:** Understanding the requirement is paramount before typing a single line.
2.  **Statement Economy:** Every line counts. The Execution Limit is the number one enemy.
3.  **Security First:** Zero secrets in code. Always externalize configuration.
4.  **Clean Code:** Code is read more often than it is written. Write for humans.

## I. Optimization & Performance (Deluge Specific)

### 1. Loops & Data Structures
*   **FORBIDDEN:** Nested loops (`for each` inside `for each`). O(n²) complexity is unacceptable.
*   **SOLUTION:** Use Maps to index data and cross-reference information (O(n) complexity).
    *   *Pattern:* Iterate list A -> Create Map (Key: ID) -> Iterate List B -> Lookup in Map (`map.get(id)`).
*   **NATIVE FUNCTIONS:** Mandatory use of native collection functions to avoid manual loops: `.get()`, `.put()`, `.contain()`, `.intersect()`, `.distinct()`, `.sort()`.

### 2. Bulk/Batch Operations
*   **FORBIDDEN:** Executing `invokeurl`, `zoho.crm.createRecord`, `updateRecord`, `sendmail`, or any I/O operation INSIDE a loop.
*   **SOLUTION:**
    1.  Iterate and build a `List` of maps with the data to process.
    2.  Execute the operation in bulk outside the loop (`zoho.crm.bulkCreate`, etc.).
    3.  If the external API does not support bulk operations, group into small batches (chunks) of 50/100, but never 1-by-1 without control.

### 3. Initialization & Strings
*   **STATEMENT ECONOMY:**
    *   Use JSON notation or `putAll` for initialization: `myMap = {"key": "val", "k2": "v2"};`.
    *   *Forbidden:* `m = Map(); m.put("k","v");` (Wasted statements).
*   **STRING BUFFERS:** Do not concatenate strings inside loops (`txt = txt + "..."`).
    *   Use a List as a buffer: `bufferList.add("line")`.
    *   At the end, join the list: `result = bufferList.toString(SEPARATOR);`.
    *   *Separator:* Choose based on context (`\n`, `<br>`, `,`).

## II. Security & Configuration (Architecture)

### 4. Credential & Secret Management
*   **ZERO HARDCODING:** Strictly forbidden to write tokens, passwords, client_secrets, or API keys directly in the code.
*   **ZOHO CONNECTIONS:** Always use **Zoho Connections** for `invokeurl`.
    *   *Correct:* `resp = invokeurl [ ... connection: "connection_name" ... ];`
*   **SECURE HEADERS:** If an external API requires manual headers (e.g., Basic Auth), the token must come from an encrypted Organization Variable or the extension's secure storage, never a string literal.

### 5. Configurability (No Magic Numbers/Strings)
*   **PARAMETERIZATION:** API URLs, notification email addresses, thresholds (e.g., "max discount"), Status IDs, or Pipeline IDs MUST NOT be in the code.
*   **STORAGE:**
    *   **Organization Variables:** For simple global values (e.g., `API_BASE_URL`).
    *   **Configuration Module (Custom Module):** For complex mapping tables or configurations that a functional user needs to change without touching code.
*   **DYNAMIC IDS:** Never use literal IDs (`if id == "392000..."`). Search by `API Name` or load from configuration.

## III. Robustness & Typing (Defensive Programming)

### 6. Null Safety & Type Safety
*   **DEFENSE:** Never access `map.get("key")` or `list.get(0)` without verifying its existence first.
    *   Use `ifnull(variable, default_value)` extensively.
    *   Verify lists with `if(list.size() > 0)`.
*   **EXPLICIT CASTING:** Do not trust Deluge's automatic conversion.
    *   Use `.toLong()`, `.toString()`, `.toDecimal()` before any logical or arithmetic comparison.
*   **CONSISTENT RETURNS:** Functions must never return `null`.
    *   Success: Returns the expected data.
    *   Failure/Empty: Returns empty list `[]`, empty map `{}`, or standard error map `{"error": true, "msg": "..."}`.

## IV. Clean Code & Style (General)

### 7. Readability & Structure
*   **EARLY RETURNS (Guard Clauses):** Avoid deep `if/else` nesting.
    *   *Pattern:* Validate negative conditions at the start and return.
    *   `if(!input.get("id")) { return; }` -> The rest of the code stays at the first indentation level.
*   **SINGLE RESPONSIBILITY (SRP):** A function should do one thing only.
    *   If a function "Calculates Taxes" and also "Updates CRM", split it into two.
*   **DESCRIPTIVE NAMING:**
    *   Variables: camelCase. Mandatory suffixes for complex types: `clientMap`, `invoicesList`, `responseJson`.
    *   Forbidden: `x`, `y`, `data`, `info` (too generic).
*   **COMMENTS:**
    *   The code should explain itself.
    *   Use comments only to explain the **WHY** of complex business logic, not the **WHAT** the code does (that is already readable).

## V. Form & Event Architecture (Creator Specific)

### 8. Thin Forms
*   **FORBIDDEN:** Writing complex business logic (>15 lines) directly in `on success`, `on validate`, or `on user input` events.
*   **SOLUTION:** Encapsulate logic in **Global Functions** and invoke them from the event.
    *   *Pattern:* `thisapp.namespace.calculateProfitability(input.ID);`
    *   *Benefit:* Reusability and cleaner event handlers.

### 9. Loop Aggregation
*   **FORBIDDEN:** Iterating over the same list multiple times for different calculations (e.g., one loop for costs, another for sales).
*   **SOLUTION:** Perform ALL necessary calculations (sums, counts, updates) in a **single iteration** of the loop.
    *   *Benefit:* Reduces complexity from O(2n) to O(n), saving execution statements.

## VI. Naming Conventions (Creator Specific)

### 11. Forms
*   **Display name:** `form_X` where `X` is the descriptive name (e.g., `form_ClientOrder`).
*   **API name:** `form_X` in camelCase, replacing accented characters with their unaccented equivalents (e.g., `form_clientOrder`).
*   **Rule:** Every form must follow this prefix — no exceptions.

### 12. Reports
*   **Main report per form:** `form_X_report` — one mandatory report per form containing ALL fields and configured as a list.
    *   API name: `form_X_report` in camelCase.
*   **Page-specific report:** `report_X_pageName` — created only when used in a page, containing only the fields needed for that context.
    *   API name: `report_X_pageName` in camelCase.
*   **Rule:** At minimum one `form_X_report` must exist per form before a page report is created.

### 13. Pages
*   **Display name:** `page_X` where `X` is the descriptive name.
*   **API name:** `page_X` in camelCase, replacing accented characters.

### 14. Custom Actions
*   **Display name:** `ca_X` where `X` describes the action (e.g., `ca_SendInvoice`).
*   **API name:** `ca_X` in camelCase, replacing accented characters.

### 15. API Name General Rules
*   **Always camelCase:** All API names (forms, reports, pages, functions, variables, custom actions) use camelCase.
*   **No accents:** Replace `á→a`, `é→e`, `í→i`, `ó→o`, `ú→u`, `ñ→n`.
*   **No spaces or special characters:** Use underscore `_` only as the prefix separator (`form_`, `report_`, `page_`, `ca_`). The `X` part itself is camelCase with no underscores.

### 16. Function Reuse
*   **MANDATORY:** Extract repeated logic (>3 occurrences or >10 lines) into **Global Functions**.
*   **Naming:** `namespace_verbNoun` pattern (e.g., `thisapp.orders.calculateTotalTax`).
*   **Goal:** No duplicated business logic across events or forms. A single change must propagate everywhere automatically.

## VII. Portability & Environment

### 10. Dynamic URLs
*   **FORBIDDEN:** Hardcoding domains or usernames in URLs (e.g., `https://creator.zoho.eu/user/...`).
*   **SOLUTION:**
    *   Use `zoho.appuri` for internal links.
    *   Construct URLs dynamically: `url = "https://creator.zoho." + var.location_zoho + "/" + zoho.adminuser + "/" + ...`.
    *   Use Organization Variables for external domains that might change between environments (Sandbox/Prod).
