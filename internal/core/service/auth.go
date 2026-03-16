package service

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *LTIService) Register(ctx context.Context, openidUrl, registrationToken string) error {
	configuration, err := s.ltiClient.GetOpenidConfiguration(openidUrl)
	if err != nil {
		return err
	}
	registrationBody := s.prepareToolRegistrationRequest()

	toolCfg, err := s.ltiClient.SendRegistrationRequest(configuration.RegistrationEndpoint, registrationToken, registrationBody)
	if err != nil {
		return err
	}

	str, err := json.Marshal(*toolCfg)
	println(string(str))

	platform := &domain.Platform{
		ClientID:        toolCfg.ClientId,
		DeploymentID:    toolCfg.LtiConfiguration.DeploymentID,
		JwksUrl:         configuration.JwksUri,
		AuthUrl:         configuration.AuthorizationEndpoint,
		Issuer:          configuration.Issuer,
		Name:            configuration.Issuer,
		TokenUrl:        configuration.TokenEndpoint,
		RegistrationUrl: configuration.RegistrationEndpoint,
	}
	_, err = s.platformRepo.Create(ctx, platform)
	if err != nil {
		return err
	}
	return nil
}

func (s *LTIService) Login(ctx context.Context, req *domain.LoginRequest) (string, error) {
	platform, err := s.platformRepo.GetByIssuerAndClientID(ctx, req.Iss, req.ClientID)
	if err != nil {
		return "", err
	}

	nonce := generateRandomString(32)
	state := generateRandomString(32)

	loginSession := &domain.LoginSession{
		Nonce:     nonce,
		State:     state,
		Issuer:    req.Iss,
		CreatedAt: time.Now(),
	}

	err = s.loginSessionRepo.Save(ctx, loginSession)
	if err != nil {
		return "", err
	}

	query, err := s.buildAuthURL(platform, req, loginSession)

	if err != nil {
		return "", err
	}

	return query, nil
}

func (s *LTIService) prepareToolRegistrationRequest() *domain.ToolRegistrationRequest {
	registrationBody := domain.ToolRegistrationRequest{
		ClientID:                "709sdfnjkds12",
		RegistrationClientUri:   s.cfg.LTIConfig.RegistrationClientUri,
		InitiateLoginUri:        s.cfg.LTIConfig.LoginClientUri,
		RegistrationAccessToken: "iDPzMyKHMX_4CkTpwLDCK",
		ApplicationType:         s.cfg.LTIConfig.ApplicationType,
		JWKSUri:                 s.cfg.OAuthConfig.JWKSUri,
		TokenEndpointAuthMethod: "private_key_jwt",
		ResponseTypes:           []string{"id_token"},
		GrantTypes:              []string{"client_credentials", "implicit"},
		RedirectUris:            []string{s.cfg.OAuthConfig.RedirectURI},
		ClientName:              s.cfg.LTIConfig.ToolName,
		Scope: strings.Join([]string{
			"https://purl.imsglobal.org/spec/lti-ags/scope/lineitem",
			"https://purl.imsglobal.org/spec/lti-ags/scope/lineitem.readonly",
			"https://purl.imsglobal.org/spec/lti-ags/scope/result.readonly",
			"https://purl.imsglobal.org/spec/lti-ags/scope/score",
			"https://purl.imsglobal.org/spec/lti-nrps/scope/contextmembership.readonly",
		}, " "),
		ToolConfiguration: domain.ToolConfiguration{
			Domain: s.cfg.LTIConfig.Domain,
			Claims: []string{
				"iss",
				"sub",
				"aud",
				"exp",
				"iat",
				"nonce",
				"name",
				"given_name",
				"family_name",
				"email",
				"https://purl.imsglobal.org/spec/lti/claim/context",
				"https://purl.imsglobal.org/spec/lti/claim/deployment_id",
				"https://purl.imsglobal.org/spec/lti/claim/roles",
				"https://purl.imsglobal.org/spec/lti/claim/custom",
				"https://purl.imsglobal.org/spec/lti/claim/launch_presentation",
				"https://purl.imsglobal.org/spec/lti-ags/claim/endpoint",
				"https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice",
				"https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings",
			},
			Messages: []domain.LTIMessage{
				{
					Type: "LtiResourceLinkRequest",
					Placements: []string{
						"course_assignments_menu",
						"assignment_edit_menu",
						"link_embed_button",
						"module_quicklinks_menu",
					},
				},
				{
					Type: "LtiDeepLinkingRequest",
					Placements: []string{
						"course_assignments_menu",
						"link_embed_button",
					},
				},
			},
		},
	}
	return &registrationBody
}

func (s *LTIService) buildAuthURL(platform *domain.Platform, req *domain.LoginRequest, loginSession *domain.LoginSession) (string, error) {

	authURL, err := url.Parse(platform.AuthUrl)
	if err != nil {
		return "", err
	}

	q := authURL.Query()
	q.Set("response_type", "id_token")
	q.Set("response_mode", "form_post")
	q.Set("scope", "openid")
	q.Set("id_token_signed_response_alg", "RS256")
	q.Set("client_id", platform.ClientID)
	q.Set("redirect_uri", s.cfg.OAuthConfig.RedirectURI)
	q.Set("login_hint", strconv.Itoa(req.LoginHint))
	q.Set("lti_message_hint", strconv.Itoa(req.LTIMessageHint))
	q.Set("state", loginSession.State)
	q.Set("nonce", loginSession.Nonce)
	q.Set("prompt", "none")

	authURL.RawQuery = q.Encode()

	return authURL.String(), nil
}
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
