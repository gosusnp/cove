// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
)

// -----------------------------------------------------------------------------
// Workout Session
// -----------------------------------------------------------------------------

type WorkoutSessionID IntID[struct{ workoutSession struct{} }]

// SessionSensitiveData holds the sensitive fields of a workout session that are
// stored encrypted in the database. It is only decrypted by the handler layer.
//
// String fields use *crypto.SensitiveString instead of *string so that the
// backing bytes can be explicitly zeroed after use, avoiding Go string interning.
type SessionSensitiveData struct {
	PerceivedEffort  *int                    `json:"perceived_effort,omitempty"`
	SessionNotes     *crypto.SensitiveString `json:"session_notes,omitempty"`
	ProgramName      *crypto.SensitiveString `json:"program_name,omitempty"`
	ProgramStructure *crypto.SensitiveString `json:"program_structure,omitempty"`
	Summary          *crypto.SensitiveString `json:"summary,omitempty"`
}

// Format implements fmt.Formatter. All verbs emit "[REDACTED]" to prevent
// sensitive fields from appearing in logs or error messages.
func (s SessionSensitiveData) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, "SessionSensitiveData[REDACTED]")
}

// GoString implements fmt.GoStringer so that %#v also emits "SessionSensitiveData[REDACTED]".
func (s SessionSensitiveData) GoString() string {
	return "SessionSensitiveData[REDACTED]"
}

// IsEmpty reports whether all sensitive fields are unset. When true, the
// service skips encryption and stores NULL in the database instead.
// Update this method whenever a new field is added to SessionSensitiveData.
func (s SessionSensitiveData) IsEmpty() bool {
	return s.PerceivedEffort == nil &&
		s.SessionNotes == nil &&
		s.ProgramName == nil &&
		s.ProgramStructure == nil &&
		s.Summary == nil
}

// WorkoutSession represents a single training session.
// Sensitive data is accessed exclusively via UseSensitiveData.
type WorkoutSession struct {
	ID                 WorkoutSessionID `json:"id"`
	OrgID              OrgID            `json:"org_id"`
	UserID             UserID           `json:"user_id"`
	ProgramID          *ProgramID       `json:"program_id,omitempty"`
	Activity           *string          `json:"activity,omitempty"`
	DurationS          *int             `json:"duration_s,omitempty"`
	StartedAt          *time.Time       `json:"started_at,omitempty"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty"`
	SummaryGeneratedAt *time.Time       `json:"summary_generated_at,omitempty"`
	CreatedBy          UserID           `json:"created_by"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedBy          *UserID          `json:"updated_by,omitempty"`
	UpdatedAt          time.Time        `json:"updated_at"`

	// sensitiveData contains a sync.Mutex; WorkoutSession must always be used
	// as a pointer (*WorkoutSession) and must never be copied after first use.
	sensitiveData crypto.EncryptedField[SessionSensitiveData] `json:"-"`
}

// UseSensitiveData decrypts the session's sensitive payload, calls fn with it,
// then zeros the struct in place before returning. It is the only way to access
// sensitive fields — no plaintext escapes this call.
//
// Because string fields use crypto.SensitiveString ([]byte-backed), zeroing the
// struct overwrites their backing bytes, avoiding Go string interning.
//
// The session's UserID is passed as GCM additional data, binding the ciphertext
// to this specific user. Decryption will fail if the stored user_id does not
// match.
func (ws *WorkoutSession) UseSensitiveData(ctx context.Context, fn func(SessionSensitiveData) error) error {
	return ws.sensitiveData.Use(ctx, func(p *SessionSensitiveData) error {
		return fn(*p)
	}, ws.UserID.UUID[:])
}

// SetEncryptor injects the encryptor needed to decrypt sensitive data.
// Called by the service after a store read.
func (ws *WorkoutSession) SetEncryptor(enc crypto.Encryptor) {
	ws.sensitiveData.SetEncryptor(enc)
}

// SensitiveDataScanner returns a pointer suitable for sql.Scan to populate the
// sensitive_data column. Called by the store only.
func (ws *WorkoutSession) SensitiveDataScanner() interface{ Scan(any) error } {
	return &ws.sensitiveData
}

// SensitiveDataBytes returns the raw ciphertext for writing to the store.
// Called by the store only.
func (ws *WorkoutSession) SensitiveDataBytes() []byte {
	return ws.sensitiveData.Value()
}

// -----------------------------------------------------------------------------
// Program
// -----------------------------------------------------------------------------

type ProgramID IntID[struct{ program struct{} }]
type ProgramVersionID IntID[struct{ programVersion struct{} }]

// ProgramLite is a trimmed version of a program.
type ProgramLite struct {
	ID       ProgramID `json:"id"`
	Name     string    `json:"name"`
	Activity *string   `json:"activity,omitempty"`
	OrgID    OrgID     `json:"org_id"`
	IsPublic bool      `json:"is_public"`
}

// Program is the complete program hierarchy.
type Program struct {
	ID          ProgramID    `json:"id"`
	Name        string       `json:"name"`
	Activity    *string      `json:"activity,omitempty"`
	Description *string      `json:"description,omitempty"`
	OrgID       OrgID        `json:"org_id"`
	IsPublic    bool         `json:"is_public"`
	CreatedBy   UserID       `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedBy   *UserID      `json:"updated_by,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Sets        []ProgramSet `json:"sets"`
}

// ProgramVersionMeta is the trimmed projection returned by list endpoints (no snapshot).
type ProgramVersionMeta struct {
	ID        ProgramVersionID `json:"id"`
	ProgramID ProgramID        `json:"program_id"`
	OrgID     OrgID            `json:"org_id"`
	CreatedBy UserID           `json:"created_by"`
	CreatedAt time.Time        `json:"created_at"`
}

// ProgramVersion represents a historical snapshot of a program.
type ProgramVersion struct {
	ID        ProgramVersionID `json:"id"`
	ProgramID ProgramID        `json:"program_id"`
	OrgID     OrgID            `json:"org_id"`
	Snapshot  ProgramSnapshot  `json:"snapshot"`
	CreatedBy UserID           `json:"created_by"`
	CreatedAt time.Time        `json:"created_at"`
}

// ProgramSnapshot captures the user-facing state of a program.
type ProgramSnapshot struct {
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	Activity    *string      `json:"activity,omitempty"`
	IsPublic    bool         `json:"is_public"`
	Sets        []ProgramSet `json:"sets"`
}

type ProgramSet struct {
	ID                  int64             `json:"id"`
	Name                *string           `json:"name,omitempty"`
	Rounds              int               `json:"rounds"`
	IntraSetRestSeconds *int              `json:"rest_s,omitempty"`
	Exercises           []ProgramExercise `json:"exercises"`
}

type ProgramExercise struct {
	ID                    int64      `json:"id"`
	ExerciseID            ExerciseID `json:"exercise_id"`
	Name                  string     `json:"name"`
	Laterality            *string    `json:"laterality,omitempty"`
	TargetReps            *int       `json:"reps,omitempty"`
	TargetDurationSeconds *int       `json:"duration_s,omitempty"`
	TargetWeight          *float64   `json:"weight,omitempty"`
	WeightUnit            *Unit      `json:"weight_unit,omitempty"`
}

// -----------------------------------------------------------------------------
// Exercise
// -----------------------------------------------------------------------------

type ExerciseID IntID[struct{ exercise struct{} }]

// Exercise is the complete exercise definition.
type Exercise struct {
	ID          ExerciseID `json:"id"`
	Name        string     `json:"name"`
	Progression *string    `json:"progression,omitempty"`
	Description *string    `json:"description,omitempty"`
	OrgID       OrgID      `json:"org_id"`
	IsPublic    bool       `json:"is_public"`
	CreatedBy   UserID     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedBy   *UserID    `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ExerciseLite is a trimmed version of an exercise.
type ExerciseLite struct {
	ID   ExerciseID `json:"id"`
	Name string     `json:"name"`
}

// -----------------------------------------------------------------------------
// User Training Profile
// -----------------------------------------------------------------------------

// TrainingProfileSensitiveData holds the sensitive training profile information.
type TrainingProfileSensitiveData struct {
	Motivation  *crypto.SensitiveString     `json:"motivation,omitempty"`
	Disciplines []TrainingProfileDiscipline `json:"disciplines,omitempty"`
	Constraints *crypto.SensitiveString     `json:"constraints,omitempty"`
}

type TrainingProfileDiscipline struct {
	Name          *crypto.SensitiveString `json:"name,omitempty"`
	YearsPractice *float64                `json:"years_practice,omitempty"`
	Level         *crypto.SensitiveString `json:"level,omitempty"`
	Notes         *crypto.SensitiveString `json:"notes,omitempty"`
}

// Format implements fmt.Formatter.
func (s TrainingProfileSensitiveData) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, "TrainingProfileSensitiveData[REDACTED]")
}

// GoString implements fmt.GoStringer.
func (s TrainingProfileSensitiveData) GoString() string {
	return "TrainingProfileSensitiveData[REDACTED]"
}

// IsEmpty reports whether all sensitive fields are unset.
func (s TrainingProfileSensitiveData) IsEmpty() bool {
	return s.Motivation == nil && len(s.Disciplines) == 0 && s.Constraints == nil
}

// UserTrainingProfile represents a user's training goals and history.
type UserTrainingProfile struct {
	UserID    UserID    `json:"user_id"`
	OrgID     OrgID     `json:"org_id"`
	CreatedBy UserID    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedBy *UserID   `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`

	sensitiveData crypto.EncryptedField[TrainingProfileSensitiveData] `json:"-"`
}

// UseSensitiveData decrypts the profile's sensitive payload, calls fn with it,
// then zeros the struct in place before returning.
func (tp *UserTrainingProfile) UseSensitiveData(ctx context.Context, fn func(TrainingProfileSensitiveData) error) error {
	return tp.sensitiveData.Use(ctx, func(p *TrainingProfileSensitiveData) error {
		return fn(*p)
	}, tp.UserID.UUID[:])
}

// SetEncryptor injects the encryptor needed to decrypt sensitive data.
func (tp *UserTrainingProfile) SetEncryptor(enc crypto.Encryptor) {
	tp.sensitiveData.SetEncryptor(enc)
}

// SensitiveDataScanner returns a pointer suitable for sql.Scan.
func (tp *UserTrainingProfile) SensitiveDataScanner() interface{ Scan(any) error } {
	return &tp.sensitiveData
}

// SensitiveDataBytes returns the raw ciphertext for writing to the store.
func (tp *UserTrainingProfile) SensitiveDataBytes() []byte {
	return tp.sensitiveData.Value()
}
