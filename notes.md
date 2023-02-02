## Migrations

- `migrate create -seq -ext=.sql -dir=./migrations [name]`
- `migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up`
