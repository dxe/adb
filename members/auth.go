package members

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"golang.org/x/oauth2"
)

// Cookie names.
const (
	membersIDToken = "members_id_token"
	membersState   = "members_state"
)

var conf, verifier = func() (*oauth2.Config, *oidc.IDTokenVerifier) {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}
	conf := &oauth2.Config{
		ClientID:     config.MembersClientID,
		ClientSecret: config.MembersClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  absURL("/auth"),
		Scopes:       []string{"email"},
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.MembersClientID,
	})
	return conf, verifier
}()

func (s *server) googleEmail() (string, error) {
	c, err := s.r.Cookie(membersIDToken)
	if err != nil {
		return "", err
	}

	token, err := verifier.Verify(s.r.Context(), c.Value)
	if err != nil {
		return "", err
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	err = token.Claims(&claims)
	if err != nil {
		return "", err
	}

	if !claims.EmailVerified {
		return "", errors.New("email not verified")
	}

	return claims.Email, nil
}

func (s *server) login() {
	state, err := nonce()
	if err != nil {
		s.error(err)
		return
	}
	http.SetCookie(s.w, &http.Cookie{
		Name:   membersState,
		Value:  state,
		MaxAge: 3600,
	})

	var opts []oauth2.AuthCodeOption
	if s.r.URL.Query()["force"] != nil {
		// If the user is currently only signed into one
		// Google Account, we need to set
		// prompt=select_account to force the account chooser
		// dialog to appear. Otherwise, Google will just
		// redirect back to us again immediately.
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "select_account"))
	}

	s.redirect(conf.AuthCodeURL(state, opts...))
}

func (s *server) auth() {
	c, err := s.r.Cookie(membersState)
	if err != nil {
		s.error(err)
		return
	}
	if c.Value != s.r.FormValue("state") {
		s.error(errors.New("state mismatch"))
		return
	}

	token, err := conf.Exchange(s.r.Context(), s.r.FormValue("code"))
	if err != nil {
		s.error(err)
		return
	}

	idToken := token.Extra("id_token").(string)
	http.SetCookie(s.w, &http.Cookie{
		Name:   membersIDToken,
		Value:  idToken,
		MaxAge: 3600,
	})
	s.redirect(absURL("/"))
}

// nonce returns a 256-bit random hex string.
func nonce() (string, error) {
	var buf [32]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
