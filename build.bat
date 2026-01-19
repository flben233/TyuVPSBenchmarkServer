@echo off
set CGO_ENABLED=0

set GOOS=windows
set GOARCH=amd64
go build -o bin/server-windows.exe ./cmd/server
go build -o bin/exporter-windows.exe ./cmd/exporter

set GOOS=linux
set GOARCH=amd64
go build -o bin/server-linux ./cmd/server
go build -o bin/exporter-linux ./cmd/exporter

echo Build complete
