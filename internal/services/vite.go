package services

import (
	"embed"
	"io/fs"
	"log"

	"github.com/olivere/vite"
	"go.uber.org/zap"
)

type (
	ViteFS interface {
		embed.FS
	}
	ViteService struct {
		tags string
		fs   fs.FS
	}
)

func NewViteService(distFs *embed.FS, logger *zap.Logger) *ViteService {
	dist, err := fs.Sub(distFs, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	fragment, err := vite.HTMLFragment(vite.Config{
		FS:        dist,
		ViteEntry: "main.ts",
	})
	if err != nil {
		logger.Fatal("failed to create Vite HTML fragment", zap.Error(err))
	}
	return &ViteService{
		tags: string(fragment.Tags),
		fs:   dist,
	}
}

func (s *ViteService) Tags() string {
	return s.tags
}

func (s *ViteService) AssetsFS() fs.FS {
	return s.fs
}
