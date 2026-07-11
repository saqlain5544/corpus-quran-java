# AGENTS.md — Quran Reader App (Project-Specific)

This document defines the sub-agents required to design, verify, and implement the Quran Reader App, using the Minimax M3 multi-agent orchestration system. All agents must adhere to the project constraints (lowest memory footprint, vanilla stack, editorial UI, only trusted data), and follow the Plan → Pre-implementation → Test-driven Implementation → Debug lifecycle.

The global `~/.config/opencode/AGENTS.md` overrides anything in this file that conflicts. This document is project-specific guidance — read the global one first for tools, plugins, MCP, and skill-system conventions.

This file follows the opencode/AGENTS.md convention: top-level sections use uppercase (`## CORE PRINCIPLES`), subsections use sentence case, code references use `path:line`, and examples are fenced with language hints.

---

## CORE PRINCIPLES (project-specific)

1. **Memory footprint is the primary constraint.** Design for the minimum working set in RAM; the 80MB RSS target is a hard ceiling. No defensive caching, no pre-loaded indexes, no eager parsing. Stream large files (quran-uthmani.xml) rather than DOM-load them.
2. **Vanilla stack only.** No React, no Vue, no jQuery, no Lodash. Vanilla JS + CSS + HTML, with optional Go server-side templating. Every dependency must justify its weight in KB.
3. **Editorial UI typography-first.** No garish colors, no icon-heavy design, no decorative animation. Treat the corpus as primary content — typography and spacing carry the design.
4. **Only trusted data sources.** Three sources only:
   - `./data/quran/quran-uthmani.xml` (Uthmani rasm)
   - `./data/morphology/MASAQ.csv` (Buckwalter-lemma tagged)
   - `./data/fonts/hafs.woff2` (display font)
   All other data files must be flagged "unverified" in the UI unless the Research Agent has validated them.
5. **Accessibility is non-negotiable.** Semantic HTML, ARIA labels where needed, keyboard-navigable modals and tooltips. Every interactive element must work without a mouse.
6. **Reference-grounded, always.** Cite `path:line` for code, surah:ayah for scripture references, edition/version for data citations.

---

## AGENT HIERARCHY

- **Orchestrator Agent** (Minimax M3 main agent): delegates tasks, monitors progress, resolves conflicts, ensures global constraints respected. Only agent that interacts with the user for clarifications and sign-off.
- **Specialist Sub-agents**: Research, Planner, Data-Pipeline, Test-Writer, Backend-Implementer, Frontend-Implementer, Integration-Debugger. Each receives a structured task, accesses allowed tools, returns results to Orchestrator.

---

## 1. RESEARCH AGENT

**Role:** Gathers information, validates external data, clarifies ambiguities.
**Goal:** Every design decision backed by accurate domain knowledge, reliable data, authoritative sources before any code is written.

### Capabilities

- Search the internet (web search) for scholarly resources on Quranic text, Tajweed marks, morphological analysis, Indo-Pak vs Uthmani script differences.
- Access local trusted data files (`./data/quran/quran-uthmani.xml`, `./data/morphology/MASAQ.csv`, `./data/fonts/hafs.woff2`) and verify integrity by spot-checking against known references.
- Cross-check supplementary data files (`root_meaning.csv`, frequency files, etc.) and flag any untrustworthy data with explicit warnings.
- Investigate fuzzy matching algorithms for Arabic (with/without tashkeel, script variants) and English glosses.
- Research best practices for low-memory Go servers, efficient XML parsing, minimal client-side rendering.

### Tools

- `web_search` (to query tanzil.net documentation, Arabic NLP libraries, etc.)
- `read_file` (to inspect CSV/XML/JSON files in `./data`)
- `run_terminal` (to run checks like `wc -l`, `head`, `xmllint`, etc.)

### Inputs

- Specific research queries from the Orchestrator (e.g., "How is a Rub' mark encoded in Uthmani XML?", "List all pause marks in the XML schema", "What are the Indo-Pak vs Uthmani alif wasla differences?")
- Paths to data files to be validated.

### Outputs

- A structured research note (markdown) with findings, citations, explicit verdicts: "Trusted", "Needs manual review", or "Do not use".
- Recommendations for algorithmic approaches (e.g., which fuzzy matching library or technique).

---

## 2. PLANNING & ARCHITECTURE AGENT

**Role:** Decomposes the entire project into atomic, implementable components; defines data flows, UI states, route hierarchy, and system architecture.
**Goal:** Produce a complete "paper-based" blueprint before any code is written, including flowcharts, component trees, and route-to-handler mappings.

### Capabilities

- Create detailed flowcharts (in Mermaid or ASCII) for user interactions: search flows, surah navigation, tooltip loading, modal opening.
- Design the data pipeline: how XML is parsed into memory-efficient Go structs, how MASAQ.csv is indexed for O(1) word lookups, how roots are linked.
- Define all Go structs (Surah, Ayah, Word, Token, RootEntry, MasaqRecord, etc.) with field annotations and JSON serialization.
- Specify the 12-column grid system, design tokens (colours, typography, z-indices, spacing), and editorial UI layout for every page and component.
- Map HTTP routes (`/`, `/surah/{id}`, `/surah/{id}/{verse}`, `/search`, `/roots`, `/root/detailed/{root}`) to their data sources and templates.
- Identify atomic components: GlobalHeader, SurahHeader, SurahParagraph, WordTooltip, WordModal, SearchResultsList, RootsList, RootDetailList, etc.
- Define client-side JavaScript modules (state management, event delegation, dynamic tooltip positioning, modal fetching).

### Tools

- `write_file` (to save flowcharts and specification documents in `docs/pen-and-paper/`)
- `mermaid` (if supported) or ASCII art drawing.

### Inputs

- The full project brief.
- Research Agent outputs (validated data schemas, script differences, fuzzy matching advice).

### Outputs

- A set of documents in `docs/pen-and-paper/`:
  - `architecture.md` (overall system, data structures, memory strategy)
  - `routes-api.md` (URL patterns, query params, server handler pseudocode)
  - `component-tree.md` (atomic components, their props/state, nesting)
  - `ui-design-system.md` (colours, fonts, grid, spacing, z-indices, editorial rules)
  - `data-flow-diagrams.md` (flow from request to response, tooltip loading, modal)
- Updated `docs/to-dos/tasks.md` with initial task breakdown.

---

## 3. DATA PIPELINE & PRE-IMPLEMENTATION AGENT

**Role:** Builds and tests terminal-based scripts (in any language, e.g., Python, Go, shell) that transform raw data into the exact formats needed by the Go server, verifying correctness and edge cases before implementation.
**Goal:** All data pre-processing is correct, complete, verifiable, so the Go runtime can consume pre-validated, optimized data structures directly.

### Capabilities

- Write scripts to:
  - Parse `quran-uthmani.xml` and extract surahs, ayahs, words, pause marks, sajdah, rub' marks, separating textual tokens from annotation marks.
  - Parse `MASAQ.csv` and build an in-memory index (hashmap) keyed by `(surah, ayah, word_position)` or a unique word ID, storing morphological details and root.
  - Cross-join the two sources, producing a unified JSON/GOB/binary file per surah, containing all word data and annotations, ready for fast loading.
  - Verify that every word in the XML has a corresponding entry in MASAQ or is correctly flagged as "no root".
  - Validate the root frequency data, check for synonym roots if required, and generate the root listing.
  - Generate a search index for root, English gloss (fuzzy-compatible), and Arabic word tokens.
- Run exhaustive edge-case tests: first surah, last surah, surah with many marks, words with alif wasla variants, etc.
- Compare script output with trusted manual spot-checks from the Research Agent.

### Tools

- `run_terminal` (to execute scripts, run diff checks, validate JSON)
- `read_file` / `write_file` for intermediate data files.

### Inputs

- Paths to raw data and Research Agent's validated schema.
- Architecture Agent's data structure definitions.

### Outputs

- Validated, compressed data files (e.g., `./data/processed/quran_words.gob`, `./data/processed/masaq_index.gob`, `./data/processed/root_index.json`, search index) ready to be embedded or loaded at server start.
- Verification reports (CSV of mismatches, if any) in `docs/data-verification/`.
- A `README.md` in `./data/processed` explaining generation steps and how to regenerate.

---

## 4. TEST-WRITER AGENT

**Role:** Before any production code is written, creates comprehensive test suites for every function, struct, HTTP handler, and UI component.
**Goal:** Test-driven development (TDD) approach where every piece of behavior is specified as a failing test first.

### Capabilities

- Write Go table-driven tests for:
  - XML parsing functions (with sample XML snippets, error injection)
  - MASAQ CSV parsing (missing fields, extra commas, Arabic encoding)
  - Word-to-root lookup
  - Surah pagination/build logic (correct ayah grouping, mark insertion)
  - HTTP handlers (using `httptest`) for all routes, verifying response status, content-type, and body snippets.
  - Search function (root, English, Arabic fuzzy) with various input cases (empty, no results, many results).
- Write JavaScript unit tests (using a minimal test framework or vanilla test harness) for:
  - Tooltip positioning logic (viewport edges, scroll)
  - Modal fetch and render pipeline
  - Client-side search highlighting
  - Theme toggling and persistence
- Define integration tests that simulate a full request-response cycle with pre-loaded test data.
- Ensure all tests run in a memory-constrained environment (simulate low memory if possible).

### Tools

- `run_terminal` to execute `go test`, `node` for JS tests.
- `write_file` to create test files.

### Inputs

- Architecture Agent's component and route specifications.
- Data Pipeline Agent's sample data sets.

### Outputs

- Complete test files (`*_test.go`, `*.test.js`) with instructions for running.
- A test coverage plan document indicating which parts of the code are covered.

---

## 5. BACKEND IMPLEMENTATION AGENT (Go)

**Role:** Writes all Go server code, following TDD and the atomic, modular, DRY principles.
**Goal:** Build a low-memory, fast, correct backend that serves HTML pages, tooltip/modal data, and search endpoints using only the vanilla Go standard library (plus maybe a minimal router like `http.ServeMux`).

### Capabilities

- Implement:
  - `main.go` with server startup, embedding of pre-processed data (using `embed` if feasible), route registration.
  - Data loading module that reads the pre-validated binary files into thread-safe structs.
  - Handlers for:
    - Homepage: list surahs (from data), render template.
    - Surah page: load full surah, render with words in paragraph form, attach JavaScript for tooltips/modals.
    - Verse-anchored surah page: same but scroll to verse.
    - Search: parse `q` and `type`, perform fuzzy matching using server-side logic, render results.
    - Roots list: paginated, with frequency sorting (asc/desc) and surah filter.
    - Root detail: list occurrences, support filtering.
    - API endpoints (JSON) for tooltip data (`/api/word/{surah}/{ayah}/{pos}`) and modal data (`/api/word/detail/...`), returning MASAQ info and root links.
  - Template rendering (using `html/template`) for all pages, applying the editorial design system passed from the Architecture Agent.
  - Proper error handling, 404/500 pages.
  - Middleware for logging, recovery, and memory monitoring.
- Ensure all Go code passes the Test-Writer's tests (run `go test ./...` before marking task complete).
- Enforce strict memory usage: avoid loading entire XML into DOM-like structures; use streaming, and keep only essential indexes in memory.

### Tools

- `run_terminal` (to compile, run tests, check memory with `runtime` stats)
- `write_file`

### Inputs

- Pre-processed data files.
- Architecture Agent's route specifications.
- Test-Writer Agent's test suites.

### Outputs

- Complete Go source code in `./backend/` or at project root.
- A `go.mod` file with zero external dependencies (or only `golang.org/x/net` for additional HTTP features if absolutely necessary, with approval).

---

## 6. FRONTEND IMPLEMENTATION AGENT (Vanilla JS/CSS/HTML)

**Role:** Crafts all client-side code and templates according to the editorial design system, ensuring pixel-perfect layouts, responsive grid, and flawless interactivity.
**Goal:** Deliver a lightweight, accessible, and beautiful UI that works entirely without frontend frameworks.

### Capabilities

- Write semantic HTML templates for all pages, following the 12-column grid, typography scale, and colour tokens.
- Create a comprehensive CSS design system:
  - Utility classes (margins, padding, text alignment, flex helpers) but primarily context-specific token classes (e.g., `.surah-text`, `.ayah-marker`, `.tooltip-container`).
  - Dark/light theme via CSS custom properties and a `data-theme` attribute on `<html>`.
  - Strict z-index scale for header (sticky), tooltip, modal, etc.
  - `hafs.woff2` font-face declaration and correct Arabic rendering (RTL, ligatures).
- Implement JavaScript modules:
  - `global-header.js`: search box with type selector, theme toggler (persisted in localStorage), navigation links.
  - `surah-header.js`: local surah search (scroll to ayah, error feedback), font size & line height sliders, previous/next surah buttons.
  - `word-tooltip.js`: attach event delegation for hover/click; fetch tooltip data from API and position (center-aligned above/below word), handle edge-of-viewport flipping.
  - `word-modal.js`: on click, open modal with detailed data, root link; handle loading state, close behaviour.
  - `verse-scroll.js`: if URL hash `#verseX`, smoothly scroll to that ayah.
  - `infinite-scroll.js` (if needed for roots list) or simple pagination.
- Ensure all interactive elements are keyboard-accessible.
- Optimize for lowest memory: attach event listeners on parent containers, use `innerHTML` sparingly, clean up modals after close.

### Tools

- `write_file` for HTML, CSS, JS files.
- `run_terminal` to optionally use a static analysis linter for JS/CSS.

### Inputs

- Architecture Agent's component tree, design tokens, and UI flow charts.
- Backend Agent's API contract (endpoint URLs, response format).
- Validated font file.

### Outputs

- `./templates/` directory with Go-templatable HTML files.
- `./static/css/` with main stylesheet and any needed component CSS.
- `./static/js/` with all vanilla scripts, bundled into as few files as needed (perhaps one core bundle).

---

## 7. INTEGRATION & DEBUGGING AGENT

**Role:** Assembles all parts, runs full system tests, profiles memory, and fixes any issues discovered.
**Goal:** Ensure the whole app works as specified, with zero regressions and under the memory footprint target.

### Capabilities

- Set up a local test environment, starting the Go server with the embedded data.
- Run automated Playwright-style end-to-end tests (if possible using a minimal headless browser script) or manual test plan execution.
- Profile memory with `go tool pprof` and verify that the app stays within an acceptable footprint (e.g., < 80 MB RSS).
- Debug discrepancies between test data and live behaviour.
- Fix bugs in any layer (backend, frontend, data pipeline) while preserving test coverage; may invoke other sub-agents for complex fixes.
- Verify all editorial design requirements: typography, spacing, theme, sticky headers stacking.
- Validate that all marks (pause, sajdah, rub') appear correctly in the text and are not confused with words.

### Tools

- `run_terminal` (for server start, profiling, curl tests)
- `web_fetch` or browser tools for visual inspection (if available)

### Inputs

- Complete codebase from Backend and Frontend Agents.
- All test suites.
- Pre-processed data files.

### Outputs

- A final, tested application.
- A test report detailing passed/failed edge cases.
- Memory and performance profile summary.
- A list of any known issues (if any) with explanations.

---

## AGENT WORKFLOW SEQUENCE

1. **Research Agent** validates data and answers domain questions → results sent to Orchestrator.
2. **Planner Agent** creates full architecture and UI design documents → Orchestrator approves.
3. **Data Pipeline Agent** pre-processes and verifies data → produces trusted binary/index files.
4. **Test-Writer Agent** writes all tests (Go + JS) based on architecture documents.
5. **Backend Implementation Agent** writes Go server code until all backend tests pass.
6. **Frontend Implementation Agent** writes templates, CSS, and JS until all UI tests pass.
7. **Integration & Debugging Agent** runs the full system, profiles, fixes integration bugs, and delivers the finished app.

Throughout each phase, the Orchestrator may invoke the **Research Agent** for additional clarifications or the **Test-Writer Agent** for additional edge cases found during implementation.

---

## SKILL SYSTEM (project-specific guidance)

In addition to the global skill system (`~/.agentic-engineering/skills/` and `~/.config/opencode/skill/`), this project has Quran-specific skills auto-loaded by opencode's `skill` tool when relevant:

- `~/.config/opencode/skill/quran-data/` — Quranic morphology, tajweed, qira'at, rasm context. Triggered by `quran`, `surah`, `ayah`, `sajda`, `hizb`, `juz`, `rasm`, `sarf`, `tajweed`, `qira'at`, `morphology`. **Read this before writing any data-shape, validation, or test code that involves Arabic Quran text.**

When writing project-specific code:

1. If your task matches a `quran-data` trigger keyword, **load that skill first** via opencode's `skill` tool or `mcp__agent-memory__skill_view(name="quran-data")`.
2. If you discover a multi-step procedure that will recur (e.g. "validate and append a root to the corpus"), consider promoting it to a skill via `mcp__agent-memory__skill_create(name, content, description, triggers)`. **Always pass `description`** — it tells the agent WHEN to invoke the skill (Claude protocol requires a "Use when..." or `when_to_use` field).
3. Skill files use YAML frontmatter + markdown body. The protocol is enforced by tests at `~/.agentic-engineering/ov-mcp/ov_mcp/core/test_claude_skill_protocol.py`.

**Note on skill pre-loading:** the agent-memory stack does NOT pre-load skills at session start. Skills are surfaced lazily when the model calls `skill_list`/`skill_view` (MCP) or when the user explicitly invokes via the `skill` tool. Trigger matching happens at query time, not preemptively.

---

## GLOBAL CONSTRAINTS

The constraints below apply to every agent on every task:

- **Data trust**: only `quran-uthmani.xml`, `MASAQ.csv`, and `hafs.woff2` are initially trusted. All other data must be flagged "unverified" in the UI unless fully verified by the Research and Data Pipeline agents.
- **Memory**: Go server must remain under 80MB RSS with the full dataset loaded; no caching of large XML trees.
- **No external frontend frameworks**: only vanilla JS, CSS, HTML. Use `fetch` API, no jQuery.
- **Editorial UI**: typography-first design, 12-column grid, no garish colours.
- **Accessibility**: semantic HTML, aria labels where needed, keyboard-navigable modals and tooltips.

All agents must respect these constraints, and the Orchestrator must reject any implementation that violates them.