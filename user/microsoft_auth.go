package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrMicrosoftAuthFailed = errors.New("microsoft authentication failed")
)

type MicrosoftAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TenantID     string // Your organization's tenant ID
}

type MicrosoftUserInfo struct {
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

func SetupMicrosoftAuth(msConfig MicrosoftAuthConfig) (*oauth2.Config, *oidc.Provider) {
	ctx := context.Background()

	// Microsoft endpoints for your tenant
	providerURL := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", msConfig.TenantID)
	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create OIDC provider: %v", err))
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     msConfig.ClientID,
		ClientSecret: msConfig.ClientSecret,
		RedirectURL:  msConfig.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return oauth2Config, provider
}

// ProcessMicrosoftCallback handles the OAuth2 callback and retrieves user information
func ProcessMicrosoftCallback(
	ctx context.Context,
	oauth2Config *oauth2.Config,
	provider *oidc.Provider,
	code string,
) (*MicrosoftUserInfo, error) {
	// Exchange code for token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	// Extract ID token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in token response")
	}

	// Verify ID token
	verifier := provider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
	parsedToken, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract user claims
	var claims MicrosoftUserInfo
	if err := parsedToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}
