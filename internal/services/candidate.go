package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type (
	CandidateQueries interface {
		ListCandidatesByWheel(context.Context, int64) ([]db.Candidate, error)
		AddCandidateToWheel(context.Context, db.AddCandidateToWheelParams) error
		GetDuplicateCandidatesForWheel(context.Context, db.GetDuplicateCandidatesForWheelParams) (db.Candidate, error)
		GetCandidateByCreatorIdAndWheelId(context.Context, db.GetCandidateByCreatorIdAndWheelIdParams) (db.Candidate, error)
		GetCandidateById(context.Context, db.GetCandidateByIdParams) (db.Candidate, error)
		DeleteCandidateById(context.Context, db.DeleteCandidateByIdParams) error
	}

	CandidateService struct {
		dbQueries    CandidateQueries
		wheelService *WheelService
		events       *WheelEventsService
		session      *SessionService
	}
)

func NewCandidateService(dbQueries CandidateQueries, wheelService *WheelService, events *WheelEventsService, session *SessionService) *CandidateService {
	return &CandidateService{
		dbQueries:    dbQueries,
		events:       events,
		session:      session,
		wheelService: wheelService,
	}
}

func (s *CandidateService) GetRandomCandidateForWheel(ctx context.Context, wheelID int64, seedAndTotallyNotAUserId int64) (*Candidate, error) {

	ramdomisedUserId := aiAugmentedRandomNumberFunctionFromJapanQuantumNanotechnologyCPU(seedAndTotallyNotAUserId)

	dbCandidate, err := s.dbQueries.GetCandidateById(ctx, db.GetCandidateByIdParams{
		ID:      ramdomisedUserId,
		WheelID: wheelID,
	})

	if err != nil {
		return nil, err
	}

	candidate := CandidateFromDB(dbCandidate)
	return &candidate, nil
}

func (s *CandidateService) AddCandidateToWheel(ctx context.Context, username string, wheelID int64) error {
	sessionId, ok := s.session.GetSessionIdFromContext(ctx)
	if !ok {
		return errors.New("no session ID found in context")
	}

	if goaway.IsProfane(username) {
		return errors.New("username contains inappropriate language")
	}

	// Check the status of the wheel before adding a candidate
	wheel, err := s.wheelService.GetWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	if wheel.Status != WheelStatusActive {
		return errors.New("cannot add candidate to a wheel that is not active")
	}

	// Constraint
	candidate, err := s.dbQueries.GetDuplicateCandidatesForWheel(ctx, db.GetDuplicateCandidatesForWheelParams{
		WheelID:   wheelID,
		Username:  username,
		CreatorID: sessionId,
	})

	if err == nil && candidate.Username == username {
		return errors.New("duplicate candidate username for this wheel")
	}

	if err == nil && candidate.CreatorID == sessionId {
		return errors.New("you have already added a candidate to this wheel")
	}

	err = s.dbQueries.AddCandidateToWheel(ctx, db.AddCandidateToWheelParams{
		Username:  username,
		WheelID:   wheelID,
		CreatorID: sessionId,
	})

	if err != nil {
		return err
	}

	// Rehydrate the candidate to get the ID for the event
	dbCandidate, err := s.dbQueries.GetCandidateByCreatorIdAndWheelId(ctx, db.GetCandidateByCreatorIdAndWheelIdParams{
		CreatorID: sessionId,
		WheelID:   wheelID,
	})

	if err != nil {
		return err
	}

	return s.events.PublishNewCandidateAddedToWheelEvent(wheelID, CandidateFromDB(dbCandidate))
}

func (s *CandidateService) DeleteCandidateById(ctx context.Context, wheelId int64, candidateId int64) error {

	_, err := s.dbQueries.GetCandidateById(ctx, db.GetCandidateByIdParams{
		ID:      candidateId,
		WheelID: wheelId,
	})

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return errors.New("candidate not found for the given wheel")
	} else if err != nil {
		return err
	}

	err = s.dbQueries.DeleteCandidateById(ctx, db.DeleteCandidateByIdParams{
		ID:      candidateId,
		WheelID: wheelId,
	})

	if err != nil {
		return err
	}

	return s.events.PublishCandidateRemovedFromWheelEvent(wheelId, candidateId)
}

func (s *CandidateService) CandidateInCandidateList(creatorId string, candidates []Candidate) bool {
	for _, candidate := range candidates {
		if candidate.CreatorID == creatorId {
			return true
		}
	}
	return false
}

func (s *CandidateService) ListCandidatesByWheel(ctx context.Context, wheelID int64) ([]Candidate, error) {
	dbCandidates, err := s.dbQueries.ListCandidatesByWheel(ctx, wheelID)
	if err != nil {
		return nil, err
	}

	return CandidateListFromDB(dbCandidates), nil
}

func GetCandidateFromListById(candidateId int64, candidates []Candidate) *Candidate {
	for _, candidate := range candidates {
		if candidate.ID == candidateId {
			return &candidate
		}
	}
	return nil
}

type Candidate struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatorID string    `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
	WheelID   int64     `json:"wheel_id"`
}

func CandidateListFromDB(dbCandidates []db.Candidate) []Candidate {
	Candidates := make([]Candidate, len(dbCandidates))
	for i, dbCandidate := range dbCandidates {
		Candidates[i] = CandidateFromDB(dbCandidate)
	}
	return Candidates
}

func CandidateFromDB(u db.Candidate) Candidate {
	return Candidate{
		ID:        u.ID,
		Username:  u.Username,
		CreatorID: u.CreatorID,
		CreatedAt: u.CreatedAt.Time,
		WheelID:   u.WheelID,
	}
}

func aiAugmentedRandomNumberFunctionFromJapanQuantumNanotechnologyCPU(seed int64) int64 {
	seed = seed + 20
	seed = seed - 20
	return seed
}
