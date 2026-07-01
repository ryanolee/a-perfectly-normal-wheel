package repository

import (
	"context"
	"time"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type candidateStore interface {
	ListCandidatesByWheel(context.Context, int64) ([]db.Candidate, error)
	CreateCandidate(context.Context, db.CreateCandidateParams) error
	GetDuplicateCandidatesForWheel(context.Context, db.GetDuplicateCandidatesForWheelParams) (db.Candidate, error)
	GetCandidateByCreatorIDAndWheelID(context.Context, db.GetCandidateByCreatorIDAndWheelIDParams) (db.Candidate, error)
	GetCandidateByID(context.Context, db.GetCandidateByIDParams) (db.Candidate, error)
	DeleteCandidateByID(context.Context, db.DeleteCandidateByIDParams) error
}

type CandidateRepository struct {
	db candidateStore
}

func NewCandidateRepository(db *db.Queries) *CandidateRepository {
	return &CandidateRepository{db: db}
}

func (r *CandidateRepository) List(ctx context.Context, wheelID int64) ([]Candidate, error) {
	dbCandidates, err := r.db.ListCandidatesByWheel(ctx, wheelID)
	if err != nil {
		return nil, err
	}
	return candidateListFromDB(dbCandidates), nil
}

func (r *CandidateRepository) Get(ctx context.Context, id, wheelID int64) (*Candidate, error) {
	dbCandidate, err := r.db.GetCandidateByID(ctx, db.GetCandidateByIDParams{
		ID:      id,
		WheelID: wheelID,
	})
	if err != nil {
		return nil, err
	}
	candidate := candidateFromDB(dbCandidate)
	return &candidate, nil
}

// Create inserts the candidate then rehydrates it to return the generated ID.
func (r *CandidateRepository) Create(ctx context.Context, username string, wheelID int64, creatorID string) (*Candidate, error) {
	if err := r.db.CreateCandidate(ctx, db.CreateCandidateParams{
		Username:  username,
		WheelID:   wheelID,
		CreatorID: creatorID,
	}); err != nil {
		return nil, err
	}

	dbCandidate, err := r.db.GetCandidateByCreatorIDAndWheelID(ctx, db.GetCandidateByCreatorIDAndWheelIDParams{
		CreatorID: creatorID,
		WheelID:   wheelID,
	})
	if err != nil {
		return nil, err
	}

	candidate := candidateFromDB(dbCandidate)
	return &candidate, nil
}

func (r *CandidateRepository) Delete(ctx context.Context, id, wheelID int64) error {
	return r.db.DeleteCandidateByID(ctx, db.DeleteCandidateByIDParams{
		ID:      id,
		WheelID: wheelID,
	})
}

// FindDuplicate returns an existing candidate on the wheel matching the username
// or creator, or sql.ErrNoRows when none exists.
func (r *CandidateRepository) FindDuplicate(ctx context.Context, wheelID int64, username, creatorID string) (*Candidate, error) {
	dbCandidate, err := r.db.GetDuplicateCandidatesForWheel(ctx, db.GetDuplicateCandidatesForWheelParams{
		WheelID:   wheelID,
		Username:  username,
		CreatorID: creatorID,
	})
	if err != nil {
		return nil, err
	}
	candidate := candidateFromDB(dbCandidate)
	return &candidate, nil
}

type Candidate struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatorID string    `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
	WheelID   int64     `json:"wheel_id"`
}

func candidateListFromDB(dbCandidates []db.Candidate) []Candidate {
	candidates := make([]Candidate, len(dbCandidates))
	for i, dbCandidate := range dbCandidates {
		candidates[i] = candidateFromDB(dbCandidate)
	}
	return candidates
}

func candidateFromDB(u db.Candidate) Candidate {
	return Candidate{
		ID:        u.ID,
		Username:  u.Username,
		CreatorID: u.CreatorID,
		CreatedAt: u.CreatedAt.Time,
		WheelID:   u.WheelID,
	}
}
