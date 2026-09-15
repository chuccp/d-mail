package auth

import (
	"net/http"

	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
)

type Authentication struct {
	ctx *core.Context
}

func (authentication *Authentication) Init(ctx *core.Context) error {
	authentication.ctx = ctx
	return nil
}

func (authentication *Authentication) SignIn(user any, request *web.Request) (any, error) {
	encrypt, err := Encrypt(user)
	if err != nil {
		return nil, err
	}
	// The session travels as an HttpOnly cookie, so the browser sends it automatically and
	// page script cannot read it.
	//
	// The name must be a valid RFC 6265 token. The ciphertext is base64 and contains '='
	// padding, so using it as the cookie name makes http.SetCookie drop the cookie
	// silently — which is why no Set-Cookie was ever emitted and User() could never find
	// the cookie it looks up by entity.UserToken.
	request.Cookie().Set(entity.UserToken, encrypt,
		web.WithHttpOnly(true),
		web.WithSameSite(http.SameSiteLaxMode),
	)
	return encrypt, nil
}
func (authentication *Authentication) SignOut(request *web.Request) (any, error) {
	request.Cookie().Delete(entity.UserToken)
	return web.Ok("success"), nil
}

func (authentication *Authentication) User(request *web.Request) (*model.User, error) {
	return User(request, authentication.ctx)
}

func (authentication *Authentication) NewUser() any {
	return &entity.LoginUser{}
}

func User(request *web.Request, ctx *core.Context) (*model.User, error) {
	token := request.Cookie().Get(entity.UserToken)
	if token == "" {
		token = request.GetHeader("Authorization")
	}
	if token == "" {
		return nil, auth2.NoLogin
	}
	loginUser := &entity.LoginUser{}
	err := Decrypt(token, loginUser)
	if err != nil {
		return nil, err
	}
	userModel := core.GetModel[*model.UserModel](ctx)
	var dbUser *model.User
	var dbErr error
	if loginUser.Id != 0 {
		dbUser, dbErr = userModel.FindByPK(loginUser.Id)
	} else if loginUser.Username != "" {
		dbUser, dbErr = userModel.FindOneByName(loginUser.Username)
	}
	if dbUser != nil && dbErr == nil {
		if dbUser.Salt == loginUser.Salt {
			return dbUser, nil
		}
	}
	return nil, auth2.NoLogin
}
