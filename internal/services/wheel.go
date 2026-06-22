package services

import (
	"context"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type (
	WheelQueries interface {
		ListWheels(context.Context) ([]db.Wheel, error)
		GetWheelByID(context.Context, int64) (db.Wheel, error)
	}

	WheelService struct {
		dbQueries WheelQueries
	}
)

func NewWheelService(dbQueries WheelQueries) *WheelService {
	return &WheelService{
		dbQueries: dbQueries,
	}
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

type Wheel struct {
	ID          int64
	Name        string
	Description string
	WinnerID    *int64
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
	}
}
