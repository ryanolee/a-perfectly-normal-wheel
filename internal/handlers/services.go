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
		ListWheels(context.Context) ([]services.Wheel, error)
		GetWheelByID(context.Context, int64) (*services.Wheel, error)
	}

	WheelEventsService interface {
		SubscribeToWheelEvents(ctx context.Context, wheelId int64) (<-chan interface{}, error)
	}

	CandidateService interface {
		ListCandidatesByWheel(context.Context, int64) ([]services.Candidate, error)
		AddCandidateToWheel(context.Context, string, int64) error
		CandidateInCandidateList(string, []services.Candidate) bool
	}

	SessionService interface {
		GetSessionIdFromContext(context.Context) (string, bool)
	}
)
