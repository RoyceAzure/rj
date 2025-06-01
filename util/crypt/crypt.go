package crypt

import (
	"errors"
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePassword 檢查密碼是否符合基本要求
// ValidatePassword 檢查密碼是否符合基本要求
func ValidateStringPassword(password string) error {
	var errs []error

	if len(password) < 8 {
		errs = append(errs, fmt.Errorf("密碼長度至少需要8個字符"))
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errs = append(errs, fmt.Errorf("密碼必須包含大寫字母"))
	}
	if !hasLower {
		errs = append(errs, fmt.Errorf("密碼必須包含小寫字母"))
	}
	if !hasNumber {
		errs = append(errs, fmt.Errorf("密碼必須包含數字"))
	}
	if !hasSpecial {
		errs = append(errs, fmt.Errorf("密碼必須包含特殊字符"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
