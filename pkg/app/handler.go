package app

import (
	"errors"
	"fmt"
	"net/http"
	"pdserver/pkg/api"
	"pdserver/pkg/api/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

type Handler struct {
	Engin           *gin.Engine
	UserAPIService  *api.UserDB
	TokenAPIService *api.TokenDB
	NaverAPIService *api.NaverOAuth
	OTPAPIService   *api.OTPAPIService
}

func NewHandler(userService *api.UserDB, tokenService *api.TokenDB, naverService *api.NaverOAuth, otpService *api.OTPAPIService) *Handler {
	return &Handler{
		Engin:           gin.Default(),
		UserAPIService:  userService,
		TokenAPIService: tokenService,
		NaverAPIService: naverService,
		OTPAPIService:   otpService,
	}
}

func (h *Handler) LocalRegister(ctx *gin.Context) {
	var user model.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := h.UserAPIService.Post(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	token, err := h.Authenticate(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, model.LoginResponse{Token: *token, User: user})
}

func (h *Handler) LocalLogin(ctx *gin.Context) {
	var user model.User
	if err := ctx.ShouldBindJSON(user); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	_, err := h.UserAPIService.Get(&user)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err.Error())
		return
	}
	token, err := h.Authenticate(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, model.LoginResponse{Token: *token, User: user})
}

func (h *Handler) NaverLogin(ctx *gin.Context) {
	url := h.NaverAPIService.Config.AuthCodeURL(h.NaverAPIService.State, oauth2.AccessTypeOffline)
	http.Redirect(ctx.Writer, ctx.Request, url, http.StatusTemporaryRedirect)
}

func (h *Handler) NaverCallback(ctx *gin.Context) {
	naverToken, err := h.NaverAPIService.Auth(ctx.Writer, *ctx.Request)
	if err != nil {
		ctx.Redirect(http.StatusUnauthorized, "plantdoctor://")
		return
	}
	user, err := h.NaverAPIService.GetUser(naverToken)
	if err != nil {
		ctx.Redirect(http.StatusInternalServerError, "plantdoctor://")
		return
	}
	foundUser, err := h.UserAPIService.Get(user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err = h.UserAPIService.Post(user); err != nil {
			ctx.Redirect(http.StatusInternalServerError, "plantdoctor://")
			return
		} else {
			foundUser = user
		}
	}
	if err != nil {
		ctx.Redirect(http.StatusInternalServerError, "plantdoctor://")
		return
	}
	token, err := h.Authenticate(foundUser.ID)
	if err != nil {
		ctx.Redirect(http.StatusInternalServerError, "plantdoctor://")
		return
	}
	userID := strconv.FormatUint(foundUser.ID, 10)
	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("plantdoctor://?access_token=%s&&refresh_token=%s&&user_id=%s", token.AccessToken, token.RefreshToken, userID))
}

func (h *Handler) Authenticate(userID uint64) (*model.Token, error) {
	tmd, err := h.TokenAPIService.Create(userID)
	if err != nil {
		return nil, err
	}
	token, err := h.TokenAPIService.Post(tmd)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (h *Handler) SendEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(http.StatusUnprocessableEntity, "Wrong query")
		return
	}
	available, err := h.UserAPIService.Available(email, "email")
	if !available {
		if err == nil {
			ctx.JSON(http.StatusConflict, "Cannot use this email")
		} else {
			ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	err = h.OTPAPIService.SendEmail(email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, "Successfully sent email")
}

func (h *Handler) VerifyCode(ctx *gin.Context) {
	emailJson := map[string]string{}
	if err := ctx.ShouldBindJSON(&emailJson); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	code := emailJson["code"]
	isValidCode := h.OTPAPIService.Validate(code)
	if isValidCode {
		ctx.JSON(http.StatusOK, "Success")
	} else {
		ctx.JSON(http.StatusUnauthorized, "Failed")
	}
}

func (h *Handler) Available(ctx *gin.Context) {
	nickname := ctx.Query("nickname")
	if nickname == "" {
		ctx.JSON(http.StatusUnprocessableEntity, "Invalid query")
		return
	}
	res, err := h.UserAPIService.Available(nickname, "nickname")
	if !res {
		if err == nil {
			ctx.JSON(http.StatusConflict, "Unavailable nickname")
		} else {
			ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, "Available")
}

func (h *Handler) Logout(ctx *gin.Context) {
	accessToken, err := h.ExtractAccessToken(ctx.Request)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	tmd, err := h.TokenAPIService.GetTokenMetaData(accessToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err.Error())
		return
	}
	if err := h.TokenAPIService.Delete(tmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, "Successfully logged out")
}

func (h *Handler) RefreshToken(ctx *gin.Context) {
	mapToken := map[string]string{}
	if err := ctx.ShouldBindJSON(&mapToken); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]
	token, err := h.TokenAPIService.RePost(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, token)
}

func (h *Handler) GetUser(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, "Invalid userID query")
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.UserAPIService.GetWithID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *Handler) ValidateTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := h.ExtractAccessToken(ctx.Request)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, err.Error())
			ctx.Abort()
			return
		}
		if err := h.TokenAPIService.Validate(accessToken); err != nil {
			fmt.Println(err.Error())
			ctx.JSON(http.StatusUnauthorized, err.Error())
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func (h *Handler) ExtractAccessToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	strArr := strings.Split(authorization, " ")
	if len(strArr) != 2 {
		return "", fmt.Errorf("Invalid authorization")
	}
	return strArr[1], nil
}
