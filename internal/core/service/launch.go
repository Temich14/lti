package service

import (
	"LTICore/internal/core/domain"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"math/big"
	"net/http"
	"time"
)

func (s *LTIService) Launch(ctx context.Context, idToken, state string) ([]domain.NRPSMember, error) {

	start := time.Now()

	defer func() {
		s.metrics.ObserveLaunchDuration(time.Since(start))
	}()

	session, err := s.validateSession(ctx, state)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	unverifiedToken, err := s.parseUnverifiedToken(idToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	kid, err := s.extractKID(unverifiedToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	platform, err := s.resolvePlatform(ctx, unverifiedToken)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	pubKey, err := s.fetchPlatformPublicKey(platform.JwksUrl, kid)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	claims, err := s.verifyToken(idToken, pubKey)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	if err := s.validateClaims(claims, session, platform); err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	members, err := s.fetchNRPSMembers(ctx, claims, platform)
	if err != nil {
		s.metrics.IncLaunchError()
		return nil, err
	}

	s.metrics.IncLaunchSuccess()

	return members, nil
}

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

	iss := claims["iss"].(string)
	aud := claims["aud"].(string)

	platform, err := s.platformRepo.GetByIssuerAndClientID(ctx, iss, aud)
	if err != nil || platform == nil {
		return nil, fmt.Errorf("unknown platform")
	}

	return platform, nil
}

func (s *LTIService) fetchPlatformPublicKey(jwksURL, kid string) (*rsa.PublicKey, error) {

	start := time.Now()

	defer func() {
		s.metrics.ObserveJWKSFetch(time.Since(start))
	}()

	jwks, err := fetchJWKS(jwksURL)
	if err != nil {
		s.metrics.IncJWKSFetchError()
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	for _, key := range jwks.Keys {

		if key.Kid == kid {
			return buildRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("matching key not found")
}

func (s *LTIService) verifyToken(idToken string, pubKey *rsa.PublicKey) (jwt.MapClaims, error) {

	token, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {

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

	now := time.Now()

	if !claims.VerifyExpiresAt(now.Unix(), true) {
		return fmt.Errorf("token expired")
	}

	if !claims.VerifyIssuedAt(now.Unix(), true) {
		return fmt.Errorf("token used before issued")
	}

	if !claims.VerifyNotBefore(now.Unix(), true) {
		return fmt.Errorf("token not valid yet")
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

	defer func() {
		s.metrics.ObserveNRPSRequest(time.Since(start))
	}()

	nrpsClaimRaw, ok := claims["https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice"]
	if !ok {
		return nil, nil
	}

	nrpsJSON, _ := json.Marshal(nrpsClaimRaw)

	var nrps domain.NRPSClaim

	if err := json.Unmarshal(nrpsJSON, &nrps); err != nil {
		s.metrics.IncNRPSError()
		return nil, fmt.Errorf("invalid NRPS claim: %w", err)
	}

	if nrps.ContextMembershipsURL == "" {
		s.metrics.IncNRPSError()
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
		return nil, err
	}

	return s.nrpsClient.GetMembers(ctx, nrps.ContextMembershipsURL, accessToken)
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
