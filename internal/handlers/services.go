package handlers

import (
	"context"
	"io/fs"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
)

type (
	ViteService interface {
		Tags() string
		AssetsFS() fs.FS
	}

	WheelService interface {
		CountWheels(context.Context) (int64, error)
		ListWheels(context.Context) ([]services.Wheel, error)
		GetWheelByID(context.Context, int64) (*services.Wheel, error)
		SetWheelStatus(context.Context, int64, services.WheelStatus) error
		DeleteWheelByID(context.Context, int64) error
		CreateWheel(context.Context, string, string) (*services.Wheel, error)
		DeclareWinnerForWheel(context.Context, int64, *services.Candidate) error
	}

	WheelEventsService interface {
		SubscribeToWheelEvents(ctx context.Context, wheelId int64) (<-chan interface{}, error)
		SubscribeToGlobalWheelEvents(ctx context.Context) (<-chan interface{}, error)
	}

	CandidateService interface {
		ListCandidatesByWheel(context.Context, int64) ([]services.Candidate, error)
		AddCandidateToWheel(context.Context, string, int64) error
		CandidateInCandidateList(string, []services.Candidate) bool
		DeleteCandidateById(context.Context, int64, int64) error
		GetRandomCandidateForWheel(context.Context, int64, int64) (*services.Candidate, error)
	}

	SessionService interface {
		GetSessionIdFromContext(context.Context) (string, bool)
	}

	AdminService interface {
		GetWheelMetadata(context.Context) ([]services.AdminWheelMetadata, error)
	}
)
