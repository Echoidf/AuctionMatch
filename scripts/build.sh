#! /bin/bash
# build for linux
go build -o auctionMatch main.go

# build for windows
# GOOS=windows GOARCH=amd64 go build -o auctionMatch.exe main.go