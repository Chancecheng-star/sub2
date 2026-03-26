package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type invalidCleanupRepoStub struct {
	accounts      []Account
	deletedIDs    []int64
	listBefore    time.Time
	listKeywords  []string
	listCalled    int
	deleteCalled  int
	deleteErrByID map[int64]error
}

func (r *invalidCleanupRepoStub) ListInvalidAccountsForCleanup(_ context.Context, before time.Time, matchKeywords []string) ([]Account, error) {
	r.listCalled++
	r.listBefore = before
	r.listKeywords = append([]string(nil), matchKeywords...)
	return append([]Account(nil), r.accounts...), nil
}

func (r *invalidCleanupRepoStub) Delete(_ context.Context, id int64) error {
	r.deleteCalled++
	if err := r.deleteErrByID[id]; err != nil {
		return err
	}
	r.deletedIDs = append(r.deletedIDs, id)
	return nil
}

func TestInvalidAccountCleanupService_RunOnceDeletesMatchedAccounts(t *testing.T) {
	repo := &invalidCleanupRepoStub{accounts: []Account{{ID: 1, Platform: PlatformGemini, ErrorMessage: "Authentication failed (401)", UpdatedAt: time.Now().Add(-2 * time.Hour)}}}
	svc := &InvalidAccountCleanupService{
		repo: repo,
		cfg: config.TokenRefreshInvalidCleanupConfig{
			Enabled:            true,
			GracePeriodMinutes: 60,
			MatchKeywords:      []string{"401", "invalid_grant"},
		},
	}

	beforeStart := time.Now().Add(-61 * time.Minute)
	svc.runOnce()
	afterEnd := time.Now().Add(-59 * time.Minute)

	require.Equal(t, 1, repo.listCalled)
	require.Len(t, repo.deletedIDs, 1)
	require.Equal(t, int64(1), repo.deletedIDs[0])
	require.True(t, !repo.listBefore.Before(beforeStart) && !repo.listBefore.After(afterEnd), "before should be around now-60m")
	require.Equal(t, []string{"401", "invalid_grant"}, repo.listKeywords)
}

func TestInvalidAccountCleanupService_RunOnceDryRunDoesNotDelete(t *testing.T) {
	repo := &invalidCleanupRepoStub{accounts: []Account{{ID: 7, Platform: PlatformOpenAI, ErrorMessage: "invalid_grant", UpdatedAt: time.Now().Add(-24 * time.Hour)}}}
	svc := &InvalidAccountCleanupService{
		repo: repo,
		cfg: config.TokenRefreshInvalidCleanupConfig{
			Enabled:            true,
			GracePeriodMinutes: 30,
			DryRun:             true,
			MatchKeywords:      []string{"invalid_grant"},
		},
	}

	svc.runOnce()

	require.Equal(t, 1, repo.listCalled)
	require.Empty(t, repo.deletedIDs)
	require.Equal(t, 0, repo.deleteCalled)
}

func TestNormalizedCleanupKeywords_DefaultsAndDeduplicates(t *testing.T) {
	got := normalizedCleanupKeywords([]string{" 401 ", "INVALID_GRANT", "401", ""})
	require.Equal(t, []string{"401", "invalid_grant"}, got)

	defaults := normalizedCleanupKeywords(nil)
	require.Contains(t, defaults, "401")
	require.Contains(t, defaults, "invalid_grant")
	require.Contains(t, defaults, "unauthorized_client")
}
