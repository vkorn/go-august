//go:generate enumer -type=AuthenticationMethods -transform=snake -trimprefix=AuthMeth -json -text -yaml

package august

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthenticationStates describes possible authentication states.
type AuthenticationStates int

const (
	// AuthRequiresAuthentication describes default state.
	AuthRequiresAuthentication AuthenticationStates = iota
	// AuthRequiresValidation describes verification required state.
	AuthRequiresValidation
	// AuthAuthenticated describes success authentication.
	AuthAuthenticated
	// AuthBadPassword describes wrong password.
	AuthBadPassword
)

// AuthenticationMethods defines possible authentication methods.
type AuthenticationMethods int

const (
	// AuthMethPhone describes cell auth method.
	AuthMethPhone AuthenticationMethods = iota
	// AuthMethEmail describes email auth method.
	AuthMethEmail
)

// AuthenticationState defines August API authentication state.
type AuthenticationState struct {
	InstallID   string
	AccessToken string
	State       AuthenticationStates
}

// Authenticator defines August API authenticator.
type Authenticator struct {
	UserName   string
	Password   string
	AuthMethod AuthenticationMethods

	State *AuthenticationState
}

// Authentication request.
type authRequest struct {
	Identifier string `json:"identifier"`
	InstallID  string `json:"install_id"`
	Password   string `json:"password"`
}

// Authentication response.
type authResponse struct {
	TokenExpiresAt string `json:"expiresAt"`
	Password       bool   `json:"vPassword"`
	InstallID      bool   `json:"vInstallId"`
}

// Verification code request.
type sendRequest struct {
	UserName string `json:"value"`
}

// NewAuthenticator creates a new Authenticator.
func NewAuthenticator(authMethod AuthenticationMethods, user string, password string,
	accessToken string, installID string) *Authenticator {
	auth := &Authenticator{
		UserName:   user,
		Password:   password,
		AuthMethod: authMethod,
		State: &AuthenticationState{
			InstallID:   installID,
			AccessToken: accessToken,
			State:       AuthRequiresAuthentication,
		},
	}

	if "" != accessToken {
		auth.State.State = AuthAuthenticated
	}

	return auth
}

// Authenticate makes an attempt to authenticate user.
func (a *Authenticator) Authenticate() error {
	if AuthAuthenticated == a.State.State {
		return nil
	}

	aReq := &authRequest{
		Identifier: fmt.Sprintf("%s:%s", a.AuthMethod.String(), a.UserName),
		InstallID:  a.State.InstallID,
		Password:   a.Password,
	}

	b, err := json.Marshal(aReq)
	if err != nil {
		return &ErrorGeneric{Message: err.Error()}
	}

	resp, err := doAPICall(a, http.MethodPost, "/session", bytes.NewReader(b))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	authResp := &authResponse{}
	err = dec.Decode(authResp)
	if err != nil {
		return &ErrorGeneric{Message: err.Error()}
	}

	token := resp.Header.Get(tokenHeader)
	if "" == token {
		return &ErrorGeneric{Message: "empty token"}
	}

	a.State.AccessToken = token
	if !authResp.Password {
		a.State.State = AuthBadPassword
	} else if !authResp.InstallID {
		a.State.State = AuthRequiresValidation
	} else {
		a.State.State = AuthAuthenticated
	}

	return nil
}

// SendVerificationCode requests new verification code to either email or phone, depends
// on which method has being used for Authenticate.
func (a *Authenticator) SendVerificationCode() error {
	req := &sendRequest{UserName: a.UserName}
	b, err := json.Marshal(req)
	if err != nil {
		return &ErrorGeneric{Message: err.Error()}
	}

	resp, err := doAPICall(a, http.MethodPost, "/validation/"+a.AuthMethod.String(), bytes.NewReader(b))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// ValidateCode performs the auth code validation.
func (a *Authenticator) ValidateCode(code string) error {
	req := map[string]string{
		a.AuthMethod.String(): a.UserName,
		"code":                code,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return &ErrorGeneric{Message: err.Error()}
	}

	resp, err := doAPICall(a, http.MethodPost, "/validate/"+a.AuthMethod.String(), bytes.NewReader(b))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
