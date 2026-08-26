module github.com/reviveos/api

go 1.25.6

require (
	github.com/jackc/pgx/v5 v5.10.0
	github.com/reviveos/packages/types v0.0.0-00010101000000-000000000000
	github.com/reviveos/schemas v0.0.0-00010101000000-000000000000
	github.com/reviveos/utils v0.0.0-00010101000000-000000000000
)

replace github.com/reviveos/packages/types => ../../packages/types

replace github.com/reviveos/types => ../../packages/types

replace github.com/reviveos/schemas => ../../packages/schemas

replace github.com/reviveos/utils => ../../packages/utils
