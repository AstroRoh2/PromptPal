package schema

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
)

type authAuthData struct {
	Address   string
	Signature string
	Message   string
}

type authInput struct {
	Auth authAuthData
}

type authResponse struct {
	token string
	u     *ent.User
}

func (q QueryResolver) Auth(ctx context.Context, args authInput) (result authResponse, err error) {
	payload := args.Auth
	verified, err := web3Service.VerifySignature(payload.Address, payload.Message, payload.Signature)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusBadRequest, err)
		return
	}
	if !verified {
		err = NewGraphQLHttpError(http.StatusBadRequest, errors.New("invalid signature"))
		return
	}

	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.Addr(strings.ToLower(payload.Address))).
		Only(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusNotFound, err)
		return
	}

	token, err := service.SignJWT(u, time.Hour*24*30)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	result.token = token
	result.u = u

	return
}
func (a authResponse) Token() string {
	return a.token
}
func (a authResponse) User() userResponse {
	return userResponse{u: a.u}
}
