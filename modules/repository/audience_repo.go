package repository

import (
	"audience-export-lambda/modules/entity"
	"context"
	"database/sql"
)

type AudienceRepo struct {
	db *sql.DB
}

func NewAudienceRepo(db *sql.DB) *AudienceRepo {
	return &AudienceRepo{
		db: db,
	}
}

func (r *AudienceRepo) FindByProjectID(ctx context.Context, projectID string) ([]entity.AudienceRow, error) {
	query := `
		SELECT
			a.project_id,
			b.retailer_code,
			b.retailer_name,
			a.product_flagging,
			p.name  AS product_1_name,
			p.price AS product_1_price
		FROM audience a
		JOIN product  p ON p.id = a.product_id
		JOIN business b ON b.id = a.business_id
		WHERE a.project_id = $1;
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []entity.AudienceRow
	for rows.Next() {
		var row entity.AudienceRow
		if err := rows.Scan(
			&row.ProjectID,
			&row.RetailerCode,
			&row.RetailerName,
			&row.ProductFlagging,
			&row.ProductName,
			&row.ProductPrice,
		); err != nil {
			return nil, err
		}

		results = append(results, row)
	}

	return results, rows.Err()
}
