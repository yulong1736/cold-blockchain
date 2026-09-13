package services

import (
	"cold-chain-trace/backend/models"
	"errors"
	"testing"
	"time"
)

func TestIsResetCodeUsable(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	hash, err := hashResetCode("123456")
	if err != nil {
		t.Fatalf("hash reset code: %v", err)
	}

	validRecord := &models.PasswordResetCode{
		CodeHash:  hash,
		ExpiresAt: now.Add(time.Minute),
	}
	if err := isResetCodeUsable(validRecord, "123456", now); err != nil {
		t.Fatalf("valid record should pass: %v", err)
	}

	expiredRecord := &models.PasswordResetCode{
		CodeHash:  hash,
		ExpiresAt: now.Add(-time.Second),
	}
	if err := isResetCodeUsable(expiredRecord, "123456", now); !errors.Is(err, ErrResetCodeInvalid) {
		t.Fatalf("expired code should be invalid, got: %v", err)
	}

	usedAt := now.Add(-time.Minute)
	usedRecord := &models.PasswordResetCode{
		CodeHash:  hash,
		ExpiresAt: now.Add(time.Minute),
		UsedAt:    &usedAt,
	}
	if err := isResetCodeUsable(usedRecord, "123456", now); !errors.Is(err, ErrResetCodeInvalid) {
		t.Fatalf("used code should be invalid, got: %v", err)
	}

	if err := isResetCodeUsable(validRecord, "654321", now); !errors.Is(err, ErrResetCodeInvalid) {
		t.Fatalf("wrong code should be invalid, got: %v", err)
	}
}

func TestApplyResetPassword(t *testing.T) {
	user := &models.User{}
	if err := applyResetPassword(user, "newpass123"); err != nil {
		t.Fatalf("reset password should hash successfully: %v", err)
	}
	if user.Password == "" {
		t.Fatal("expected hashed password to be stored")
	}
	if !user.CheckPassword("newpass123") {
		t.Fatal("expected reset password to pass login password verification")
	}
	if user.CheckPassword("oldpass123") {
		t.Fatal("old password should not match after reset")
	}

	shortPasswordUser := &models.User{}
	if err := applyResetPassword(shortPasswordUser, "123"); !errors.Is(err, ErrResetPasswordTooShort) {
		t.Fatalf("short password should be rejected, got: %v", err)
	}
}
