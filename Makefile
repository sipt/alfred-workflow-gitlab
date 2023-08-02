build:
	GOOS=darwin GOARCH=amd64 go build -o bin/awgitlab.amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/awgitlab.arm64 main.go
	mkdir -p bin
	lipo -create -output ./bin/awgitlab bin/awgitlab.amd64 bin/awgitlab.arm64