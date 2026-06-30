package repository

import (
	"context"
	"database/sql"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type wheelStore interface {
	ListWheels(context.Context) ([]db.Wheel, error)
	GetWheelByID(context.Context, int64) (db.Wheel, error)
	SetWheelStatus(context.Context, db.SetWheelStatusParams) error
	DeleteWheelByID(context.Context, int64) error
	CreateWheel(context.Context, db.CreateWheelParams) (int64, error)
	CountWheels(context.Context) (int64, error)
	DeclareWinnerForWheel(context.Context, db.DeclareWinnerForWheelParams) error
}

type WheelRepository struct {
	db wheelStore
}

func NewWheelRepository(db wheelStore) *WheelRepository {
	return &WheelRepository{db: db}
}

func (r *WheelRepository) List(ctx context.Context) ([]Wheel, error) {
	dbWheels, err := r.db.ListWheels(ctx)
	if err != nil {
		return nil, err
	}
	return wheelListFromDB(dbWheels), nil
}

func (r *WheelRepository) Get(ctx context.Context, id int64) (*Wheel, error) {
	dbWheel, err := r.db.GetWheelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	wheel := wheelFromDB(&dbWheel)
	return &wheel, nil
}

func (r *WheelRepository) Count(ctx context.Context) (int64, error) {
	return r.db.CountWheels(ctx)
}

func (r *WheelRepository) Create(ctx context.Context, name, description string) (*Wheel, error) {
	id, err := r.db.CreateWheel(ctx, db.CreateWheelParams{
		Name: name,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	})
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *WheelRepository) SetStatus(ctx context.Context, id int64, status WheelStatus) error {
	return r.db.SetWheelStatus(ctx, db.SetWheelStatusParams{
		Status: status.String(),
		ID:     id,
	})
}

func (r *WheelRepository) Delete(ctx context.Context, id int64) error {
	return r.db.DeleteWheelByID(ctx, id)
}

func (r *WheelRepository) DeclareWinner(ctx context.Context, id, winnerID int64) error {
	return r.db.DeclareWinnerForWheel(ctx, db.DeclareWinnerForWheelParams{
		WinnerID: sql.NullInt64{
			Int64: winnerID,
			Valid: true,
		},
		ID: id,
	})
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
