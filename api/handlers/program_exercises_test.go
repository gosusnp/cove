package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosusnp/cove/api/service"
	"github.com/gosusnp/cove/api/store"
)

type programExerciseTestServer struct {
	mux        http.Handler
	svc        *service.ProgramExerciseService
	programID  int64
	setID      int64
	exerciseID int64
}

func newTestProgramExerciseServer(t *testing.T) programExerciseTestServer {
	t.Helper()
	db := newTestDB(t)

	p, err := store.NewProgramStore(db).Create("Test Program")
	if err != nil {
		t.Fatal(err)
	}
	ps, err := store.NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	e, err := store.NewExerciseStore(db).Create("Pull-up", nil)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewProgramExerciseService(store.NewProgramExerciseStore(db))
	mux := http.NewServeMux()
	NewProgramExerciseHandler(svc).RegisterRoutes(mux)

	return programExerciseTestServer{
		mux:        mux,
		svc:        svc,
		programID:  p.ID,
		setID:      ps.ID,
		exerciseID: e.ID,
	}
}

func (s programExerciseTestServer) url(path string) string {
	return fmt.Sprintf("/programs/%d/sets/%d/exercises%s", s.programID, s.setID, path)
}

func TestProgramExerciseHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		r := httptest.NewRequest(http.MethodGet, ts.url(""), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns exercises for set", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)
		if _, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, ts.url(""), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d exercises, want 2", len(got))
		}
	})
}

func TestProgramExerciseHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)
		created, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, ts.url(fmt.Sprintf("/%d", created.ID)), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ExerciseID != ts.exerciseID {
			t.Errorf("got exercise_id %d, want %d", got.ExerciseID, ts.exerciseID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		r := httptest.NewRequest(http.MethodGet, ts.url("/999"), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramExerciseHandler_Create(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		body := fmt.Sprintf(`{"exercise_id":%d,"laterality":"bilateral","target_reps":10,"target_weight_kg":20.5}`, ts.exerciseID)
		r := httptest.NewRequest(http.MethodPost, ts.url(""), strings.NewReader(body))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ExerciseID != ts.exerciseID {
			t.Errorf("got exercise_id %d, want %d", got.ExerciseID, ts.exerciseID)
		}
		if got.Laterality == nil || *got.Laterality != "bilateral" {
			t.Errorf("got laterality %v, want %q", got.Laterality, "bilateral")
		}
	})

	t.Run("missing exercise_id returns 400", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		r := httptest.NewRequest(http.MethodPost, ts.url(""), strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		r := httptest.NewRequest(http.MethodPost, ts.url(""), strings.NewReader(`not json`))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramExerciseHandler_Update(t *testing.T) {
	t.Run("updates exercise", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)
		created, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		reps := 12
		body := fmt.Sprintf(`{"exercise_id":%d,"target_reps":%d}`, ts.exerciseID, reps)

		r := httptest.NewRequest(http.MethodPut, ts.url(fmt.Sprintf("/%d", created.ID)), strings.NewReader(body))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.TargetReps == nil || *got.TargetReps != 12 {
			t.Errorf("got target_reps %v, want 12", got.TargetReps)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		body := fmt.Sprintf(`{"exercise_id":%d}`, ts.exerciseID)
		r := httptest.NewRequest(http.MethodPut, ts.url("/999"), strings.NewReader(body))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("missing exercise_id returns 400", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)
		created, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, ts.url(fmt.Sprintf("/%d", created.ID)), strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramExerciseHandler_Delete(t *testing.T) {
	t.Run("deletes exercise", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)
		created, err := ts.svc.Create(ts.setID, ts.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodDelete, ts.url(fmt.Sprintf("/%d", created.ID)), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ts := newTestProgramExerciseServer(t)

		r := httptest.NewRequest(http.MethodDelete, ts.url("/999"), nil)
		w := httptest.NewRecorder()
		ts.mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
