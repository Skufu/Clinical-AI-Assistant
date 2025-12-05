# Clinical AI Assistant

Vanilla HTML/CSS UI plus a Go backend that runs deterministic clinical safety checks and returns a structured treatment plan.

## Run
- `go run main.go` (serves UI and API at http://localhost:8080)
- `make run`

## Test
- `go test ./...`
- `make test`

## Prerequisites
- Go 1.22+
- (Optional) OpenAI API key if enabling a real LLM call instead of the stub.

## API
- POST `/api/analyze`
- Request (example):
```json
{
  "patientName": "Juan Dela Cruz",
  "age": 45,
  "weight": 78,
  "height": 175,
  "bp": "135/88",
  "bmi": 25.5,
  "conditions": ["Hypertension"],
  "allergies": ["None"],
  "medications": [
    {"name": "Amlodipine", "dosage": "5mg", "frequency": "Daily"}
  ],
  "smoking": "Former",
  "alcohol": "Occasional",
  "exercise": "1-2x/week",
  "complaint": "ED"
}
```
- Response (fields):
  - `riskLevel`: LOW | MEDIUM | HIGH | INVALID
  - `riskScore`: integer
  - `flaggedIssues`: list of `{type, severity, description}`
  - `recommendedPlan`: `{medication, dosage, frequency, duration, rationale}`
  - `planConfidence`: number 0-1
  - `alternatives`: list of `{medication, dosage, pros[], cons[], confidence}`
  - `computedBmi`: number
  - `validationErrors`: present on 400 with details
  - `auditId`: opaque audit reference
  - `auditAt`: RFC3339 timestamp
- GET `/api/audit` returns recent audit summaries.

## Notes
- HTML page calls the API directly (same origin).
- Rule engine handles BMI/BP parsing, comorbidity scoring, nitrate/PDE5 contraindications, alpha-blocker/PDE5 warning, alcohol/PDE5 warning, allergy cross-check, dose caps, and complaint-specific plans (ED, hair loss, weight loss, general).
- Response is validated against `internal/analysis/schema/response.schema.json` before returning.
- LLM guardrail: deterministic rules merged with a stubbed LLM confidence scorer; swap `callLLMStub` for a real LLM client (system prompt in `analysis.go`) if keys are available.
- Wizard flow includes a doctor review/edit step and shows audit ID in the approval summary.
- Docker: `docker build -t clinical-ai .` then `docker run -p 8080:8080 clinical-ai`.

## LLM integration (how to replace the stub)
- Implement a real client in `analysis.go` where `callLLMStub` is defined; keep the system prompt string provided there.
- Ensure the LLM response conforms to the schema before returning; keep deterministic guardrails as a fail-safe.
- Add any API keys via environment variables and avoid logging PHI.

## Wizard flow
- Sections: Intake → Analysis → Doctor Review/Edit → Approval.
- Review step allows editing medication/dose/frequency/duration/rationale; approval summary includes audit ID.

## Env example (place in your shell or env file)
```
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1  # optional override
PORT=8080
```

## Safety measures
- Deterministic guardrails for contraindications/interactions/dose caps and allergy checks remain authoritative even with LLM output.
- Schema validation on every response; invalid outputs return 400 with details.
- Audit logging (redacted patient ref) with audit ID/timestamp; minimal logging of PHI (no payloads).
- Client/server validation for required fields and BP format; UI blocks submission until valid.
- Review/edit step before approval to keep clinician-in-the-loop.

