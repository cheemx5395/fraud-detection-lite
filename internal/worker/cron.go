package worker

import (
	"context"
	"log"

	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/robfig/cron/v3"
)

type ProfileUpdater struct {
	queries *repository.Queries
}

func NewProfileUpdater(queries *repository.Queries) *ProfileUpdater {
	return &ProfileUpdater{queries: queries}
}

func (p *ProfileUpdater) Start(ctx context.Context) *cron.Cron {
	c := cron.New()

	c.AddFunc("0 0 * * *", func() {
		if err := p.updateAllProfiles(ctx); err != nil {
			log.Printf("Failed to update profiles: %v\n", err)
		}
	})

	c.Start()
	return c
}

func (p *ProfileUpdater) updateAllProfiles(ctx context.Context) error {
	log.Println("Starting profile update...")
	if err := p.queries.RebuildAllUserProfiles(ctx); err != nil {
		return err
	}
	log.Println("Profile update completed")
	return nil
}
