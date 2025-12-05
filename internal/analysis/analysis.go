package analysis

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"
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
	Confidence float64  `json:"confidence,omitempty"`
}

type Response struct {
	RiskLevel        string        `json:"riskLevel"`
	RiskScore        int           `json:"riskScore"`
	FlaggedIssues    []Issue       `json:"flaggedIssues"`
	RecommendedPlan  Plan          `json:"recommendedPlan"`
	PlanConfidence   float64       `json:"planConfidence,omitempty"`
	Alternatives     []Alternative `json:"alternatives"`
	ComputedBMI      float64       `json:"computedBmi"`
	ValidationErrors []string      `json:"validationErrors,omitempty"`
	AuditID          string        `json:"auditId,omitempty"`
	AuditAt          string        `json:"auditAt,omitempty"`
}

//go:embed schema/response.schema.json
var responseSchema []byte

var systemPrompt = `
You are a clinical decision support assistant. Apply conservative, guideline-informed rules:
- Flag contraindications: nitrates + PDE5 inhibitors, uncontrolled hypertension (>160/100), severe hepatic/renal disease with dose adjustments, cardiac clearance for sexual activity in CAD/heart disease.
- Flag interactions: amlodipine + PDE5 (hypotension), tamsulosin + PDE5 (hypotension), alcohol + PDE5 (hypotension/dizziness).
- Check dosing: PDE5 starting doses 5-10mg (tadalafil) or 25-50mg (sildenafil); warn >20mg tadalafil single dose.
- Consider comorbidities: BMI >27 elevated risk; BMI >=30 obesity. Diabetes, hypertension, heart/kidney/liver disease increase risk.
- Always include rationale and alternatives with pros/cons and confidence 0-1.
- Safety > everything: prefer flagging potential risks.
Return structured JSON per schema: riskLevel, riskScore, flaggedIssues, recommendedPlan, planConfidence, alternatives, computedBmi, auditId, validationErrors (if any).
`

func Analyze(in Intake) Response {
	if errs := Validate(in); len(errs) > 0 {
		return Response{
			RiskLevel:        "INVALID",
			RiskScore:        0,
			FlaggedIssues:    nil,
			RecommendedPlan:  Plan{},
			Alternatives:     nil,
			ComputedBMI:      0,
			ValidationErrors: errs,
		}
	}

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

	if usesPDE5(plan.Medication) && meds["tamsulosin"] {
		riskScore++
		issues = append(issues, Issue{
			Type:        "drug_interaction",
			Severity:    "warning",
			Description: "PDE5 inhibitor plus tamsulosin may increase hypotension risk. Consider spacing doses and monitoring.",
		})
	}

	if usesPDE5(plan.Medication) && cond["heart disease"] {
		issues = append(issues, Issue{
			Type:        "cardiac_clearance",
			Severity:    "warning",
			Description: "Cardiac history—confirm patient is cleared for sexual activity before PDE5 use.",
		})
	}

	if usesPDE5(plan.Medication) && strings.EqualFold(in.Alcohol, "heavy") {
		issues = append(issues, Issue{
			Type:        "alcohol",
			Severity:    "info",
			Description: "Heavy alcohol use with PDE5 inhibitors can worsen hypotension and dizziness. Counsel moderation.",
		})
	}

	// Additional interaction datasource checks (local ruleset).
	issues = append(issues, interactionIssues(meds)...)

	// Allergy cross-checks against plan and alternatives.
	if allergy := intersectsAllergy(in.Allergies, plan.Medication); allergy != "" {
		riskScore += 3
		issues = append(issues, Issue{
			Type:        "allergy",
			Severity:    "danger",
			Description: fmt.Sprintf("Allergy match detected for planned medication (%s).", allergy),
		})
	}

	for _, alt := range alts {
		if allergy := intersectsAllergy(in.Allergies, alt.Medication); allergy != "" {
			issues = append(issues, Issue{
				Type:        "allergy",
				Severity:    "warning",
				Description: fmt.Sprintf("Alternative %s conflicts with allergy (%s).", alt.Medication, allergy),
			})
		}
	}

	if exceedsDose(plan.Medication, plan.Dosage) {
		riskScore += 2
		issues = append(issues, Issue{
			Type:        "dose_cap",
			Severity:    "warning",
			Description: fmt.Sprintf("Dosage %s for %s may exceed common starting caps. Consider reducing.", plan.Dosage, plan.Medication),
		})
	}

	riskLevel := classifyRisk(riskScore)

	llm := callLLMStub(in, plan, alts)
	planConfidence := llm.PlanConfidence
	alts = mergeAltConfidence(alts, llm.AlternativeConf)

	if issues == nil {
		issues = []Issue{}
	}
	if alts == nil {
		alts = []Alternative{}
	}

	auditID, auditAt := recordAudit(in, riskLevel, riskScore)

	resp := Response{
		RiskLevel:       riskLevel,
		RiskScore:       riskScore,
		FlaggedIssues:   issues,
		RecommendedPlan: plan,
		PlanConfidence:  planConfidence,
		Alternatives:    alts,
		ComputedBMI:     bmi,
		AuditID:         auditID,
		AuditAt:         auditAt,
	}

	if verrs := ValidateResponse(resp); len(verrs) > 0 {
		resp.ValidationErrors = append(resp.ValidationErrors, verrs...)
	}

	return resp
}

type llmResult struct {
	PlanConfidence  float64
	AlternativeConf []float64
}

// callLLMStub simulates an LLM scoring step while keeping deterministic guardrails.
func callLLMStub(in Intake, plan Plan, alts []Alternative) llmResult {
	// Simple heuristic confidence based on risk and completeness of intake.
	coverage := 0.6
	if in.BP != "" {
		coverage += 0.05
	}
	if len(in.Conditions) > 0 {
		coverage += 0.05
	}
	if len(in.Medications) > 0 {
		coverage += 0.05
	}
	if in.Allergies != nil {
		coverage += 0.05
	}

	planConfidence := clamp(0.55+coverage*0.3, 0, 0.95)
	altConf := make([]float64, len(alts))
	for i := range alts {
		altConf[i] = clamp(planConfidence-0.05*float64(i+1), 0.4, 0.9)
	}
	return llmResult{
		PlanConfidence:  planConfidence,
		AlternativeConf: altConf,
	}
}

func mergeAltConfidence(alts []Alternative, conf []float64) []Alternative {
	for i := range alts {
		if i < len(conf) {
			alts[i].Confidence = conf[i]
		}
	}
	return alts
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

func exceedsDose(medication, dose string) bool {
	// Simple guard: flag PDE5 doses >20mg.
	if !usesPDE5(medication) {
		return false
	}
	num := extractMg(dose)
	return num > 20
}

func extractMg(dose string) float64 {
	re := regexp.MustCompile(`([\d.]+)\s*mg`)
	m := re.FindStringSubmatch(strings.ToLower(dose))
	if len(m) < 2 {
		return 0
	}
	val, _ := strconv.ParseFloat(m[1], 64)
	return val
}

func intersectsAllergy(allergies []string, medication string) string {
	med := strings.ToLower(medication)
	for _, a := range allergies {
		if strings.Contains(med, strings.ToLower(strings.TrimSpace(a))) && strings.TrimSpace(a) != "" {
			return strings.TrimSpace(a)
		}
	}
	return ""
}

// Validate performs basic intake validation before deeper analysis.
func Validate(in Intake) []string {
	var errs []string
	if strings.TrimSpace(in.PatientName) == "" {
		errs = append(errs, "patientName is required")
	}
	if in.Age <= 0 {
		errs = append(errs, "age must be greater than 0")
	}
	if in.WeightKg <= 0 {
		errs = append(errs, "weight must be greater than 0")
	}
	if in.HeightCm <= 0 {
		errs = append(errs, "height must be greater than 0")
	}
	if strings.TrimSpace(in.BP) == "" {
		errs = append(errs, "bp is required")
	}
	if strings.TrimSpace(in.Complaint) == "" {
		errs = append(errs, "complaint is required")
	}
	return errs
}

// ValidateResponse ensures responses conform to schema before returning.
func ValidateResponse(resp Response) []string {
	body, err := json.Marshal(resp)
	if err != nil {
		return []string{"failed to marshal response"}
	}
	schemaLoader := gojsonschema.NewBytesLoader(responseSchema)
	docLoader := gojsonschema.NewBytesLoader(body)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return []string{"schema validation error: " + err.Error()}
	}
	if result.Valid() {
		return nil
	}
	out := make([]string, 0, len(result.Errors()))
	for _, e := range result.Errors() {
		out = append(out, e.String())
	}
	return out
}

type auditEntry struct {
	ID         string
	PatientRef string
	Complaint  string
	RiskLevel  string
	RiskScore  int
	At         time.Time
}

var auditLog []auditEntry

const auditLimit = 50

func recordAudit(in Intake, risk string, score int) (string, string) {
	id := fmt.Sprintf("audit-%d", time.Now().UnixNano())
	ref := strings.TrimSpace(in.PatientName)
	if len(ref) > 2 {
		ref = ref[:1] + "***"
	}
	entry := auditEntry{
		ID:         id,
		PatientRef: ref,
		Complaint:  in.Complaint,
		RiskLevel:  risk,
		RiskScore:  score,
		At:         time.Now(),
	}
	auditLog = append(auditLog, entry)
	if len(auditLog) > auditLimit {
		auditLog = auditLog[len(auditLog)-auditLimit:]
	}
	return id, entry.At.UTC().Format(time.RFC3339)
}

type AuditSummary struct {
	AuditID    string `json:"auditId"`
	PatientRef string `json:"patientRef"`
	Complaint  string `json:"complaint"`
	RiskLevel  string `json:"riskLevel"`
	RiskScore  int    `json:"riskScore"`
	At         string `json:"at"`
}

func LatestAudits(limit int) []AuditSummary {
	if limit <= 0 || limit > auditLimit {
		limit = 10
	}
	n := len(auditLog)
	start := n - limit
	if start < 0 {
		start = 0
	}
	out := make([]AuditSummary, 0, n-start)
	for _, a := range auditLog[start:] {
		out = append(out, AuditSummary{
			AuditID:    a.ID,
			PatientRef: a.PatientRef,
			Complaint:  a.Complaint,
			RiskLevel:  a.RiskLevel,
			RiskScore:  a.RiskScore,
			At:         a.At.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

type interactionRule struct {
	Drug      string
	With      string
	Severity  string
	Desc      string
	RiskDelta int
}

var interactionRules = []interactionRule{
	{
		Drug:      "amlodipine",
		With:      "simvastatin",
		Severity:  "warning",
		Desc:      "Amlodipine can raise simvastatin levels; consider limiting simvastatin to 20mg/day.",
		RiskDelta: 1,
	},
	{
		Drug:      "metformin",
		With:      "contrast",
		Severity:  "info",
		Desc:      "Hold metformin around iodinated contrast if eGFR is low to reduce lactic acidosis risk.",
		RiskDelta: 0,
	},
	{
		Drug:      "finasteride",
		With:      "pregnancy",
		Severity:  "warning",
		Desc:      "Finasteride is teratogenic; avoid handling in pregnancy.",
		RiskDelta: 1,
	},
}

func interactionIssues(meds map[string]bool) []Issue {
	var out []Issue
	for _, rule := range interactionRules {
		if meds[rule.Drug] && meds[rule.With] {
			out = append(out, Issue{
				Type:        "drug_interaction",
				Severity:    rule.Severity,
				Description: rule.Desc,
			})
		}
	}
	return out
}
