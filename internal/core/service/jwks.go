package service

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

type JwksService struct {
	cfg          *config.Config
	PrivateKey   *rsa.PrivateKey
	PrivateKeyID string
}

func NewJwksService(cfg *config.Config, key *rsa.PrivateKey) *JwksService {
	return &JwksService{cfg: cfg, PrivateKey: key, PrivateKeyID: cfg.OAuthConfig.KeyID}
}

func (s *JwksService) PrepareJwks() *domain.JWKS {
	pub := s.PrivateKey.PublicKey

	jwk := buildJWK(&pub, s.PrivateKeyID)

	jwks := domain.JWKS{
		Keys: []domain.JWK{jwk},
	}
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
