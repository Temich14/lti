package service

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"crypto/rsa"
	"encoding/base64"
	"log/slog"
	"math/big"
)

type JwksService struct {
	cfg          *config.Config
	PrivateKey   *rsa.PrivateKey
	PrivateKeyID string
	log          *slog.Logger
}

func NewJwksService(cfg *config.Config, key *rsa.PrivateKey) *JwksService {
	return &JwksService{cfg: cfg, PrivateKey: key, PrivateKeyID: cfg.OAuthConfig.KeyID, log: slog.Default()}
}

func (s *JwksService) PrepareJwks() *domain.JWKS {
	s.log.Info("jwks.prepare.start", "kid", s.PrivateKeyID)
	pub := s.PrivateKey.PublicKey

	jwk := buildJWK(&pub, s.PrivateKeyID)

	jwks := domain.JWKS{
		Keys: []domain.JWK{jwk},
	}
	s.log.Info("jwks.prepare.ok", "kid", s.PrivateKeyID, "key_count", len(jwks.Keys))
	return &jwks
}
func buildJWK(pub *rsa.PublicKey, kid string) domain.JWK {

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
