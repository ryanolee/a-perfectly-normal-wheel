package services

import "context"

type (
	AdminService struct {
		wheelService     *WheelService
		candidateService *CandidateService
	}

	AdminWheelMetadata struct {
		Wheel      Wheel
		Candidates []Candidate
	}
)

func NewAdminService(wheelService *WheelService, candidateService *CandidateService) *AdminService {
	return &AdminService{
		wheelService:     wheelService,
		candidateService: candidateService,
	}
}

func (s *AdminService) GetWheelMetadata(ctx context.Context) ([]AdminWheelMetadata, error) {
	wheels, err := s.wheelService.ListWheels(ctx)
	if err != nil {
		return nil, err
	}

	metadataList := make([]AdminWheelMetadata, len(wheels))
	for i, wheel := range wheels {
		candidates, err := s.candidateService.ListCandidatesByWheel(ctx, wheel.ID)
		if err != nil {
			return nil, err
		}

		metadataList[i] = AdminWheelMetadata{
			Wheel:      wheel,
			Candidates: candidates,
		}
	}

	return metadataList, nil
}
