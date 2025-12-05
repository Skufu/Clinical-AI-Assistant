package analysis

import (
	"testing"

	"github.com/Skufu/Clinical-AI-Assistant/internal/audit"
)

func TestAnalyze_EDAmlodipineInteraction(t *testing.T) {
	input := Intake{
		PatientName: "Juan Dela Cruz",
		Age:         45,
		WeightKg:    78,
		HeightCm:    175,
		BP:          "135/88",
		Conditions:  []string{"Hypertension"},
		Medications: []Medication{
			{Name: "Amlodipine", Dosage: "5mg", Frequency: "Daily"},
		},
		Complaint: "ED",
	}

	resp := Analyze(input)

	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}

	if resp.RiskLevel != "LOW" {
		t.Fatalf("expected LOW risk, got %s (score %d)", resp.RiskLevel, resp.RiskScore)
	}

	if resp.RecommendedPlan.Medication != "Tadalafil" {
		t.Fatalf("expected Tadalafil plan, got %s", resp.RecommendedPlan.Medication)
	}

	if !hasIssue(resp.FlaggedIssues, "drug_interaction") {
		t.Fatalf("expected drug interaction warning for amlodipine + PDE5")
	}
}

func TestAnalyze_NitrateContraindication(t *testing.T) {
	input := Intake{
		PatientName: "High Risk",
		Age:         68,
		WeightKg:    90,
		HeightCm:    170,
		BP:          "168/102",
		Conditions:  []string{"Heart Disease", "Hypertension"},
		Medications: []Medication{
			{Name: "Nitroglycerin", Dosage: "0.4mg", Frequency: "PRN"},
		},
		Complaint: "ED",
	}

	resp := Analyze(input)

	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}

	if resp.RiskLevel != "HIGH" {
		t.Fatalf("expected HIGH risk, got %s (score %d)", resp.RiskLevel, resp.RiskScore)
	}

	if !hasIssue(resp.FlaggedIssues, "contraindication") {
		t.Fatalf("expected nitrate contraindication to be flagged")
	}

	if usesPDE5(resp.RecommendedPlan.Medication) {
		t.Fatalf("plan should avoid PDE5 when nitrates present, got %s", resp.RecommendedPlan.Medication)
	}
}

func TestAnalyze_WeightLossRiskStratification(t *testing.T) {
	input := Intake{
		PatientName: "Weight Loss",
		Age:         50,
		WeightKg:    110,
		HeightCm:    175,
		BP:          "150/95",
		Conditions:  []string{"Hypertension"},
		Complaint:   "Weight Loss",
	}

	resp := Analyze(input)

	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}

	if resp.RiskLevel != "MEDIUM" {
		t.Fatalf("expected MEDIUM risk, got %s (score %d)", resp.RiskLevel, resp.RiskScore)
	}

	if resp.RecommendedPlan.Medication != "Metformin" {
		t.Fatalf("expected Metformin plan, got %s", resp.RecommendedPlan.Medication)
	}
}

func TestAnalyze_TamsulosinInteraction(t *testing.T) {
	input := Intake{
		PatientName: "Alpha Blocker",
		Age:         55,
		WeightKg:    82,
		HeightCm:    178,
		BP:          "138/90",
		Conditions:  []string{"Hypertension"},
		Medications: []Medication{
			{Name: "Amlodipine", Dosage: "5mg", Frequency: "Daily"},
			{Name: "Tamsulosin", Dosage: "0.4mg", Frequency: "Daily"},
		},
		Complaint: "ED",
	}

	resp := Analyze(input)
	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}
	if !hasIssue(resp.FlaggedIssues, "drug_interaction") {
		t.Fatalf("expected drug interaction warning for tamsulosin + PDE5")
	}
}

func TestAnalyze_AllergyCrossCheck(t *testing.T) {
	input := Intake{
		PatientName: "Allergy",
		Age:         40,
		WeightKg:    70,
		HeightCm:    170,
		BP:          "120/80",
		Allergies:   []string{"tadalafil"},
		Complaint:   "ED",
	}

	resp := Analyze(input)
	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}
	if !hasIssue(resp.FlaggedIssues, "allergy") {
		t.Fatalf("expected allergy issue flagged")
	}
}

func TestAnalyze_AuditAndSchema(t *testing.T) {
	input := Intake{
		PatientName: "Schema",
		Age:         45,
		WeightKg:    78,
		HeightCm:    175,
		BP:          "125/80",
		Complaint:   "Hair Loss",
	}

	resp := Analyze(input)
	if len(resp.ValidationErrors) > 0 {
		t.Fatalf("unexpected validation errors: %v", resp.ValidationErrors)
	}
	if resp.AuditID == "" {
		t.Fatalf("expected audit id to be set")
	}
	if resp.PlanConfidence <= 0 {
		t.Fatalf("expected plan confidence to be set")
	}
	if errs := ValidateResponse(resp); len(errs) > 0 {
		t.Fatalf("response should satisfy schema, got: %v", errs)
	}
}

func TestLatestAuditsLimit(t *testing.T) {
	SetAuditStore(audit.NewMemoryStore())

	input := Intake{
		PatientName: "Audit",
		Age:         45,
		WeightKg:    78,
		HeightCm:    175,
		BP:          "125/80",
		Complaint:   "Hair Loss",
	}

	for i := 0; i < 55; i++ {
		Analyze(input)
	}

	audits := LatestAudits(50)
	if len(audits) != 50 {
		t.Fatalf("expected 50 audits returned, got %d", len(audits))
	}
}
func TestAnalyze_Validation(t *testing.T) {
	input := Intake{}
	resp := Analyze(input)
	if len(resp.ValidationErrors) == 0 {
		t.Fatalf("expected validation errors for empty intake")
	}
	if resp.RiskLevel != "INVALID" {
		t.Fatalf("expected INVALID risk level for validation failures, got %s", resp.RiskLevel)
	}
}

func hasIssue(issues []Issue, issueType string) bool {
	for _, i := range issues {
		if i.Type == issueType {
			return true
		}
	}
	return false
}
