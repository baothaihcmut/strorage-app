package interactors

import (
	"context"

	"github.com/baothaihcmut/Storage-app/internal/common/logger"
	"github.com/baothaihcmut/Storage-app/internal/modules/auth/presenter"
	"github.com/baothaihcmut/Storage-app/internal/modules/auth/services"
	"github.com/baothaihcmut/Storage-app/internal/modules/users/models"
	"github.com/baothaihcmut/Storage-app/internal/modules/users/repositories"
)

type AuthInteractor interface {
	ExchangeToken(context.Context, *presenter.ExchangeTokenInput) (*presenter.ExchangeTokenOutput, error)
}

type AuthInteractorImpl struct {
	oauth2ServiceFactory services.Oauth2ServiceFactory
	jwtService           services.JwtService
	userRepository       repositories.UserRepository
	logger               logger.Logger
}

func NewAuthInteractor(oauth2 services.Oauth2ServiceFactory, userRepo repositories.UserRepository, jwtService services.JwtService, logger logger.Logger) AuthInteractor {
	return &AuthInteractorImpl{
		userRepository:       userRepo,
		oauth2ServiceFactory: oauth2,
		jwtService:           jwtService,
		logger:               logger,
	}
}
func (a *AuthInteractorImpl) ExchangeToken(ctx context.Context, input *presenter.ExchangeTokenInput) (*presenter.ExchangeTokenOutput, error) {
	//get service
	oauth2Service := a.oauth2ServiceFactory.GetOauth2Service(services.Oauth2ServiceToken(input.Provider))
	//get user info
	userInfo, err := oauth2Service.ExchangeToken(ctx, input.AuthCode)
	if err != nil {
		return nil, err
	}
	//check if user exist in system
	user, err := a.userRepository.FindUserByEmail(ctx, userInfo.GetEmail())
	if err != nil {
		return nil, err
	}
	//if user not exist create new user
	if user == nil {
		user = models.NewUser(userInfo.GetFirstName(), userInfo.GetLastName(), userInfo.GetEmail(), userInfo.GetImage())
		err = a.userRepository.CreateUser(ctx, user)
		if err != nil {
			a.logger.Errorf(ctx, map[string]interface{}{
				"email": userInfo.GetEmail(),
			}, "Error create new user:", err)
		}
		a.logger.Info(ctx, map[string]interface{}{
			"email":   userInfo.GetEmail(),
			"user_id": user.ID.Hex(),
		}, "User created")
	}
	//generate system token
	accessToken, err := a.jwtService.GenerateAccessToken(ctx, user.ID.Hex())
	if err != nil {
		a.logger.Errorf(ctx, map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}, "Error generate token: ", err)
		return nil, err
	}
	refreshToken, err := a.jwtService.GenerateRefreshToken(ctx, user.ID.Hex())
	if err != nil {
		a.logger.Errorf(ctx, map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}, "Error generate token: ", err)
		return nil, err
	}
	return &presenter.ExchangeTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}
