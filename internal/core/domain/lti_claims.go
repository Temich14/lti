package domain

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
)

// ClaimsUserSub returns the LMS subject identifier from LTI JWT claims (`sub`).
func ClaimsUserSub(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub
	}
	return ""
}

// ClaimsContextID returns https://purl.imsglobal.org/spec/lti/claim/context `id`.
func ClaimsContextID(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}
	raw, ok := claims["https://purl.imsglobal.org/spec/lti/claim/context"]
	if !ok {
		return ""
	}
	ctx, ok := raw.(map[string]interface{})
	if !ok {
		return ""
	}
	if id, ok := ctx["id"].(string); ok {
		return id
	}
	return ""
}

// ClaimsResourceLinkID returns https://purl.imsglobal.org/spec/lti/claim/resource_link `id` when present.
func ClaimsResourceLinkID(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}
	raw, ok := claims["https://purl.imsglobal.org/spec/lti/claim/resource_link"]
	if !ok {
		return ""
	}
	rl, ok := raw.(map[string]interface{})
	if !ok {
		return ""
	}
	if id, ok := rl["id"].(string); ok {
		return id
	}
	return ""
}

// ClaimsNRPSMembershipsURL returns the NRPS context memberships URL when the claim exists and is valid.
func ClaimsNRPSMembershipsURL(claims jwt.MapClaims) string {
	if claims == nil {
		return ""
	}
	nrpsClaimRaw, ok := claims["https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice"]
	if !ok {
		return ""
	}
	nrpsJSON, err := json.Marshal(nrpsClaimRaw)
	if err != nil {
		return ""
	}
	var nrps NRPSClaim
	if err := json.Unmarshal(nrpsJSON, &nrps); err != nil {
		return ""
	}
	return nrps.ContextMembershipsURL
}

// ClaimsDeepLinkingSettings parses deep linking settings or returns an empty defaults object.
func ClaimsDeepLinkingSettings(claims jwt.MapClaims) *DeepLinkingSettings {
	if claims == nil {
		return &DeepLinkingSettings{}
	}
	raw, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"]
	if !ok {
		return &DeepLinkingSettings{}
	}
	settingsMap, ok := raw.(map[string]interface{})
	if !ok {
		return &DeepLinkingSettings{}
	}
	settings := &DeepLinkingSettings{}
	if v, ok := settingsMap["accept_types"].([]interface{}); ok {
		settings.AcceptTypes = stringSliceFromAny(v)
	}
	if v, ok := settingsMap["accept_presentation_document_targets"].([]interface{}); ok {
		settings.AcceptPresentationDocumentTargets = stringSliceFromAny(v)
	}
	if v, ok := settingsMap["accept_multiple"].(bool); ok {
		settings.AcceptMultiple = v
	}
	if v, ok := settingsMap["auto_create"].(bool); ok {
		settings.AutoCreate = v
	}
	if v, ok := settingsMap["title"].(string); ok {
		settings.Title = v
	}
	if v, ok := settingsMap["text"].(string); ok {
		settings.Text = v
	}
	if v, ok := settingsMap["deep_link_return_url"].(string); ok {
		settings.DeepLinkReturnURL = v
	}
	return settings
}

func stringSliceFromAny(input []interface{}) []string {
	out := make([]string, 0, len(input))
	for _, v := range input {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
