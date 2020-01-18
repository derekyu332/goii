package components

import (
	"errors"
	"fmt"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/derekyu332/goii/helper/security"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	DFT_LOCKER_JWT_SECRET = "5JLjQCKyxsFXbgcq6OAhJKxbtSZtBKmg"
)

type ILocker interface {
	IsLocked() bool
	OwnerId() string
	Lock(ownerId string, d time.Duration, behavior interface{}) error
	Can(ownerId string, behavior interface{}) bool
	Unlock(ownerId string) error
}

type TransactionLock struct {
	LockerJwtSecret string
}

func (this *TransactionLock) Initialize() error {
	if this.LockerJwtSecret == "" {
		this.LockerJwtSecret = DFT_LOCKER_JWT_SECRET
	}

	return nil
}

func (this *TransactionLock) Lock(locker ILocker, ownerId string, d time.Duration, behavior interface{}) (string, error) {
	if locker.IsLocked() && locker.OwnerId() != ownerId {
		return "", errors.New("Locker is locked")
	}

	if err := locker.Lock(ownerId, d, behavior); err != nil {
		return "", err
	}

	now := time.Now().Unix()
	locker_key := security.MD5(fmt.Sprintf("%x%v%v", now, ownerId, this.LockerJwtSecret))
	var token *jwt.Token

	if d > 0 {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat":      now,
			"exp":      now + int64(d.Seconds()),
			"owner":    ownerId,
			"auth_key": locker_key,
		})
	} else {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat":      now,
			"owner":    ownerId,
			"auth_key": locker_key,
		})
	}

	tokenString, _ := token.SignedString([]byte(this.LockerJwtSecret))

	return tokenString, nil
}

func (this *TransactionLock) ownerFromToken(lock_token string) string {
	if lock_token == "" {
		return ""
	}

	token, err := jwt.Parse(lock_token, func(token *jwt.Token) (interface{}, error) {
		signMethod, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("Unexpected signing method")
		} else if signMethod != jwt.SigningMethodHS256 {
			return nil, errors.New("Unexpected signing method")
		}

		return []byte(this.LockerJwtSecret), nil
	})

	if err != nil {
		logger.Warning("jwt.Parse error %v", err.Error())
		return ""
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		logger.Info("%v", token.Claims)
		claimsMap, _ := token.Claims.(jwt.MapClaims)
		owner, ok0 := claimsMap["owner"].(string)
		iat, ok1 := claimsMap["iat"].(float64)
		auth_key, ok2 := claimsMap["auth_key"].(string)

		if !ok0 || !ok1 || !ok2 {
			return ""
		}

		locker_key := security.MD5(fmt.Sprintf("%x%v%v", int64(iat), owner, this.LockerJwtSecret))

		if locker_key != auth_key {
			logger.Warning("locker_key %v unmatch with %v", locker_key, auth_key)
			return ""
		}

		return owner
	}

	return ""
}

func (this *TransactionLock) Can(locker ILocker, lock_token string, behavior interface{}) bool {
	if !locker.IsLocked() {
		return true
	}

	owner := this.ownerFromToken(lock_token)
	return locker.Can(owner, behavior)
}

func (this *TransactionLock) Unlock(locker ILocker, lock_token string) error {
	if !locker.IsLocked() {
		return nil
	}

	owner := this.ownerFromToken(lock_token)

	if owner == "" {
		return errors.New("Invalid sign")
	}

	return locker.Unlock(owner)
}
