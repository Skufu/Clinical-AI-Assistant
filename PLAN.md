# Delivery Plan

Scope: finish required/bonus items from instructions.md with guardrails-first approach.

## Workstream 1 – LLM + Schema
- [x] Add system prompt encoding dosing rules, contraindications, interactions, risk escalation.
- [x] Implement LLM call (stub fallback) and merge with deterministic checks.
- [x] Define JSON Schema for response; validate server-side before returning to UI.
- [x] Surface validation errors to frontend (list + inline cues).
- [ ] Swap stub with real LLM client (e.g., OpenAI go-openai, model gpt-4o-mini). Keep guardrails and schema validation; use env vars `OPENAI_API_KEY`, optional `OPENAI_BASE_URL`.

## Workstream 2 – Safety/Rules/Data
- [x] Extend rule engine with more drug/condition interactions and dose caps.
- [x] Integrate drug interaction/contra datasource (API or local ruleset) and combine results.
- [x] Add allergy cross-checks against plan/alternatives.
- [x] Add confidence scores per recommendation and per alternative.

## Workstream 3 – UX Flow
- [x] Convert to wizard: Intake → AI analysis → Doctor review/edit → Final summary. *(Basic review step added; could deepen)*
- [x] Allow clinician edits to plan before approval; track diffs. *(Edits captured in summary; no diff log)*
- [x] Strengthen rationale display (why + safety notes + confidence).
- [x] Add multiple sample patients, including high-risk, to demo flags.

## Workstream 4 – Audit/Logging
- [x] Implement audit log (who/when/what) for analysis, edits, approvals. *(In-memory; no user identity)*
- [x] Add minimal request/result logging (redact PHI as needed).
- [x] Expose audit entries in UI summary step. *(Audit ID shown; endpoint added)*

## Workstream 5 – Tooling/Quality
- [x] Add Make targets (run/test/lint/schema).
- [x] Add Dockerfile for deploy.
- [x] Add tests covering validation, schema conformance, interactions, audit logging.
- [x] Document usage in README (LLM config, schema contract, wizard flow).***
- [ ] Add .env.example (OPENAI_API_KEY, OPENAI_BASE_URL, PORT) or document env block for deployment.

