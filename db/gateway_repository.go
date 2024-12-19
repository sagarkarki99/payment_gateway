package db

import (
	"database/sql"
	"payment-gateway/internal/models"
)

type GatewayRepository interface {
	// GetAvailableGateways returns all gateways available for the given country
	GetAvailableGateways(countryID int) ([]*Gateway, error)
}

type gatewayRepository struct {
	db *sql.DB
}

func NewGatewayRepository(db *sql.DB) GatewayRepository {
	return &gatewayRepository{db: db}
}

func (r *gatewayRepository) GetAvailableGateways(countryID int) ([]*Gateway, error) {
	// Get all gateways that support this country
	query := `
		SELECT g.id, g.name, g.data_format_supported, g.created_at, g.updated_at
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		WHERE gc.country_id = $1`

	rows, err := r.db.Query(query, countryID)
	if err != nil {
		return nil, models.NewServiceError(
			models.ErrorCodeUnknown,
			"Failed to fetch available gateways",
		)
	}
	defer rows.Close()

	var gateways []*Gateway
	for rows.Next() {
		var gateway Gateway
		err := rows.Scan(
			&gateway.ID,
			&gateway.Name,
			&gateway.DataFormatSupported,
			&gateway.CreatedAt,
			&gateway.UpdatedAt,
		)
		if err != nil {
			return nil, models.NewServiceError(
				models.ErrorCodeUnknown,
				"Failed to scan gateway information",
			)
		}
		gateways = append(gateways, &gateway)
	}

	if err = rows.Err(); err != nil {
		return nil, models.NewServiceError(
			models.ErrorCodeUnknown,
			"Error iterating through gateways",
		)
	}

	if len(gateways) == 0 {
		return nil, models.NewServiceError(
			models.ErrorCodeNotFound,
			"No gateways available for this country",
		)
	}

	return gateways, nil
}
