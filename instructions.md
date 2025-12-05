Your mission, should you choose to accept it, is to build an AI-powered clinical assistant that turns patient intake data and medical history into a personalized, safety-checked treatment plan.

This tool should help doctors make faster, data-driven decisions while flagging drug interactions, contraindications, and risk factors—not replacing clinicians, but supercharging them. 
Andre Dalwin
OP
 — 10:37 am
Inputs
Design a patient intake flow that captures at least:

Medical history (conditions, allergies)
Current medications (drug name, dosage, frequency)
Health metrics (age, weight, BMI, blood pressure)
Lifestyle factors (smoking, alcohol, exercise)
Primary complaint (e.g. ED, hair loss, weight loss)

You may hardcode sample patients or build a live form, as long as the flow is clear. 
Andre Dalwin
OP
 — 10:45 am
Core Logic
Use an LLM API (e.g. GPT / Claude / Gemini) with medical guidelines encoded in the system prompt.
The model must:
Check for drug–drug interactions
Check for contraindications from conditions/allergies
Check dosage appropriateness and basic risk factors
Produce structured, parseable output (e.g. JSON schema).
Extra credit for using structured medical databases (drug interaction rules, contraindications) to validate or enrich the LLM output.
Expected Output / UX
Build a clinical decision support dashboard that surfaces:

Recommended treatment plan (medication, dosage, duration)
Safety risk score (e.g. Low / Medium / High)
Flagged contraindications or drug interactions
Alternative treatment options
Explanation / rationale for each recommendation

Design for busy doctors: show critical risks first, then details and rationale. The UI should make it easy to approve, modify, or reject the plan.
To Complete This Bounty
To count as "solved", your submission must include:

A working intake flow that collects the required patient data.
At least one LLM integration using medical rules in the system prompt.
Structured output (e.g. JSON) with: treatment plan, risk level, and rationale.
A basic dashboard UI that:
Displays the treatment plan
Shows a safety risk indicator
Lists any flagged issues (contraindications / interactions).
At least one realistic example patient that demonstrates risks being flagged.
Bonus / Advanced
Bonus points (not required, but strongly valued):

Implement a JSON schema for the LLM’s response and validate against it.
Use a drug interaction / contraindication database (PostgreSQL, API, etc.) to cross-check model output.

Build a multi-step wizard:
Intake → AI analysis → Doctor review/edit → Final summary.
Add confidence scores per recommendation and alternative treatment paths.
Include audit logging for medical compliance (who reviewed what, when).
 
Additional Details
Safety > everything:
Always flag high-risk cases aggressively; false positives are better than misses.
Prompting:
Encode dosing rules, contraindications, and interaction rules directly in your system prompts.
24h feasibility:
Start with a simple form → LLM call → text/JSON display.
You don’t need a complex backend; the LLM can handle most of the medical logic.
UX guidelines:
Show red/orange/green risk indicators.
Make the "why" behind recommendations easy to see (short rationales).

Hackathon optimizations (fast, safe, minimal)
- No auth layer; use a lightweight user identifier passed through requests for audit and approvals.
- Persist audit logs to SQLite (file-backed) with timestamps and reviewer ids; fail approval if persistence fails.
- Keep external dependencies optional: prefer a bundled drug interaction/contra ruleset with optional API enrichment plus caching/backoff.
- Share the JSON Schema contract with the frontend; validate LLM responses both server- and client-side, and fall back to deterministic checks on invalid output.
- Use a single LLM client with stub fallback; keep system prompt packed with dosing/contra/interaction rules.
- Enforce wizard flow: Intake → AI analysis → Doctor review/edit → Final summary; show risks and rationales inline with approvals logged.
