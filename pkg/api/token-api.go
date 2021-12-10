package api

import (
	"fmt"
	"pdserver/pkg/api/model"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type TokenDB struct {
	Storage *redis.Client
}

type TokenAPIService interface {
	Create(id uint64) (*model.TokenMetaData, error)
	Post(tmd *model.TokenMetaData) (*model.Token, error)
	RePost(refreshToken string) (*model.Token, error)
	Validate(token string) error
	FetchUserID(accessToken string) (*model.TokenMetaData, error)
	Delete(tmd *model.TokenMetaData) error
}

// NewTokenDB - Create redis db
func NewTokenDB(redisDB *redis.Client) *TokenDB {
	return &TokenDB{Storage: redisDB}
}

// Create - Create token meta data
func (db *TokenDB) Create(id uint64) (*model.TokenMetaData, error) {
	atExpire := time.Now().Add(time.Hour * 1).Unix()
	rtExpire := time.Now().Add(time.Hour * 24).Unix()
	accessUUID := uuid.NewString()
	refreshUUID := accessUUID + "++" + strconv.Itoa(int(id))

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["exp"] = atExpire
	atClaims["access_uuid"] = accessUUID
	atClaims["user_id"] = id
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(model.ACCESS_SECRET))
	if err != nil {
		return nil, err
	}
	rtClaims := jwt.MapClaims{}
	rtClaims["authorized"] = true
	rtClaims["exp"] = rtExpire
	rtClaims["refresh_uuid"] = refreshUUID
	rtClaims["user_id"] = id
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(model.REFRESH_SECRET))
	if err != nil {
		return nil, err
	}
	return &model.TokenMetaData{AccessToken: accessToken, RefreshToken: refreshToken,
		AtExpire: atExpire, RtExpire: rtExpire, UserID: id}, nil
}

// Post - Save token information to db
func (db *TokenDB) Post(tmd *model.TokenMetaData) (*model.Token, error) {
	atExpire := time.Unix(tmd.AtExpire, 0)
	rtExpire := time.Unix(tmd.RtExpire, 0)
	now := time.Now()

	if res := db.Storage.Set(tmd.AccessUUID, strconv.Itoa(int(tmd.UserID)), atExpire.Sub(now)); res != nil {
		return nil, res.Err()
	}
	if res := db.Storage.Set(tmd.RefreshUUID, strconv.Itoa(int(tmd.UserID)), rtExpire.Sub(now)); res != nil {
		return nil, res.Err()
	}
	return &model.Token{AccessToken: tmd.AccessToken, RefreshToken: tmd.RefreshToken}, nil
}

// Repost - Update token information
func (db *TokenDB) RePost(refreshToken string) (*model.Token, error) {
	token, err := jwt.Parse(refreshToken, KeyFunc)
	if err != nil {
		return nil, err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("Something went wrong")
		}
		userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		res, err := DeleteAToken(db.Storage, refreshUUID)
		if err != nil || res == 0 {
			return nil, err
		}
		newTmd, err := db.Create(userID)
		if err != nil {
			return nil, err
		}
		newToken, err := db.Post(newTmd)
		if err != nil {
			return nil, err
		}
		return newToken, nil
	} else {
		return nil, fmt.Errorf("Refresh token has expired")
	}
}

// Validate - Verify the token is valid
func (db *TokenDB) Validate(token string) error {
	jwtToken, err := jwt.Parse(token, KeyFunc)
	if err != nil {
		return err
	}
	if _, ok := jwtToken.Claims.(jwt.Claims); !ok || !jwtToken.Valid {
		return fmt.Errorf("Invalid token")
	}
	return nil
}

// FetchUserID - Fetch User ID from access token
func (db *TokenDB) GetTokenMetaData(accessToken string) (*model.TokenMetaData, error) {
	jwtToken, err := jwt.Parse(accessToken, KeyFunc)
	if err != nil {
		return nil, err
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok && jwtToken.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("Something went wrong in access uuid")
		}
		refreshUUID, ok := claims["refresh_uuid"].(string)
		if !ok {
			return nil, fmt.Errorf("Something went wrong in refresh uuid")
		}
		userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &model.TokenMetaData{AccessUUID: accessUUID, RefreshUUID: refreshUUID, UserID: userID}, nil
	}
	return nil, fmt.Errorf("Something went wrong in jwt token claims")
}

// Delete - Delete tokens
func (db *TokenDB) Delete(tmd *model.TokenMetaData) error {
	atRes, err := DeleteAToken(db.Storage, tmd.AccessUUID)
	if err != nil {
		return err
	}
	rtRes, err := DeleteAToken(db.Storage, tmd.RefreshUUID)
	if err != nil {
		return err
	}
	if atRes != 1 || rtRes != 1 {
		return fmt.Errorf("Something went wrong")
	}
	return nil
}

func KeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(model.REFRESH_SECRET), nil
}

func DeleteAToken(storage *redis.Client, tokenUUID string) (uint64, error) {
	deleted, err := storage.Del(tokenUUID).Result()
	if err != nil {
		return 0, err
	}
	return uint64(deleted), nil
}
