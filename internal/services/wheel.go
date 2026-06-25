package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type (
	WheelQueries interface {
		ListWheels(context.Context) ([]db.Wheel, error)
		GetWheelByID(context.Context, int64) (db.Wheel, error)
		SetWheelStatus(context.Context, db.SetWheelStatusParams) error
		DeleteWheelByID(context.Context, int64) error
		CreateWheel(context.Context, db.CreateWheelParams) (int64, error)
		CountWheels(context.Context) (int64, error)
		DeclareWinnerForWheel(context.Context, db.DeclareWinnerForWheelParams) error
	}

	WheelService struct {
		dbQueries WheelQueries
		events    *WheelEventsService
	}
)

func NewWheelService(dbQueries WheelQueries, events *WheelEventsService) *WheelService {
	return &WheelService{
		dbQueries: dbQueries,
		events:    events,
	}
}

func (s *WheelService) DeclareWinnerForWheel(ctx context.Context, wheelID int64, randomCandidate *Candidate) error {
	wheel, err := s.GetWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	if wheel.Status == WheelStatusWinnerDeclared {
		return errors.New("cannot declare a winner for a wheel that has already declared a winner")
	}

	if randomCandidate.WheelID != wheelID {
		return errors.New("candidate does not belong to the specified wheel")
	}

	if err = s.dbQueries.DeclareWinnerForWheel(ctx, db.DeclareWinnerForWheelParams{
		WinnerID: sql.NullInt64{
			Int64: randomCandidate.ID,
			Valid: true,
		},
		ID: wheelID,
	}); err != nil {
		return err
	}

	return s.events.PublishWinnerDeclaredEvent(wheelID, *randomCandidate)
}

func (s *WheelService) CreateWheel(ctx context.Context, name string, description string) (*Wheel, error) {
	wheelID, err := s.dbQueries.CreateWheel(ctx, db.CreateWheelParams{
		Name: name,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	})

	if err != nil {
		return nil, err
	}

	wheel, err := s.GetWheelByID(ctx, wheelID)
	if err != nil {
		return nil, err
	}

	if err = s.events.PublishWheelAddedEvent(*wheel); err != nil {
		return nil, err
	}

	return wheel, err

}

func (s *WheelService) ListWheels(ctx context.Context) ([]Wheel, error) {
	dbWheels, err := s.dbQueries.ListWheels(ctx)
	if err != nil {
		return nil, err
	}

	return wheelListFromDB(dbWheels), nil
}

func (s *WheelService) GetWheelByID(ctx context.Context, id int64) (*Wheel, error) {
	dbWheel, err := s.dbQueries.GetWheelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wheel := wheelFromDB(&dbWheel)
	return &wheel, nil
}

func (s *WheelService) CountWheels(ctx context.Context) (int64, error) {
	count, err := s.dbQueries.CountWheels(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *WheelService) SetWheelStatus(ctx context.Context, wheelID int64, status WheelStatus) error {
	wheel, err := s.GetWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	if wheel.Status == WheelStatusWinnerDeclared {
		return errors.New("cannot change status of a wheel that has already declared a winner")
	}

	if err = s.dbQueries.SetWheelStatus(ctx, db.SetWheelStatusParams{
		Status: status.String(),
		ID:     wheelID,
	}); err != nil {
		return err
	}

	return s.events.PublishWheelStatusChangedEvent(wheelID, status)
}

func (s *WheelService) DeleteWheelByID(ctx context.Context, wheelID int64) error {
	err := s.dbQueries.DeleteWheelByID(ctx, wheelID)
	if err != nil {
		return err
	}

	return s.events.PublishWheelDeletedEvent(wheelID)
}

type WheelStatus int

const (
	WheelStatusActive WheelStatus = iota
	WheelStatusLocked
	WheelStatusWinnerDeclared
	WheelStatusUnknown
)

func (s WheelStatus) String() string {
	switch s {
	case WheelStatusActive:
		return "active"
	case WheelStatusLocked:
		return "locked"
	case WheelStatusWinnerDeclared:
		return "winner_declared"
	default:
		return "unknown"
	}
}

func ParseWheelStatus(status string) WheelStatus {
	switch status {
	case "active":
		return WheelStatusActive
	case "locked":
		return WheelStatusLocked
	case "winner_declared":
		return WheelStatusWinnerDeclared
	default:
		return WheelStatusUnknown
	}
}

type Wheel struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WinnerID    *int64
	Status      WheelStatus
}

func wheelListFromDB(dbWheels []db.Wheel) []Wheel {
	wheels := make([]Wheel, len(dbWheels))
	for i, dbWheel := range dbWheels {
		wheels[i] = wheelFromDB(&dbWheel)
	}
	return wheels
}

func wheelFromDB(w *db.Wheel) Wheel {
	var winnerID *int64
	if w.WinnerID.Valid {
		winnerID = &w.WinnerID.Int64
	}

	return Wheel{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description.String,
		WinnerID:    winnerID,
		Status:      ParseWheelStatus(w.Status),
	}
}
