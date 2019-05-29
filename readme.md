Set CGO_ENABLED=1 for sqlite
Command to compile:
CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' .