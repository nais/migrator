.PHONY: migrator

migrator:
	go build -o migrator cmd/migrator/migrator.go
