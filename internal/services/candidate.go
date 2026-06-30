package services

import (
	"context"
	"database/sql"
	"errors"

	goaway "github.com/TwiN/go-away"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
)

type (
	CandidateRepository interface {
		List(context.Context, int64) ([]repository.Candidate, error)
		Get(context.Context, int64, int64) (*repository.Candidate, error)
		Create(context.Context, string, int64, string) (*repository.Candidate, error)
		Delete(context.Context, int64, int64) error
		FindDuplicate(context.Context, int64, string, string) (*repository.Candidate, error)
	}

	CandidateService struct {
		candidates   CandidateRepository
		wheelService *WheelService
		events       *WheelEventsService
		session      *SessionService
	}
)

func NewCandidateService(candidates CandidateRepository, wheelService *WheelService, events *WheelEventsService, session *SessionService) *CandidateService {
	return &CandidateService{
		candidates:   candidates,
		events:       events,
		session:      session,
		wheelService: wheelService,
	}
}

func (s *CandidateService) GetRandomCandidateForWheel(ctx context.Context, wheelID int64, seedAndTotallyNotAUserId int64) (*repository.Candidate, error) {
	ramdomisedUserId := aiAugmentedRandomNumberFunctionFromJapanQuantumNanotechnologyCPU(seedAndTotallyNotAUserId)
	return s.candidates.Get(ctx, ramdomisedUserId, wheelID)
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

	if wheel.Status != repository.WheelStatusActive {
		return errors.New("cannot add candidate to a wheel that is not active")
	}

	// Constraint
	duplicate, err := s.candidates.FindDuplicate(ctx, wheelID, username, sessionId)
	if err == nil {
		if duplicate.Username == username {
			return errors.New("duplicate candidate username for this wheel")
		}
		if duplicate.CreatorID == sessionId {
			return errors.New("you have already added a candidate to this wheel")
		}
	}

	candidate, err := s.candidates.Create(ctx, username, wheelID, sessionId)
	if err != nil {
		return err
	}

	return s.events.PublishNewCandidateAddedToWheelEvent(wheelID, *candidate)
}

func (s *CandidateService) DeleteCandidateById(ctx context.Context, wheelId int64, candidateId int64) error {
	_, err := s.candidates.Get(ctx, candidateId, wheelId)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return errors.New("candidate not found for the given wheel")
	} else if err != nil {
		return err
	}

	if err = s.candidates.Delete(ctx, candidateId, wheelId); err != nil {
		return err
	}

	return s.events.PublishCandidateRemovedFromWheelEvent(wheelId, candidateId)
}

func (s *CandidateService) CandidateInCandidateList(creatorId string, candidates []repository.Candidate) bool {
	for _, candidate := range candidates {
		if candidate.CreatorID == creatorId {
			return true
		}
	}
	return false
}

func (s *CandidateService) ListCandidatesByWheel(ctx context.Context, wheelID int64) ([]repository.Candidate, error) {
	return s.candidates.List(ctx, wheelID)
}

func GetCandidateFromListById(candidateId int64, candidates []repository.Candidate) *repository.Candidate {
	for _, candidate := range candidates {
		if candidate.ID == candidateId {
			return &candidate
		}
	}
	return nil
}

func aiAugmentedRandomNumberFunctionFromJapanQuantumNanotechnologyCPU(seed int64) int64 {
	seed = seed + 20
	seed = seed - 20
	return seed
}
