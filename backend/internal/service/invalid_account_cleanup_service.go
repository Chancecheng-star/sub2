package service

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type invalidAccountCleanupLister interface {
	ListInvalidAccountsForCleanup(ctx context.Context, before time.Time, matchKeywords []string) ([]Account, error)
}

type invalidAccountCleanupRepository interface {
	invalidAccountCleanupLister
	Delete(ctx context.Context, id int64) error
}

// InvalidAccountCleanupService periodically removes accounts that have been in
// error state for a grace period and whose error_message matches configured
// non-recoverable patterns (for example 401 / invalid_grant).
type InvalidAccountCleanupService struct {
	repo     invalidAccountCleanupRepository
	cfg      config.TokenRefreshInvalidCleanupConfig
	interval time.Duration
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewInvalidAccountCleanupService(accountRepo AccountRepository, cfg *config.Config) *InvalidAccountCleanupService {
	if accountRepo == nil || cfg == nil {
		return nil
	}
	repo, ok := accountRepo.(invalidAccountCleanupRepository)
	if !ok {
		return nil
	}
	cleanupCfg := cfg.TokenRefresh.InvalidCleanup
	interval := time.Duration(cleanupCfg.CheckIntervalMinutes) * time.Minute
	if cleanupCfg.CheckIntervalMinutes <= 0 {
		interval = 5 * time.Minute
	}
	return &InvalidAccountCleanupService{
		repo:     repo,
		cfg:      cleanupCfg,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *InvalidAccountCleanupService) Start() {
	if s == nil || s.repo == nil || !s.cfg.Enabled || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *InvalidAccountCleanupService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *InvalidAccountCleanupService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	graceMinutes := s.cfg.GracePeriodMinutes
	if graceMinutes <= 0 {
		graceMinutes = 1440
	}
	before := time.Now().Add(-time.Duration(graceMinutes) * time.Minute)
	keywords := normalizedCleanupKeywords(s.cfg.MatchKeywords)

	accounts, err := s.repo.ListInvalidAccountsForCleanup(ctx, before, keywords)
	if err != nil {
		slog.Warn("invalid_account_cleanup.list_failed", "error", err)
		return
	}
	if len(accounts) == 0 {
		return
	}

	deleted := 0
	for _, account := range accounts {
		if s.cfg.DryRun {
			slog.Info("invalid_account_cleanup.dry_run_hit",
				"account_id", account.ID,
				"platform", account.Platform,
				"updated_at", account.UpdatedAt,
				"error_message", account.ErrorMessage,
			)
			continue
		}
		if err := s.repo.Delete(ctx, account.ID); err != nil {
			slog.Warn("invalid_account_cleanup.delete_failed",
				"account_id", account.ID,
				"platform", account.Platform,
				"error", err,
			)
			continue
		}
		deleted++
		slog.Info("invalid_account_cleanup.deleted",
			"account_id", account.ID,
			"platform", account.Platform,
			"updated_at", account.UpdatedAt,
			"error_message", account.ErrorMessage,
		)
	}

	if s.cfg.DryRun {
		slog.Info("invalid_account_cleanup.dry_run_summary", "matched", len(accounts))
		return
	}
	if deleted > 0 {
		slog.Info("invalid_account_cleanup.summary", "deleted", deleted, "matched", len(accounts))
	}
}

func normalizedCleanupKeywords(keywords []string) []string {
	if len(keywords) == 0 {
		keywords = []string{"401", "invalid_grant", "invalid_client", "unauthorized_client", "access_denied"}
	}
	result := make([]string, 0, len(keywords))
	seen := make(map[string]struct{}, len(keywords))
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(strings.ToLower(keyword))
		if keyword == "" {
			continue
		}
		if _, ok := seen[keyword]; ok {
			continue
		}
		seen[keyword] = struct{}{}
		result = append(result, keyword)
	}
	return result
}
