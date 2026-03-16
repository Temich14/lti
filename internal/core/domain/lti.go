package domain

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type DeepLinkingRequest struct {
	DeepLinkingSettings DeepLinkingSettings    `json:"https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"`
	DeploymentID        string                 `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id"`
	MessageType         string                 `json:"https://purl.imsglobal.org/spec/lti/claim/message_type"`
	Version             string                 `json:"https://purl.imsglobal.org/spec/lti/claim/version"`
	Roles               []string               `json:"https://purl.imsglobal.org/spec/lti/claim/roles"`
	Context             *LTIContext            `json:"https://purl.imsglobal.org/spec/lti/claim/context,omitempty"`
	Custom              map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/custom,omitempty"`
}

type DeepLinkingSettings struct {
	AcceptTypes                       []string `json:"accept_types"`
	AcceptMediaTypes                  []string `json:"accept_media_types"`
	AcceptPresentationDocumentTargets []string `json:"accept_presentation_document_targets"`
	AcceptMultiple                    bool     `json:"accept_multiple"`
	AutoCreate                        bool     `json:"auto_create"`
	Title                             string   `json:"title,omitempty"`
	Text                              string   `json:"text,omitempty"`
	Data                              string   `json:"data,omitempty"`
	DeepLinkReturnURL                 string   `json:"deep_link_return_url"`
}

type DeepLinkResponse struct {
	JWT string `json:"jwt"`
}

type LTIContext struct {
	ID    string   `json:"id"`
	Type  []string `json:"type"`
	Title string   `json:"title"`
	Label string   `json:"label"`
}
type ContentItem struct {
	Type         string                 `json:"type"`
	URL          string                 `json:"url,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Text         string                 `json:"text,omitempty"`
	Icon         *Icon                  `json:"icon,omitempty"`
	Thumbnail    *Thumbnail             `json:"thumbnail,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
	WindowTarget string                 `json:"windowTarget,omitempty"`
	IFrame       *IFrame                `json:"iframe,omitempty"`
	LineItem     *LineItem              `json:"lineItem,omitempty"`
	Available    *Availability          `json:"available,omitempty"`
	Submission   *Submission            `json:"submission,omitempty"`
}

type Icon struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type IFrame struct {
	Src    string `json:"src"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Availability struct {
	StartAt *time.Time `json:"startAt,omitempty"`
	EndAt   *time.Time `json:"endAt,omitempty"`
}

type Submission struct {
	StartAt *time.Time `json:"startAt,omitempty"`
	EndAt   *time.Time `json:"endAt,omitempty"`
}

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
	Type                                         string   `json:"type" form:"type"`
	TargetLinkUri                                string   `json:"target_link_uri" form:"target_link_uri"`
	Label                                        string   `json:"label" form:"label"`
	Placements                                   []string `json:"placements" form:"placements"`
	DeepLinkingAcceptTypes                       []string `json:"deep_linking_accept_types,omitempty" form:"deep_linking_accept_types"`
	DeepLinkingAcceptMediaTypes                  []string `json:"deep_linking_accept_media_types,omitempty" form:"deep_linking_accept_media_types"`
	DeepLinkingAcceptPresentationDocumentTargets []string `json:"deep_linking_accept_presentation_document_targets,omitempty" form:"deep_linking_accept_presentation_document_targets"`
	DeepLinkingAutoCreate                        *bool    `json:"deep_linking_auto_create,omitempty" form:"deep_linking_auto_create"`
	DeepLinkingTitle                             string   `json:"deep_linking_title,omitempty" form:"deep_linking_title"`
	DeepLinkingText                              string   `json:"deep_linking_text,omitempty" form:"deep_linking_text"`
}

type DeepLinkingSession struct {
	Settings  *DeepLinkingSettings
	Platform  *Platform
	UserID    string
	ContextID string
	CreatedAt time.Time
}

// Константы для типов контента
const (
	ContentTypeLtiResourceLink = "ltiResourceLink"
	ContentTypeLink            = "link"
	ContentTypeHTML            = "html"
	ContentTypeImage           = "image"
	ContentTypeFile            = "file"
)

// Константы для placement'ов
const (
	PlacementCourseAssignmentsMenu = "course_assignments_menu"
	PlacementAssignmentEditMenu    = "assignment_edit_menu"
	PlacementLinkEmbedButton       = "link_embed_button"
	PlacementModuleQuicklinksMenu  = "module_quicklinks_menu"
)

type RegistrationResponse struct {
}

type LoginRequest struct {
	Iss string `json:"iss" form:"iss"`
	// TargetLinkURI is the LTI 1.3 OIDC login initiation parameter.
	TargetLinkURI string `json:"target_link_uri" form:"target_link_uri"`
	// TargetLingUri kept for backward compatibility with earlier typo.
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
