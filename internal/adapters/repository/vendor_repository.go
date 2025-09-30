package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

type vendorRepositoryPG struct {
	q *db.Queries
}

func NewVendorRepositoryPG(q *db.Queries) *vendorRepositoryPG {
	return &vendorRepositoryPG{q: q}
}

func (r *vendorRepositoryPG) Create(ctx context.Context, arg db.CreateVendorParams) (db.Vendor, error) {
	return r.q.CreateVendor(ctx, arg)
}

func (r *vendorRepositoryPG) Get(ctx context.Context, id uuid.UUID) (db.Vendor, error) {
	return r.q.GetVendor(ctx, id)
}

func (r *vendorRepositoryPG) Update(ctx context.Context, arg db.UpdateVendorParams) (db.Vendor, error) {
	return r.q.UpdateVendor(ctx, arg)
}

func (r *vendorRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteVendor(ctx, id)
}

func (r *vendorRepositoryPG) List(ctx context.Context, limit, offset int32) ([]db.Vendor, error) {
	return r.q.ListVendors(ctx, db.ListVendorsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *vendorRepositoryPG) Search(ctx context.Context, query string, limit, offset int32) ([]db.Vendor, error) {
	return r.q.SearchVendors(ctx, db.SearchVendorsParams{
		Column1: sql.NullString{String: query, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
}
