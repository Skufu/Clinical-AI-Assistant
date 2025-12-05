package analysis

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Intake struct {
	PatientName string       `json:"patientName"`
	Age         int          `json:"age"`
	WeightKg    float64      `json:"weight"`
	HeightCm    float64      `json:"height"`
	BP          string       `json:"bp"`
	BMI         float64      `json:"bmi"`
	Conditions  []string     `json:"conditions"`
	Allergies   []string     `json:"allergies"`
	Medications []Medication `json:"medications"`
	Smoking     string       `json:"smoking"`
	Alcohol     string       `json:"alcohol"`
	Exercise    string       `json:"exercise"`
	Complaint   string       `json:"complaint"`
}

type Medication struct {
	Name      string `json:"name"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency"`
}

type Issue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // danger | warning | info
	Description string `json:"description"`
}

type Plan struct {
	Medication string `json:"medication"`
	Dosage     string `json:"dosage"`
	Frequency  string `json:"frequency"`
	Duration   string `json:"duration"`
	Rationale  string `json:"rationale"`
}

type Alternative struct {
	Medication string   `json:"medication"`
	Dosage     string   `json:"dosage"`
	Pros       []string `json:"pros"`
	Cons       []string `json:"cons"`
}

type Response struct {
	RiskLevel       string        `json:"riskLevel"`
	RiskScore       int           `json:"riskScore"`
	FlaggedIssues   []Issue       `json:"flaggedIssues"`
	RecommendedPlan Plan          `json:"recommendedPlan"`
	Alternatives    []Alternative `json:"alternatives"`
	ComputedBMI     float64       `json:"computedBmi"`
}

func Analyze(in Intake) Response {
	var issues []Issue
	riskScore := 1 // start with a small baseline

	bmi := in.BMI
	if bmi == 0 {
		bmi = computeBMI(in.WeightKg, in.HeightCm)
	}

	if bmi >= 30 {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "bmi",
			Severity:    "warning",
			Description: fmt.Sprintf("BMI %.1f indicates obesity; consider dose adjustments and monitor cardiovascular risk.", bmi),
		})
	} else if bmi >= 27 {
		riskScore++
		issues = append(issues, Issue{
			Type:        "bmi",
			Severity:    "info",
			Description: fmt.Sprintf("BMI %.1f is elevated; encourage lifestyle optimization alongside therapy.", bmi),
		})
	}

	systolic, diastolic := parseBP(in.BP)
	if systolic >= 160 || diastolic >= 100 {
		riskScore += 3
		issues = append(issues, Issue{
			Type:        "blood_pressure",
			Severity:    "danger",
			Description: fmt.Sprintf("Blood pressure %s suggests uncontrolled hypertension. Optimize BP before initiating risk-increasing meds.", in.BP),
		})
	} else if systolic >= 140 || diastolic >= 90 {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "blood_pressure",
			Severity:    "warning",
			Description: fmt.Sprintf("Blood pressure %s is elevated; monitor closely when adjusting vasoactive medications.", in.BP),
		})
	}

	cond := toSet(in.Conditions)
	if cond["heart disease"] {
		riskScore += 3
		issues = append(issues, Issue{
			Type:        "cardiac_history",
			Severity:    "danger",
			Description: "History of heart disease—ensure cardiac clearance before vasoactive or androgen-modifying therapy.",
		})
	}
	if cond["kidney disease"] {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "renal_impairment",
			Severity:    "warning",
			Description: "Kidney disease—prefer conservative dosing and avoid nephrotoxic combinations.",
		})
	}
	if cond["liver disease"] {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "hepatic_impairment",
			Severity:    "warning",
			Description: "Liver disease—consider lower starting doses and monitor LFTs where applicable.",
		})
	}
	if cond["diabetes"] {
		riskScore++
		issues = append(issues, Issue{
			Type:        "metabolic_risk",
			Severity:    "info",
			Description: "Diabetes increases cardiovascular risk; reinforce glycemic and lifestyle control.",
		})
	}
	if cond["hypertension"] {
		riskScore++
	}

	if in.Age > 65 {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "age_related",
			Severity:    "info",
			Description: "Age >65—start low, go slow with vasoactive agents; monitor for orthostatic changes.",
		})
	} else if in.Age >= 55 {
		riskScore++
	}

	if strings.EqualFold(in.Smoking, "current") {
		riskScore++
		issues = append(issues, Issue{
			Type:        "lifestyle",
			Severity:    "info",
			Description: "Current smoker—encourage cessation; adds cardiovascular risk.",
		})
	}
	if strings.EqualFold(in.Alcohol, "Heavy") {
		riskScore++
		issues = append(issues, Issue{
			Type:        "alcohol",
			Severity:    "info",
			Description: "Heavy alcohol use—counsel moderation; may worsen BP and medication tolerance.",
		})
	}

	meds := normalizeMeds(in.Medications)
	hasNitrate := meds["nitroglycerin"] || meds["isosorbide"] || containsAnyMedication(meds, []string{"nitrate"})
	if hasNitrate {
		riskScore += 5
		issues = append(issues, Issue{
			Type:        "contraindication",
			Severity:    "danger",
			Description: "Nitrate therapy—PDE5 inhibitors are contraindicated. Avoid tadalafil/sildenafil and coordinate cardiology care.",
		})
	}

	plan, alts := buildPlan(in, buildPlanContext{
		BMI:        bmi,
		HasNitrate: hasNitrate,
		HasHeartDz: cond["heart disease"],
		HasRenal:   cond["kidney disease"],
		HasHepatic: cond["liver disease"],
	})

	if usesPDE5(plan.Medication) && meds["amlodipine"] {
		riskScore++
		issues = append(issues, Issue{
			Type:        "drug_interaction",
			Severity:    "warning",
			Description: "PDE5 inhibitor may enhance the hypotensive effect of amlodipine. Monitor BP closely during initiation.",
		})
	}

	if usesPDE5(plan.Medication) && cond["heart disease"] {
		issues = append(issues, Issue{
			Type:        "cardiac_clearance",
			Severity:    "warning",
			Description: "Cardiac history—confirm patient is cleared for sexual activity before PDE5 use.",
		})
	}

	riskLevel := classifyRisk(riskScore)

	return Response{
		RiskLevel:       riskLevel,
		RiskScore:       riskScore,
		FlaggedIssues:   issues,
		RecommendedPlan: plan,
		Alternatives:    alts,
		ComputedBMI:     bmi,
	}
}

type buildPlanContext struct {
	BMI        float64
	HasNitrate bool
	HasHeartDz bool
	HasRenal   bool
	HasHepatic bool
}

func buildPlan(in Intake, ctx buildPlanContext) (Plan, []Alternative) {
	switch strings.ToLower(in.Complaint) {
	case "ed":
		return edPlan(ctx)
	case "hair loss":
		return hairLossPlan()
	case "weight loss":
		return weightLossPlan(ctx)
	default:
		return generalWellnessPlan()
	}
}

func edPlan(ctx buildPlanContext) (Plan, []Alternative) {
	if ctx.HasNitrate {
		return Plan{
				Medication: "Hold PDE5 inhibitors",
				Dosage:     "N/A",
				Frequency:  "Avoid until nitrates stopped",
				Duration:   "Reassess after nitrate-free period",
				Rationale:  "Nitrate therapy makes PDE5 inhibitors unsafe. Prioritize cardiology review and lifestyle optimization for ED.",
			}, []Alternative{
				{
					Medication: "Lifestyle & psychosexual therapy",
					Dosage:     "N/A",
					Pros:       []string{"No hemodynamic risk", "Addresses vascular + psychogenic factors"},
					Cons:       []string{"Slower onset of benefit"},
				},
				{
					Medication: "Vacuum erection device",
					Dosage:     "Device-assisted",
					Pros:       []string{"Non-pharmacologic", "No drug interactions"},
					Cons:       []string{"Less spontaneity", "Training required"},
				},
			}
	}

	dose := "10mg"
	if ctx.HasRenal || ctx.HasHepatic {
		dose = "5mg (start low due to renal/hepatic risk)"
	}
	rationale := "First-line PDE5 inhibitor; long half-life for flexibility. Start low to minimize hypotension risk; reinforce BP monitoring."
	if ctx.HasHeartDz {
		rationale += " Cardiac history—ensure clearance before sexual activity."
	}
	if ctx.BMI >= 27 {
		rationale += " Encourage weight and activity changes to improve ED and cardiometabolic profile."
	}

	return Plan{
			Medication: "Tadalafil",
			Dosage:     dose,
			Frequency:  "As needed, 30-60 minutes before sexual activity",
			Duration:   "30-day supply, renew after follow-up",
			Rationale:  rationale,
		}, []Alternative{
			{
				Medication: "Sildenafil",
				Dosage:     "50mg as needed (25mg if sensitive)",
				Pros:       []string{"Lower cost", "Shorter duration if side effects occur"},
				Cons:       []string{"Shorter window (4-6h)", "Requires timing around meals"},
			},
			{
				Medication: "Tadalafil (daily)",
				Dosage:     "5mg once daily",
				Pros:       []string{"Continuous effect", "Supports spontaneity", "May aid urinary symptoms"},
				Cons:       []string{"Daily commitment", "Higher cumulative cost"},
			},
		}
}

func hairLossPlan() (Plan, []Alternative) {
	return Plan{
			Medication: "Finasteride",
			Dosage:     "1mg orally once daily",
			Frequency:  "Daily",
			Duration:   "3-6 months before full effect",
			Rationale:  "DHT blocker with best evidence for male pattern hair loss. Monitor for sexual side effects; avoid if trying to conceive.",
		}, []Alternative{
			{
				Medication: "Topical Minoxidil 5%",
				Dosage:     "Apply to scalp twice daily",
				Pros:       []string{"OTC", "Safe for many patients"},
				Cons:       []string{"Requires adherence", "Shedding may transiently increase"},
			},
			{
				Medication: "Low-level laser therapy",
				Dosage:     "Per device guidance",
				Pros:       []string{"Non-drug option"},
				Cons:       []string{"Variable evidence", "Cost"},
			},
		}
}

func weightLossPlan(ctx buildPlanContext) (Plan, []Alternative) {
	rationale := "Calorie deficit with structured activity. Metformin aids insulin sensitivity; start low to reduce GI effects."
	if ctx.BMI >= 35 {
		rationale += " Consider GLP-1 RA if no contraindications and coverage allows."
	}

	return Plan{
			Medication: "Metformin",
			Dosage:     "500mg with dinner, uptitrate as tolerated",
			Frequency:  "Once daily start; can increase to BID",
			Duration:   "12-week trial with reassessment",
			Rationale:  rationale,
		}, []Alternative{
			{
				Medication: "GLP-1 receptor agonist",
				Dosage:     "Per product labeling (e.g., weekly titration)",
				Pros:       []string{"Robust weight loss", "Cardiometabolic benefit"},
				Cons:       []string{"Cost/coverage", "GI side effects", "Avoid in medullary thyroid cancer history"},
			},
			{
				Medication: "Intensive lifestyle program",
				Dosage:     "Nutrition + activity + sleep plan",
				Pros:       []string{"Foundational", "No drug interactions"},
				Cons:       []string{"Requires adherence", "Slower results"},
			},
		}
}

func generalWellnessPlan() (Plan, []Alternative) {
	return Plan{
			Medication: "Preventive care focus",
			Dosage:     "N/A",
			Frequency:  "Per guideline schedule",
			Duration:   "Ongoing",
			Rationale:  "No specific complaint provided. Recommend preventive screening, lifestyle optimization, and targeted labs based on history.",
		}, []Alternative{
			{
				Medication: "Lifestyle coaching",
				Dosage:     "Weekly sessions",
				Pros:       []string{"Addresses root causes", "No drug risk"},
				Cons:       []string{"Requires patient engagement"},
			},
		}
}

func classifyRisk(score int) string {
	switch {
	case score >= 8:
		return "HIGH"
	case score >= 4:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func computeBMI(weightKg, heightCm float64) float64 {
	if weightKg <= 0 || heightCm <= 0 {
		return 0
	}
	m := heightCm / 100.0
	return weightKg / (m * m)
}

var bpPattern = regexp.MustCompile(`(?i)(\d{2,3})\s*/\s*(\d{2,3})`)

func parseBP(bp string) (int, int) {
	m := bpPattern.FindStringSubmatch(bp)
	if len(m) != 3 {
		return 0, 0
	}
	s, _ := strconv.Atoi(m[1])
	d, _ := strconv.Atoi(m[2])
	return s, d
}

func toSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, v := range values {
		key := strings.ToLower(strings.TrimSpace(v))
		if key != "" {
			out[key] = true
		}
	}
	return out
}

func normalizeMeds(meds []Medication) map[string]bool {
	out := make(map[string]bool, len(meds))
	for _, m := range meds {
		name := strings.ToLower(strings.TrimSpace(m.Name))
		if name != "" {
			out[name] = true
		}
	}
	return out
}

func usesPDE5(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "tadalafil") || strings.Contains(n, "sildenafil") || strings.Contains(n, "vardenafil")
}

func containsAnyMedication(meds map[string]bool, needles []string) bool {
	for med := range meds {
		for _, n := range needles {
			if strings.Contains(med, n) {
				return true
			}
		}
	}
	return false
}
