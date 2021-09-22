all: init build

init:
	@echo "Build all commands for linux, darwin platform"

build:
	@echo "Build for Linux Plaform(386)"
	@echo "Build main.go file"
	env GOOS=linux GOARCH=386 go build -o server-host-automation-linux-386 main.go

	@echo "Build for Linux Plaform(amd64)"
	@echo "Build main.go file"
	env GOOS=linux GOARCH=amd64 go build -o server-host-automation-linux-amd64 main.go

# 	Drop support https://golang.org/doc/go1.15
#	@echo "Build for Darwin Plaform(386)"
#	@echo "Build main.go file"
#	env GOOS=darwin GOARCH=386 go build -o server-host-automation-darwin-386 main.go

	@echo "Build for Darwin Plaform(amd64)"
	@echo "Build main.go file"
	env GOOS=darwin GOARCH=amd64 go build -o server-host-automation-darwin-amd64 main.go