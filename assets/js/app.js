const API_URL = '/api/analyze';
const analyzeBtn = document.getElementById('analyzeBtn');
const analyzeDefaultLabel = analyzeBtn.textContent;

document.getElementById('weight').addEventListener('input', calculateBMI);
document.getElementById('height').addEventListener('input', calculateBMI);

function calculateBMI() {
    const weight = parseFloat(document.getElementById('weight').value);
    const height = parseFloat(document.getElementById('height').value);
    if (weight && height) {
        const bmi = (weight / ((height / 100) ** 2)).toFixed(1);
        document.getElementById('bmi').value = bmi;
    }
}

function addMedication() {
    const container = document.getElementById('medications');
    const entry = document.createElement('div');
    entry.className = 'medication-entry';
    entry.innerHTML = `
        <input type="text" class="form-input med-name" placeholder="Drug name">
        <input type="text" class="form-input med-dose" placeholder="Dosage">
        <select class="form-input med-freq">
            <option value="Daily">Daily</option>
            <option value="Twice daily">Twice daily</option>
            <option value="As needed">As needed</option>
        </select>
    `;
    container.appendChild(entry);
}

function prefillSample() {
    document.getElementById('patientName').value = 'Juan Dela Cruz';
    document.getElementById('userId').value = 'demo-clinician';
    document.getElementById('age').value = '45';
    document.getElementById('weight').value = '78';
    document.getElementById('height').value = '175';
    document.getElementById('bp').value = '135/88';
    document.getElementById('allergies').value = 'None';
    document.getElementById('smoking').value = 'Former';
    document.getElementById('alcohol').value = 'Occasional';
    document.getElementById('exercise').value = '1-2x/week';
    document.getElementById('complaint').value = 'ED';
    
    document.querySelector('#conditions input[value="Hypertension"]').checked = true;
    
    const medEntry = document.querySelector('.medication-entry');
    medEntry.querySelector('.med-name').value = 'Amlodipine';
    medEntry.querySelector('.med-dose').value = '5mg';
    medEntry.querySelector('.med-freq').value = 'Daily';
    
    calculateBMI();
}

function prefillHighRisk() {
    document.getElementById('patientName').value = 'Maria Santos';
    document.getElementById('userId').value = 'demo-clinician';
    document.getElementById('age').value = '68';
    document.getElementById('weight').value = '90';
    document.getElementById('height').value = '165';
    document.getElementById('bp').value = '168/102';
    document.getElementById('allergies').value = 'Sulfa';
    document.getElementById('smoking').value = 'Current';
    document.getElementById('alcohol').value = 'Moderate';
    document.getElementById('exercise').value = 'None';
    document.getElementById('complaint').value = 'ED';
    
    document.querySelectorAll('#conditions input').forEach(cb => cb.checked = false);
    document.querySelector('#conditions input[value="Heart Disease"]').checked = true;
    document.querySelector('#conditions input[value="Hypertension"]').checked = true;
    
    const medEntry = document.querySelector('.medication-entry');
    medEntry.querySelector('.med-name').value = 'Nitroglycerin';
    medEntry.querySelector('.med-dose').value = '0.4mg';
    medEntry.querySelector('.med-freq').value = 'As needed';
    
    calculateBMI();
}

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
    document.getElementById(sectionId).classList.add('active');
    window.scrollTo(0, 0);
}

function clearErrors() {
    document.querySelectorAll('.error-text').forEach(el => el.textContent = '');
    document.querySelectorAll('.form-input').forEach(el => el.classList.remove('input-error'));
}

function setFieldError(fieldId, message) {
    const input = document.getElementById(fieldId);
    if (input) input.classList.add('input-error');
    const slot = document.querySelector(`[data-error-for="${fieldId}"]`);
    if (slot) slot.textContent = message;
}

function validateForm(data) {
    clearErrors();
    const errors = [];
    const bpPattern = /^\s*\d{2,3}\s*\/\s*\d{2,3}\s*$/;

    if (!data.patientName) {
        errors.push('Patient name is required.');
        setFieldError('patientName', 'Required');
    }
    if (!data.age || data.age <= 0) {
        errors.push('Age must be greater than 0.');
        setFieldError('age', 'Enter valid age');
    }
    if (!data.weight || data.weight <= 0) {
        errors.push('Weight must be greater than 0.');
        setFieldError('weight', 'Enter valid weight');
    }
    if (!data.height || data.height <= 0) {
        errors.push('Height must be greater than 0.');
        setFieldError('height', 'Enter valid height');
    }
    if (!data.bp || !bpPattern.test(data.bp)) {
        errors.push('Blood pressure must be in ##/## format.');
        setFieldError('bp', 'Format e.g. 120/80');
    }
    if (!data.complaint) {
        errors.push('Chief complaint is required.');
        setFieldError('complaint', 'Select a complaint');
    }

    return { valid: errors.length === 0, errors };
}

function getFormData() {
    const conditions = Array.from(document.querySelectorAll('#conditions input:checked')).map(cb => cb.value);
    const allergies = document.getElementById('allergies').value.split(',').map(a => a.trim()).filter(Boolean);
    const medications = Array.from(document.querySelectorAll('#medications .medication-entry'))
        .map(entry => ({
            name: entry.querySelector('.med-name').value.trim(),
            dosage: entry.querySelector('.med-dose').value.trim(),
            frequency: entry.querySelector('.med-freq').value
        }))
        .filter(m => m.name.length > 0);

    return {
        patientName: document.getElementById('patientName').value.trim(),
        userId: (document.getElementById('userId').value || 'demo-clinician').trim(),
        age: Number(document.getElementById('age').value) || 0,
        weight: parseFloat(document.getElementById('weight').value) || 0,
        height: parseFloat(document.getElementById('height').value) || 0,
        bp: document.getElementById('bp').value.trim(),
        bmi: parseFloat(document.getElementById('bmi').value) || 0,
        conditions,
        allergies,
        medications,
        smoking: document.getElementById('smoking').value,
        alcohol: document.getElementById('alcohol').value,
        exercise: document.getElementById('exercise').value,
        complaint: document.getElementById('complaint').value
    };
}

function analyzePatient() {
    const payload = getFormData();
    const validation = validateForm(payload);
    if (!validation.valid) {
        const firstError = document.querySelector('.input-error');
        if (firstError) firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
        return;
    }

    showSection('analysis');
    document.getElementById('loading').style.display = 'block';
    document.getElementById('results').style.display = 'none';
    document.getElementById('errorCard').style.display = 'none';
    setAnalyzing(true);

    fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    })
    .then(async resp => {
        const data = await resp.json().catch(() => null);
        if (!resp.ok) {
            const message = (data && data.error) ? data.error : `Server returned ${resp.status}`;
            const details = (data && data.details) ? data.details : [];
            throw { message, details };
        }
        return data;
    })
    .then(data => {
        document.getElementById('loading').style.display = 'none';
        renderResults(data);
        document.getElementById('results').style.display = 'block';
        if (data.computedBmi) {
            document.getElementById('bmi').value = data.computedBmi.toFixed(1);
        }
        setAnalyzing(false);
    })
    .catch(err => {
        document.getElementById('loading').style.display = 'none';
        const details = err && err.details ? err.details : [];
        const msg = err && err.message ? err.message : 'Unable to analyze patient.';
        showError(msg, details);
        setAnalyzing(false);
    });
}

function setAnalyzing(isLoading) {
    analyzeBtn.disabled = isLoading;
    analyzeBtn.textContent = isLoading ? 'Analyzing...' : analyzeDefaultLabel;
}

function showError(message, details = []) {
    const card = document.getElementById('errorCard');
    document.getElementById('errorText').textContent = message;
    const list = document.getElementById('errorList');
    list.innerHTML = details.map(d => `<li>${d}</li>`).join('');
    card.style.display = 'block';
    document.getElementById('results').style.display = 'none';
}

function renderResults(data) {
    const banner = document.getElementById('riskBanner');
    const riskClass = data.riskLevel.toLowerCase();
    banner.className = `risk-banner ${riskClass}`;
    
    const icons = { low: '🟢', medium: '🟠', high: '🔴' };
    const suitability = data.riskLevel === 'LOW'
        ? 'Patient suitable for treatment'
        : data.riskLevel === 'MEDIUM'
            ? 'Use caution and monitor closely'
            : 'High risk—optimize risks before initiating';
    banner.querySelector('.risk-icon').textContent = icons[riskClass];
    banner.querySelector('.risk-title').textContent = `${data.riskLevel} RISK`;
    banner.querySelector('.risk-score').textContent = `Risk Score: ${data.riskScore} • ${suitability}`;
    document.getElementById('auditMeta').textContent = data.auditId
        ? `Audit ID: ${data.auditId}${data.auditAt ? ' • ' + data.auditAt : ''}`
        : '';

    const issuesList = document.getElementById('issuesList');
    const issues = Array.isArray(data.flaggedIssues) ? data.flaggedIssues : [];
    if (issues.length === 0) {
        issuesList.innerHTML = '<p style="color: var(--color-success);">✓ No critical issues detected</p>';
    } else {
        issuesList.innerHTML = issues.map(issue => {
            const icons = { danger: '🔴', warning: '🟠', info: '🔵' };
            const labels = { danger: 'SEVERE', warning: 'MODERATE', info: 'INFO' };
            return `
                <div class="issue-item ${issue.severity}">
                    <div class="issue-icon">${icons[issue.severity]}</div>
                    <div>
                        <div class="issue-title">${labels[issue.severity]}: ${issue.type.replace('_', ' ').toUpperCase()}</div>
                        <div class="issue-desc">${issue.description}</div>
                    </div>
                </div>
            `;
        }).join('');
    }

    const plan = data.recommendedPlan || {};
    document.getElementById('treatmentPlan').innerHTML = `
        <div class="treatment-row">
            <span class="treatment-label">Medication</span>
            <span class="treatment-value">${plan.medication || '—'}</span>
        </div>
        <div class="treatment-row">
            <span class="treatment-label">Dosage</span>
            <span class="treatment-value">${plan.dosage || '—'}</span>
        </div>
        <div class="treatment-row">
            <span class="treatment-label">Frequency</span>
            <span class="treatment-value">${plan.frequency || '—'}</span>
        </div>
        <div class="treatment-row">
            <span class="treatment-label">Duration</span>
            <span class="treatment-value">${plan.duration || '—'}</span>
        </div>
    `;
    
    document.getElementById('rationale').innerHTML = `<strong>Clinical Rationale:</strong><br>${plan.rationale || '—'}<br><br><strong>Confidence:</strong> ${data.planConfidence ? (data.planConfidence * 100).toFixed(0) + '%' : '—'}`;

    const alternatives = Array.isArray(data.alternatives) ? data.alternatives : [];
    document.getElementById('alternatives').innerHTML = alternatives.length === 0
        ? '<p style="color: var(--color-text-secondary);">No alternatives provided.</p>'
        : alternatives.map(alt => `
            <div class="alternative-item">
                <div class="alt-name">${alt.medication}</div>
                <div class="alt-pros">✓ Pros: ${alt.pros.join(' • ')}</div>
                <div class="alt-cons">✗ Cons: ${alt.cons.join(' • ')}</div>
                <div style="margin-top: 6px; color: var(--color-text-secondary); font-size: 13px;">Confidence: ${alt.confidence ? (alt.confidence * 100).toFixed(0) + '%' : '—'}</div>
            </div>
        `).join('');

    window.__lastResults = data;
}

function approvePlan() {
    showSection('approved');
}

function goToReview() {
    if (!window.__lastResults) {
        showSection('analysis');
        return;
    }
    const plan = window.__lastResults.recommendedPlan || {};
    document.getElementById('reviewMedication').value = plan.medication || '';
    document.getElementById('reviewDosage').value = plan.dosage || '';
    document.getElementById('reviewFrequency').value = plan.frequency || '';
    document.getElementById('reviewDuration').value = plan.duration || '';
    document.getElementById('reviewRationale').value = plan.rationale || '';
    showSection('review');
}

function finalizeReview() {
    const summary = [
        `Medication: ${document.getElementById('reviewMedication').value || '—'}`,
        `Dosage: ${document.getElementById('reviewDosage').value || '—'}`,
        `Frequency: ${document.getElementById('reviewFrequency').value || '—'}`,
        `Duration: ${document.getElementById('reviewDuration').value || '—'}`,
        `Rationale: ${document.getElementById('reviewRationale').value || '—'}`,
        window.__lastResults && window.__lastResults.auditId ? `Audit ID: ${window.__lastResults.auditId}` : ''
    ].filter(Boolean).join(' • ');
    document.getElementById('approvalSummary').textContent = summary;
    showSection('approved');
}

function startOver() {
    document.getElementById('patientName').value = '';
    document.getElementById('userId').value = 'demo-clinician';
    document.getElementById('age').value = '';
    document.getElementById('weight').value = '';
    document.getElementById('height').value = '';
    document.getElementById('bp').value = '';
    document.getElementById('bmi').value = '';
    document.getElementById('allergies').value = '';
    document.querySelectorAll('#conditions input').forEach(cb => cb.checked = false);
    document.getElementById('medications').innerHTML = `
        <div class="medication-entry">
            <input type="text" class="form-input med-name" placeholder="Drug name">
            <input type="text" class="form-input med-dose" placeholder="Dosage">
            <select class="form-input med-freq">
                <option value="Daily">Daily</option>
                <option value="Twice daily">Twice daily</option>
                <option value="As needed">As needed</option>
            </select>
        </div>
    `;
    showSection('intake');
    clearErrors();
    setAnalyzing(false);
}
