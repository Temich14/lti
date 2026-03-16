package repo

import (
	"LTICore/internal/core/domain"
	"LTICore/internal/infrastructure/db"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PlatformsRepo struct {
	q *db.Queries
}

func NewPlatformRepository(q *db.Queries) *PlatformsRepo {
	return &PlatformsRepo{q: q}
}

func (r *PlatformsRepo) Create(ctx context.Context, platform *domain.Platform) (*domain.Platform, error) {
	arg := db.CreatePlatformParams{
		DeploymentID:         toText(platform.DeploymentID),
		Issuer:               platform.Issuer,
		ClientID:             platform.ClientID,
		AuthEndpoint:         toText(platform.AuthUrl),
		TokenEndpoint:        toText(platform.TokenUrl),
		RegistrationEndpoint: toText(platform.RegistrationUrl),
		JwksUri:              toText(platform.JwksUrl),
	}

	dbModel, err := r.q.CreatePlatform(ctx, arg)
	if err != nil {
		return &domain.Platform{}, err
	}
	domainPlatform := MapToDomain(dbModel)
	return &domainPlatform, nil
}

func (r *PlatformsRepo) GetByIssuerAndClientID(ctx context.Context, issuer, clientID string) (*domain.Platform, error) {
	args := db.GetPlatformByIssuerAndClientIDParams{
		Issuer:   issuer,
		ClientID: clientID,
	}
	dbModel, err := r.q.GetPlatformByIssuerAndClientID(ctx, args)
	if err != nil {
		return &domain.Platform{}, err
	}
	domainPlatform := MapToDomain(dbModel)
	return &domainPlatform, nil
}

func (r *PlatformsRepo) List(ctx context.Context) ([]domain.Platform, error) {
	dbModels, err := r.q.ListPlatforms(ctx)
	if err != nil {
		return nil, err
	}
	domainPlatforms := MapToDomainList(dbModels)
	return domainPlatforms, nil
}

func (r *PlatformsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeletePlatform(ctx, id)
}

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}
func fromText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func MapToDomain(platform db.LtiPlatform) domain.Platform {
	return domain.Platform{
		ID:              platform.ID,
		Issuer:          platform.Issuer,
		ClientID:        platform.ClientID,
		DeploymentID:    fromText(platform.DeploymentID),
		AuthUrl:         fromText(platform.AuthEndpoint),
		TokenUrl:        fromText(platform.TokenEndpoint),
		RegistrationUrl: fromText(platform.RegistrationEndpoint),
		JwksUrl:         fromText(platform.JwksUri),
	}
}
func MapToDomainList(platforms []db.LtiPlatform) []domain.Platform {
	var result []domain.Platform
	for _, platform := range platforms {
		domainPlatform := MapToDomain(platform)
		result = append(result, domainPlatform)
	}
	return result
}
