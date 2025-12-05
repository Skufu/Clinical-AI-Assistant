# Clinical AI Assistant

Vanilla HTML/CSS UI plus a Go backend that runs deterministic clinical safety checks and returns a structured treatment plan.

## Run
- `go run main.go` (serves UI and API at http://localhost:8080)

## Test
- `go test ./...`

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
  - `riskLevel`: LOW | MEDIUM | HIGH
  - `riskScore`: integer
  - `flaggedIssues`: list of `{type, severity, description}`
  - `recommendedPlan`: `{medication, dosage, frequency, duration, rationale}`
  - `alternatives`: list of `{medication, dosage, pros[], cons[]}`
  - `computedBmi`: number

## Notes
- HTML page calls the API directly (same origin).
- Rule engine handles BMI/BP parsing, comorbidity scoring, nitrate/PDE5 contraindications, and complaint-specific plans (ED, hair loss, weight loss, general).***

