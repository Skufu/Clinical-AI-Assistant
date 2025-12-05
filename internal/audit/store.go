package audit

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Entry captures an audit event for a clinical analysis or approval.
type Entry struct {
	ID         string
	PatientRef string
	Complaint  string
	RiskLevel  string
	RiskScore  int
	UserID     string
	At         time.Time
}

// Summary is a read-friendly view of an audit record.
type Summary struct {
	AuditID    string `json:"auditId"`
	PatientRef string `json:"patientRef"`
	Complaint  string `json:"complaint"`
	RiskLevel  string `json:"riskLevel"`
	RiskScore  int    `json:"riskScore"`
	UserID     string `json:"userId,omitempty"`
	At         string `json:"at"`
}

type Store interface {
	Insert(entry Entry) (Summary, error)
	Latest(limit int) ([]Summary, error)
}

const maxLimit = 50

// SQLiteStore is a simple SQLite-backed store; safe for concurrent use.
type SQLiteStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS audits (
			id TEXT PRIMARY KEY,
			patient_ref TEXT,
			complaint TEXT,
			risk_level TEXT,
			risk_score INTEGER,
			user_id TEXT,
			at_utc TEXT
		);
	`); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Insert(entry Entry) (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := entry.At
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id := entry.ID
	if id == "" {
		id = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	_, err := s.db.Exec(`
		INSERT INTO audits (id, patient_ref, complaint, risk_level, risk_score, user_id, at_utc)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, entry.PatientRef, entry.Complaint, entry.RiskLevel, entry.RiskScore, entry.UserID, now.Format(time.RFC3339))
	if err != nil {
		return Summary{}, fmt.Errorf("insert audit: %w", err)
	}
	return Summary{
		AuditID:    id,
		PatientRef: entry.PatientRef,
		Complaint:  entry.Complaint,
		RiskLevel:  entry.RiskLevel,
		RiskScore:  entry.RiskScore,
		UserID:     entry.UserID,
		At:         now.Format(time.RFC3339),
	}, nil
}

func (s *SQLiteStore) Latest(limit int) ([]Summary, error) {
	if limit <= 0 || limit > maxLimit {
		limit = 10
	}
	rows, err := s.db.Query(`
		SELECT id, patient_ref, complaint, risk_level, risk_score, user_id, at_utc
		FROM audits
		ORDER BY at_utc DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}
	defer rows.Close()

	var out []Summary
	for rows.Next() {
		var sEntry Summary
		if err := rows.Scan(&sEntry.AuditID, &sEntry.PatientRef, &sEntry.Complaint, &sEntry.RiskLevel, &sEntry.RiskScore, &sEntry.UserID, &sEntry.At); err != nil {
			return nil, fmt.Errorf("scan audit: %w", err)
		}
		out = append(out, sEntry)
	}
	return out, nil
}

// MemoryStore is a lightweight fallback for tests and offline use.
type MemoryStore struct {
	mu      sync.Mutex
	entries []Summary
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: []Summary{}}
}

func (m *MemoryStore) Insert(entry Entry) (Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := entry.At
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id := entry.ID
	if id == "" {
		id = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	sum := Summary{
		AuditID:    id,
		PatientRef: entry.PatientRef,
		Complaint:  entry.Complaint,
		RiskLevel:  entry.RiskLevel,
		RiskScore:  entry.RiskScore,
		UserID:     entry.UserID,
		At:         now.Format(time.RFC3339),
	}

	m.entries = append(m.entries, sum)
	if len(m.entries) > maxLimit {
		m.entries = m.entries[len(m.entries)-maxLimit:]
	}
	return sum, nil
}

func (m *MemoryStore) Latest(limit int) ([]Summary, error) {
	if limit <= 0 || limit > maxLimit {
		limit = 10
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	n := len(m.entries)
	start := n - limit
	if start < 0 {
		start = 0
	}
	out := make([]Summary, 0, n-start)
	out = append(out, m.entries[start:]...)
	return out, nil
}
