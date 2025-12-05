package analysis

import "testing"

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

	if resp.RiskLevel != "MEDIUM" {
		t.Fatalf("expected MEDIUM risk, got %s (score %d)", resp.RiskLevel, resp.RiskScore)
	}

	if resp.RecommendedPlan.Medication != "Metformin" {
		t.Fatalf("expected Metformin plan, got %s", resp.RecommendedPlan.Medication)
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
