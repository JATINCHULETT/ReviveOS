module github.com/reviveos/services

go 1.25.6

replace github.com/reviveos/packages/types => ../packages/types

replace github.com/reviveos/packages/recovery => ../packages/recovery

replace github.com/reviveos/schemas => ../packages/schemas

require (
	github.com/jackc/pgx/v5 v5.10.0
	github.com/reviveos/packages/recovery v0.0.0-00010101000000-000000000000
	github.com/reviveos/packages/types v0.0.0-00010101000000-000000000000
	github.com/reviveos/schemas v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)
