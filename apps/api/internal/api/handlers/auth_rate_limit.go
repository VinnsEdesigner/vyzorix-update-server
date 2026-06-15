package controllers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// getResendDelay calculates the required delay in seconds based on resend count.
// Returns 0 for first resend, then (count-1) * 30 seconds.
func GetResendDelay(resendCount int) int {
	if resendCount <= 1 {
		return 0
	}
	delay := (resendCount - 1) * 30
	if delay < 0 {
		return 0
	}
	return delay
}

// CheckResendRateLimit checks if a resend is allowed based on the tracker state.
// Returns the delay to wait (0 = allowed), whether locked out, and lockout end time.
func CheckResendRateLimit(
	tracker *models.PasswordResetResendTracker,
	now time.Time,
) (allowed bool, retryAfter int, lockedUntil *time.Time) {
	// First resend - always allowed
	if tracker == nil {
		return true, 0, nil
	}

	// Check if currently locked out
	if tracker.LockoutUntil != nil && now.Before(*tracker.LockoutUntil) {
		return false, 0, tracker.LockoutUntil
	}

	// Calculate required delay based on resend count
	requiredDelay := GetResendDelay(tracker.ResendCount)
	if requiredDelay > 0 {
		timeSinceLastResend := now.Sub(tracker.LastResendAt).Seconds()
		if timeSinceLastResend < float64(requiredDelay) {
			retryAfter = requiredDelay - int(timeSinceLastResend)
			return false, retryAfter, nil
		}
	}

	return true, 0, nil
}

// HandleRateLimitResponse sends the appropriate rate limit response.
func HandleRateLimitResponse(c *gin.Context, retryAfter int, lockedUntil *time.Time) {
	if lockedUntil != nil {
		c.JSON(429, models.ResendPasswordResetResponse{
			Success:     false,
			Message:     "Too many requests. Please try again later.",
			LockedUntil: lockedUntil.UnixMilli(),
		})
	} else {
		c.JSON(429, models.ResendPasswordResetResponse{
			Success:    false,
			Message:    "Please wait before requesting another reset link.",
			RetryAfter: retryAfter,
		})
	}
}

// UpdateResendTracker updates the resend tracker and returns the new count.
func UpdateResendTracker(
	ctx context.Context,
	store interface {
		UpsertPasswordResetResendTracker(ctx context.Context, tracker *models.PasswordResetResendTracker) error
	},
	log interface {
		WarnContext(ctx context.Context, msg string, args ...interface{})
	},
	tracker *models.PasswordResetResendTracker,
	emailHash string,
	now time.Time,
) int {
	newResendCount := 1
	var lockoutUntil *time.Time

	if tracker != nil {
		newResendCount = tracker.ResendCount + 1

		// Check if we've hit the lockout threshold (6 attempts)
		if newResendCount > 6 {
			lockoutDuration := 5 * time.Hour
			lockout := now.Add(lockoutDuration)
			lockoutUntil = &lockout
			newResendCount = tracker.ResendCount // Don't increment on lockout
		}
	}

	newTracker := &models.PasswordResetResendTracker{
		ID:           GenerateID(),
		EmailHash:    emailHash,
		ResendCount:  newResendCount,
		LastResendAt: now,
		LockoutUntil: lockoutUntil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := store.UpsertPasswordResetResendTracker(ctx, newTracker); err != nil {
		log.WarnContext(ctx, "resendPasswordReset: failed to update tracker", "err", err)
	}

	return newResendCount
}

// TriggerTrackerCleanup starts async cleanup of old trackers.
func TriggerTrackerCleanup(
	store interface {
		CleanupPasswordResetResendTrackers(ctx context.Context, maxAgeHours int) (int64, error)
	},
	log interface {
		Warn(msg string, args ...interface{})
		Info(msg string, args ...interface{})
	},
) {
	go func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if deleted, err := store.CleanupPasswordResetResendTrackers(cleanupCtx, 24); err != nil {
			log.Warn("resendPasswordReset: cleanup failed", "err", err)
		} else if deleted > 0 {
			log.Info("resendPasswordReset: cleaned up trackers", "count", deleted)
		}
	}()
}