package service

import (
	"context"
	"fmt"
	"log/slog"
)

// CleanupExpired releases up to `batchSize` expired active reservations.
// Returns the count successfully released. Per-row release failures are
// logged but don't abort the batch — transient errors on one reservation
// shouldn't block the rest.
func (s *Service) CleanupExpired(ctx context.Context, batchSize int32) (int, error) {
	expired, err := s.repo.ListExpired(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("list expired: %w", err)
	}

	released := 0
	for _, r := range expired {
		if err := s.repo.Release(ctx, r.ID); err != nil {
			slog.WarnContext(ctx, "release expired reservation failed",
				"reservation_id", r.ID,
				"product_id", r.ProductID,
				"err", err,
			)
			continue
		}
		released++
	}
	return released, nil
}
