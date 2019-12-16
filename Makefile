.PHONY: native linux windows darwin

native:
	go build -o migrator cmd/migrator/migrator.go

linux:
	GOOS=linux GOARCH=amd64 go build -o migrator-linux-amd64 cmd/migrator/migrator.go

windows:
	GOOS=windows GOARCH=amd64 go build -o migrator.exe cmd/migrator/migrator.go

darwin:
	GOOS=darwin GOARCH=amd64 go build -o migrator-darwin-amd64 cmd/migrator/migrator.go
