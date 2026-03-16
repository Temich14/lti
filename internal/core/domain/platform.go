package domain

import "github.com/google/uuid"

type Platform struct {
	ID              uuid.UUID
	Name            string
	Issuer          string
	ClientID        string
	DeploymentID    string
	AuthUrl         string
	TokenUrl        string
	RegistrationUrl string
	JwksUrl         string
}
