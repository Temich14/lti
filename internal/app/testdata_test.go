package app

import (
	"LTICore/internal/config"
	"crypto/rand"
	"crypto/rsa"
)

var testPrivateKey, _ = rsa.GenerateKey(rand.Reader, 2048)

var testConfig = &config.Config{
	LTIConfig: config.LTIConfig{
		Domain:                "tool.example",
		ToolName:              "Test Tool",
		RegistrationClientUri: "https://tool.example/lti/auth/register",
		LoginClientUri:        "https://tool.example/lti/auth/login",
		ApplicationType:       "web",
	},
	OAuthConfig: config.OAuthConfig{
		KeyID:       "test-kid",
		RedirectURI: "https://tool.example/lti/launch",
		JWKSUri:     "https://tool.example/.well-known/jwks.json",
	},
}
