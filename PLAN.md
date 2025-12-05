# Delivery Plan

Scope: finish required/bonus items from instructions.md with guardrails-first approach.

## Hackathon Track (fastest path to demo)
1) LLM + Schema
   - Use single OpenAI client with stub fallback; enrich system prompt with full dosing/contra/interaction rules (no stubs).
   - Share backend JSON Schema to frontend (generate types) and validate responses client-side; on schema failure, fall back to deterministic checks and surface errors.
   - Return model-derived confidence scores per rec/alternative (fallback: deterministic score with "low confidence" tag).
2) Safety Data
   - Bundle a local interaction/contra ruleset; optionally enrich via one external API with caching/backoff; degrade gracefully to local data.
3) Wizard UX
   - Enforce Intake → AI analysis → Doctor review/edit → Final summary; block approval if audit write fails or validation errors remain.
   - Show inline “why” for each risk/alternative; keep risk colors and rationale concise.
4) Audit/Identity
   - Persist audits to SQLite; include lightweight user identifier on requests and approvals; expose audit history in summary.
5) Demo Readiness
   - Add a clear high-risk sample patient showcasing flagged risks and rationale.
   - Keep env simple: `OPENAI_API_KEY`, optional `OPENAI_BASE_URL`, `PORT`, `SQLITE_PATH`.

## Confirmed Assumptions/Decisions
- Frontend does not yet consume or validate the backend JSON Schema.
- External drug interaction source is not configured; select an API/dataset and keep bundled fallback.
- Audit persistence will use SQLite; user identity capture still needed for approvals/history.
- No auth layer; add lightweight user identifier for audit logging and approvals.
- Confidence scores are currently stubbed and must come from the model.

## Risk Checkpoints
- Schema drift between backend and UI; share schema/types and validate client-side.
- External drug API reliability/rate limits; implement caching/backoff and local fallback.
- Missing user identity blocking audit persistence; add identifier injection in requests and UI flow.
- LLM output not matching schema; enforce server/client validation and safe fallbacks.
- UI must block approval on audit persistence/validation failures and surface errors clearly.

## Workstream 1 – LLM + Schema
- [x] Add system prompt encoding dosing rules, contraindications, interactions, risk escalation.
- [x] Implement LLM call (stub fallback) and merge with deterministic checks.
- [x] Define JSON Schema for response; validate server-side before returning to UI.
- [x] Surface validation errors to frontend (list + inline cues).
- [ ] Swap stub with real LLM client (e.g., OpenAI go-openai, model gpt-4o-mini). Keep guardrails and schema validation; use env vars `OPENAI_API_KEY`, optional `OPENAI_BASE_URL`.
- [ ] Replace placeholder medical rules in system prompt with full clinical rule content (avoid stubs).
- [ ] Share explicit JSON Schema contract with frontend and validate responses client-side to mirror backend schema.
- [ ] Return confidence scores per recommendation/alternative sourced from the model (not stubbed).

## Workstream 2 – Safety/Rules/Data
- [x] Extend rule engine with more drug/condition interactions and dose caps.
- [x] Integrate drug interaction/contra datasource (API or local ruleset) and combine results.
- [x] Add allergy cross-checks against plan/alternatives.
- [x] Add confidence scores per recommendation and per alternative.
- [ ] Swap local/stubbed interaction data with a real external drug interaction/contraindication source; keep local fallback.
- [ ] Broaden dose appropriateness and condition-specific contraindication coverage across more medications.
- [ ] Add a clear high-risk sample patient demonstrating flagged risks and rationale.

## Workstream 3 – UX Flow
- [x] Convert to wizard: Intake → AI analysis → Doctor review/edit → Final summary. *(Basic review step added; could deepen)*
- [x] Allow clinician edits to plan before approval; track diffs. *(Edits captured in summary; no diff log)*
- [x] Strengthen rationale display (why + safety notes + confidence).
- [x] Add multiple sample patients, including high-risk, to demo flags.
- [ ] Enforce explicit doctor review/edit → approval flow with approvals logged and visible in UI.
- [ ] Improve risk indicators to highlight “why” for each issue and alternative in-line.

## Workstream 4 – Audit/Logging
- [x] Implement audit log (who/when/what) for analysis, edits, approvals. *(In-memory; no user identity)*
- [x] Add minimal request/result logging (redact PHI as needed).
- [x] Expose audit entries in UI summary step. *(Audit ID shown; endpoint added)*
- [ ] Persist audit logs with user identity and timestamps (not in-memory only); surface review/approval history in UI.

## Workstream 5 – Tooling/Quality
- [x] Add Make targets (run/test/lint/schema).
- [x] Add Dockerfile for deploy.
- [x] Add tests covering validation, schema conformance, interactions, audit logging.
- [x] Document usage in README (LLM config, schema contract, wizard flow).***
- [ ] Add .env.example (OPENAI_API_KEY, OPENAI_BASE_URL, PORT) or document env block for deployment.

