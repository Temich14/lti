package domain

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type OpenidConfiguration struct {
	Issuer                            string   `json:"issuer" form:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint" form:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint" form:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint" form:"userinfo_endpoint"`
	JwksUri                           string   `json:"jwks_uri" form:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint" form:"registration_endpoint"`
	ScopesSupported                   []string `json:"scopes_supported" form:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported" form:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported" form:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported" form:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported" form:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported" form:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported" form:"claims_supported"`
	ClaimTypesSupported               []string `json:"claim_types_supported,omitempty" form:"claim_types_supported,omitempty"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported,omitempty" form:"code_challenge_methods_supported,omitempty"`
}
type ToolRegistrationRequest struct {
	ClientID                string            `json:"client_id" form:"client_id"`
	RegistrationClientUri   string            `json:"registration_client_uri" form:"registration_client_uri"`
	RegistrationAccessToken string            `json:"registration_access_token" form:"registration_access_token"`
	ApplicationType         string            `json:"application_type" form:"application_type"`
	ResponseTypes           []string          `json:"response_types" form:"response_types"`
	GrantTypes              []string          `json:"grant_types" form:"grant_types"`
	InitiateLoginUri        string            `json:"initiate_login_uri" form:"initiate_login_uri"`
	RedirectUris            []string          `json:"redirect_uris" form:"redirect_uris"`
	ClientName              string            `json:"client_name" form:"client_name"`
	JWKSUri                 string            `json:"jwks_uri" form:"jwks_uri"`
	LogoutUri               string            `json:"logout_uri" form:"logout_uri"`
	TokenEndpointAuthMethod string            `json:"token_endpoint_auth_method" form:"token_endpoint_auth_method"`
	Contacts                []string          `json:"contacts" form:"contacts"`
	Scope                   string            `json:"scope" form:"scope"`
	ToolConfiguration       ToolConfiguration `json:"https://purl.imsglobal.org/spec/lti-tool-configuration" form:"https://purl.imsglobal.org/spec/lti-tool-configuration"`
}

type ToolRegistrationResponse struct {
	ClientId                string            `json:"client_id"`
	ResponseTypes           []string          `json:"response_types"`
	JwksUri                 string            `json:"jwks_uri"`
	InitiateLoginUri        string            `json:"initiate_login_uri"`
	GrantTypes              []string          `json:"grant_types"`
	RedirectUris            []string          `json:"redirect_uris"`
	ApplicationType         string            `json:"application_type"`
	TokenEndpointAuthMethod string            `json:"token_endpoint_auth_method"`
	ClientName              string            `json:"client_name"`
	LogoUri                 string            `json:"logo_uri"`
	Scope                   string            `json:"scope"`
	LtiConfiguration        ToolConfiguration `json:"https://purl.imsglobal.org/spec/lti-tool-configuration"`
}

type ToolConfiguration struct {
	jwt.RegisteredClaims
	Domain           string            `json:"domain" form:"domain"`
	TargetLinkUri    string            `json:"target_link_uri" form:"target_link_uri"`
	CustomParameters map[string]string `json:"custom_parameters" form:"custom_parameters"`
	Claims           []string          `json:"claims" form:"claims"`
	Description      string            `json:"description" form:"description"`
	Version          string            `json:"version" form:"version"`
	DeploymentID     string            `json:"deployment_id" form:"deployment_id"`
	Messages         []LTIMessage      `json:"messages" form:"messages"`
}

type LTIMessage struct {
	Type          string `json:"type" form:"type"`
	TargetLinkUri string `json:"target_link_uri" form:"target_link_uri"`
	Label         string `json:"label" form:"label"`
}

type RegistrationResponse struct {
}

type LoginRequest struct {
	Iss             string `json:"iss" form:"iss"`
	TargetLingUri   string `json:"target_ling_uri" form:"target_ling_uri"`
	LoginHint       int    `json:"login_hint" form:"login_hint"`
	LTIMessageHint  int    `json:"lti_message_hint" form:"lti_message_hint"`
	ClientID        string `json:"client_id" form:"client_id"`
	LtiDeploymentID string `json:"lti_deployment_id" form:"lti_deployment_id"`
}

type LoginSession struct {
	State     string
	Nonce     string
	Issuer    string
	CreatedAt time.Time
}

type NRPSMember struct {
	UserID     string   `json:"user_id"`
	Name       string   `json:"name"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
}

type NRPSClaim struct {
	ContextMembershipsURL string   `json:"context_memberships_url"`
	ServiceVersions       []string `json:"service_versions"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
