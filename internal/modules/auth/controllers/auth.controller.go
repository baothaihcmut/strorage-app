package controllers

import (
	"fmt"
	"net/http"

	"github.com/baothaihcmut/Storage-app/internal/common/constant"
	middleware "github.com/baothaihcmut/Storage-app/internal/common/middlewares"
	"github.com/baothaihcmut/Storage-app/internal/common/response"
	"github.com/baothaihcmut/Storage-app/internal/config"
	"github.com/baothaihcmut/Storage-app/internal/modules/auth/interactors"
	"github.com/baothaihcmut/Storage-app/internal/modules/auth/presenter"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type AuthController interface {
	Init(g *gin.RouterGroup)
}

type AuthControllerImpl struct {
	authInteractor interactors.AuthInteractor
	jwtConfig      *config.JwtConfig
	oauth2Config   *config.Oauth2Config
}

// @Sumary Exchange Google token
// @Description Exchange Google auth code
// @Tags auth
// @Accept json
// @Produce json
// @Param authCode body presenter.ExchangeTokenInput true "auth code from google oauth2 resposne"
// @Success 201 {object} response.AppResponse{data=nil} "Login success"
// @Failure 401 {object} response.AppResponse{data=nil} "Wrong auth code"
//
//	@Router   /auth/exchange [post]
func (a *AuthControllerImpl) handleExchangeToken(c *gin.Context) {
	payload, _ := c.Get(string(constant.PayloadContext))
	res, err := a.authInteractor.ExchangeToken(c.Request.Context(), payload.(*presenter.ExchangeTokenInput))
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}
	//set cookie
	c.SetCookie("access_token", res.AccessToken, a.jwtConfig.AccessToken.Age*60*60, "/", "localhost", false, true)
	c.SetCookie("refresh_token", res.RefreshToken, a.jwtConfig.RefreshToken.Age*60*60, "/", "localhost", false, true)
	c.JSON(http.StatusCreated, response.InitResponse[any](true, "Login sucess", nil))
}

func (a *AuthControllerImpl) Init(g *gin.RouterGroup) {
	external := g.Group("/auth")
	external.POST(
		"/exchange",
		middleware.ValidateMiddleware[presenter.ExchangeTokenInput](false, binding.JSON),
		a.handleExchangeToken,
	)
	external.GET("/callback", func(c *gin.Context) {
		authCode := c.Query("code")
		if authCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing auth code"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": authCode})
	})

	external.GET("/redirect", func(c *gin.Context) {
		provider := c.Query("provider")
		var authRedirectURL string
		switch provider {
		case "google":
			authRedirectURL = fmt.Sprintf(
				"https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
				a.oauth2Config.Google.ClientId, a.oauth2Config.Google.RedirectURI, "email profile",
			)
		case "facebook":
			authRedirectURL = fmt.Sprintf(
				"https://www.facebook.com/v18.0/dialog/oauth?client_id=%s&redirect_uri=%s&state=secure_random_string&scope=,email,public_profile",
				a.oauth2Config.Facebook.ClientId, a.oauth2Config.Facebook.RedirectURI)
		}

		// Redirect user to Google OAuth
		c.Redirect(http.StatusFound, authRedirectURL)
	})

}

func NewAuthController(interactor interactors.AuthInteractor, jwtConfig *config.JwtConfig, oauth2Config *config.Oauth2Config) AuthController {
	return &AuthControllerImpl{
		authInteractor: interactor,
		jwtConfig:      jwtConfig,
		oauth2Config:   oauth2Config,
	}
}
