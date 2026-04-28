package service

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func (s *LTIService) Launch(ctx context.Context, idToken, state string) (*domain.LaunchContext, error) {

	start := time.Now()
	s.log.Info("lti.launch.start")

	defer func() {
		s.metrics.ObserveLaunchDuration(time.Since(start))
	}()

	// 1. validate state/session
	session, err := s.validateSession(ctx, state)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// 2. parse unverified token (for metadata only)
	unverifiedToken, err := s.parseUnverifiedToken(idToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("parse token failed: %w", err)
	}

	kid, err := s.extractKID(unverifiedToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("extract kid failed: %w", err)
	}

	// 3. resolve platform
	platform, err := s.resolvePlatform(ctx, unverifiedToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("resolve platform failed: %w", err)
	}

	// 4. fetch key
	pubKey, err := s.fetchPlatformPublicKey(platform.JwksUrl, kid)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("jwks fetch failed: %w", err)
	}

	// 5. verify token
	claims, err := s.verifyToken(idToken, pubKey)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// 6. validate claims
	if err := s.validateClaims(claims, session, platform); err != nil {
		s.metrics.IncLaunchError()
		return nil, fmt.Errorf("claims invalid: %w", err)
	}

	s.metrics.IncLaunchSuccess()

	// 7. build launch context (IMPORTANT)
	return &domain.LaunchContext{
		Platform: platform,
		Session:  session,
		Claims:   claims,
	}, nil
}

//
//func (s *LTIService) Launch(ctx context.Context, idToken, state string) ([]domain.NRPSMember, error) {
//
//	start := time.Now()
//	s.log.Info("lti.launch.start")
//
//	defer func() {
//		s.metrics.ObserveLaunchDuration(time.Since(start))
//	}()
//
//	session, err := s.validateSession(ctx, state)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.validate_session.failed", "err", err)
//		return nil, err
//	}
//
//	unverifiedToken, err := s.parseUnverifiedToken(idToken)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.parse_unverified_token.failed", "err", err)
//		return nil, err
//	}
//
//	kid, err := s.extractKID(unverifiedToken)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.extract_kid.failed", "err", err)
//		return nil, err
//	}
//
//	platform, err := s.resolvePlatform(ctx, unverifiedToken)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.resolve_platform.failed", "err", err)
//		return nil, err
//	}
//
//	pubKey, err := s.fetchPlatformPublicKey(platform.JwksUrl, kid)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.fetch_platform_public_key.failed", "issuer", platform.Issuer, "client_id", platform.ClientID, "err", err)
//		return nil, err
//	}
//
//	claims, err := s.verifyToken(idToken, pubKey)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.verify_token.failed", "issuer", platform.Issuer, "client_id", platform.ClientID, "err", err)
//		return nil, err
//	}
//
//	if err := s.validateClaims(claims, session, platform); err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.validate_claims.failed", "issuer", platform.Issuer, "client_id", platform.ClientID, "err", err)
//		return nil, err
//	}
//
//	members, err := s.fetchNRPSMembers(ctx, claims, platform)
//	if err != nil {
//		s.metrics.IncLaunchError()
//		s.log.Error("lti.launch.fetch_nrps_members.failed", "issuer", platform.Issuer, "client_id", platform.ClientID, "err", err)
//		return nil, err
//	}
//
//	s.metrics.IncLaunchSuccess()
//	s.log.Info("lti.launch.ok", "issuer", platform.Issuer, "client_id", platform.ClientID, "member_count", len(members), "duration_ms", time.Since(start).Milliseconds())
//
//	return members, nil
//}

func (s *LTIService) validateSession(ctx context.Context, state string) (*domain.LoginSession, error) {

	session, err := s.loginSessionRepo.GetAndDelete(ctx, state)
	if err != nil || session == nil {
		return nil, fmt.Errorf("invalid or expired state")
	}

	return session, nil
}

func (s *LTIService) parseUnverifiedToken(idToken string) (*jwt.Token, error) {

	parser := jwt.Parser{}

	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	return token, nil
}

func (s *LTIService) extractKID(token *jwt.Token) (string, error) {

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return "", fmt.Errorf("missing kid in JWT header")
	}

	return kid, nil
}

func (s *LTIService) resolvePlatform(ctx context.Context, token *jwt.Token) (*domain.Platform, error) {

	claims := token.Claims.(jwt.MapClaims)

	iss, _ := claims["iss"].(string)
	aud, _ := claims["aud"].(string)
	if iss == "" || aud == "" {
		return nil, fmt.Errorf("missing iss/aud")
	}

	platform, err := s.platformRepo.GetByIssuerAndClientID(ctx, iss, aud)
	if err != nil || platform == nil {
		return nil, fmt.Errorf("unknown platform")
	}

	return platform, nil
}

func (s *LTIService) fetchPlatformPublicKey(jwksURL, kid string) (*rsa.PublicKey, error) {

	start := time.Now()
	s.log.Info("lti.jwks.fetch.start", "jwks_url", jwksURL, "kid", kid)

	defer func() {
		s.metrics.ObserveJWKSFetch(time.Since(start))
	}()

	jwks, err := fetchJWKS(jwksURL)
	if err != nil {
		s.metrics.IncJWKSFetchError()
		s.log.Error("lti.jwks.fetch.failed", "jwks_url", jwksURL, "err", err)
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	for _, key := range jwks.Keys {

		if key.Kid == kid {
			s.log.Info("lti.jwks.fetch.ok", "jwks_url", jwksURL, "kid", kid, "duration_ms", time.Since(start).Milliseconds())
			return buildRSAPublicKey(key)
		}
	}

	s.metrics.IncJWKSFetchError()
	s.log.Warn("lti.jwks.fetch.no_matching_key", "jwks_url", jwksURL, "kid", kid)
	return nil, fmt.Errorf("matching key not found")
}

func (s *LTIService) verifyToken(idToken string, pubKey *rsa.PublicKey) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("signature validation failed: %w", err)
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (s *LTIService) validateClaims(
	claims jwt.MapClaims,
	session *domain.LoginSession,
	platform *domain.Platform,
) error {

	now := time.Now().UTC().Unix()

	if !claims.VerifyExpiresAt(now, true) {
		return fmt.Errorf("token expired")
	}

	if !claims.VerifyAudience(platform.ClientID, true) {
		return fmt.Errorf("invalid audience")
	}

	if !claims.VerifyIssuer(platform.Issuer, true) {
		return fmt.Errorf("invalid issuer")
	}

	if claims["nonce"] != session.Nonce {
		return fmt.Errorf("invalid nonce")
	}

	return nil
}

func (s *LTIService) fetchNRPSMembers(ctx context.Context, claims jwt.MapClaims, platform *domain.Platform) ([]domain.NRPSMember, error) {

	start := time.Now()
	s.log.Info("lti.nrps.fetch.start", "issuer", platform.Issuer, "client_id", platform.ClientID)

	defer func() {
		s.metrics.ObserveNRPSRequest(time.Since(start))
	}()

	nrpsClaimRaw, ok := claims["https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice"]
	if !ok {
		s.log.Info("lti.nrps.fetch.skipped", "reason", "missing_nrps_claim")
		return nil, nil
	}

	nrpsJSON, _ := json.Marshal(nrpsClaimRaw)

	var nrps domain.NRPSClaim

	if err := json.Unmarshal(nrpsJSON, &nrps); err != nil {
		s.metrics.IncNRPSError()
		s.log.Error("lti.nrps.claim.invalid", "err", err)
		return nil, fmt.Errorf("invalid NRPS claim: %w", err)
	}

	if nrps.ContextMembershipsURL == "" {
		s.metrics.IncNRPSError()
		s.log.Warn("lti.nrps.claim.missing_url")
		return nil, fmt.Errorf("missing NRPS URL")
	}

	accessToken, err := s.ltiClient.GetAccessToken(
		ctx,
		platform,
		"https://purl.imsglobal.org/spec/lti-nrps/scope/contextmembership.readonly",
		s.PrivateKeyID,
		s.PrivateKey,
	)

	if err != nil {
		s.metrics.IncNRPSError()
		s.log.Error("lti.nrps.get_access_token.failed", "err", err)
		return nil, err
	}

	members, err := s.nrpsClient.GetMembers(ctx, nrps.ContextMembershipsURL, accessToken)
	if err != nil {
		s.metrics.IncNRPSError()
		s.log.Error("lti.nrps.get_members.failed", "url", nrps.ContextMembershipsURL, "err", err)
		return nil, err
	}
	s.log.Info("lti.nrps.fetch.ok", "member_count", len(members), "duration_ms", time.Since(start).Milliseconds())
	return members, nil
}

func (s *LTIService) BuildJWK(pub *rsa.PublicKey, kid string) domain.JWK {

	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())

	eBytes := big.NewInt(int64(pub.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return domain.JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   n,
		E:   e,
	}
}

func fetchJWKS(jwksURL string) (*domain.JWKS, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks domain.JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

func buildRSAPublicKey(jwk domain.JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
