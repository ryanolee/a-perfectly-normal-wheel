package services

import (
	"context"
	"errors"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
)

type (
	WheelRepository interface {
		List(context.Context) ([]repository.Wheel, error)
		Get(context.Context, int64) (*repository.Wheel, error)
		Count(context.Context) (int64, error)
		Create(context.Context, string, string) (*repository.Wheel, error)
		SetStatus(context.Context, int64, repository.WheelStatus) error
		Delete(context.Context, int64) error
		DeclareWinner(context.Context, int64, int64) error
	}

	WheelService struct {
		wheels WheelRepository
		events *WheelEventsService
	}
)

func NewWheelService(wheels WheelRepository, events *WheelEventsService) *WheelService {
	return &WheelService{
		wheels: wheels,
		events: events,
	}
}

func (s *WheelService) DeclareWinnerForWheel(ctx context.Context, wheelID int64, randomCandidate *repository.Candidate) error {
	wheel, err := s.GetWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	if wheel.Status == repository.WheelStatusWinnerDeclared {
		return errors.New("cannot declare a winner for a wheel that has already declared a winner")
	}

	if randomCandidate.WheelID != wheelID {
		return errors.New("candidate does not belong to the specified wheel")
	}

	if err = s.wheels.DeclareWinner(ctx, wheelID, randomCandidate.ID); err != nil {
		return err
	}

	return s.events.PublishWinnerDeclaredEvent(wheelID, *randomCandidate)
}

func (s *WheelService) CreateWheel(ctx context.Context, name string, description string) (*repository.Wheel, error) {
	wheel, err := s.wheels.Create(ctx, name, description)
	if err != nil {
		return nil, err
	}

	if err = s.events.PublishWheelAddedEvent(*wheel); err != nil {
		return nil, err
	}

	return wheel, nil
}

func (s *WheelService) ListWheels(ctx context.Context) ([]repository.Wheel, error) {
	return s.wheels.List(ctx)
}

func (s *WheelService) GetWheelByID(ctx context.Context, id int64) (*repository.Wheel, error) {
	return s.wheels.Get(ctx, id)
}

func (s *WheelService) CountWheels(ctx context.Context) (int64, error) {
	return s.wheels.Count(ctx)
}

func (s *WheelService) SetWheelStatus(ctx context.Context, wheelID int64, status repository.WheelStatus) error {
	wheel, err := s.GetWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	if wheel.Status == repository.WheelStatusWinnerDeclared {
		return errors.New("cannot change status of a wheel that has already declared a winner")
	}

	if err = s.wheels.SetStatus(ctx, wheelID, status); err != nil {
		return err
	}

	return s.events.PublishWheelStatusChangedEvent(wheelID, status)
}

func (s *WheelService) DeleteWheelByID(ctx context.Context, wheelID int64) error {
	if err := s.wheels.Delete(ctx, wheelID); err != nil {
		return err
	}

	return s.events.PublishWheelDeletedEvent(wheelID)
}
