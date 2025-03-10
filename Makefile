include .env

.PHONY: test
test:
	@go test -cover ./... -coverprofile=cover.out
	@go tool cover -html=cover.out -o cover.html

.PHONY: sqlc
sqlc:
	@atlas schema inspect -u ${DATABASE_URL} --format '{{ sql . }}' > schema.sql
	@sqlc generate

.PHONY: atlas-schema-inspect
atlas-schema-inspect:
	@atlas schema inspect -u ${DATABASE_URL}

.PHONY: atlas-schema-apply
atlas-schema-apply:
	@atlas schema apply --url ${DATABASE_URL} --to file://schema.sql --dev-url "docker://postgres/15/dev"

