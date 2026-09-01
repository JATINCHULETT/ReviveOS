module github.com/reviveos/seed_data

go 1.25.6

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/reviveos/utils v0.0.0
)

replace github.com/reviveos/utils => ../../packages/utils
