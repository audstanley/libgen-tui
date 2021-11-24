#!/bin/bash
#!/bin/bash
# this will build for all processor types on a linux machine/or/container
GOOS=windows GOARCH=amd64 go build -o build/libgen-amd64.exe main.go
echo "Built: build/libgen-amd64.exe"
GOOS=windows GOARCH=386 go build -o build/libgen-386.exe main.go
echo "Built: build/libgen-386.exe"
GOOS=darwin GOARCH=amd64 go build -o build/libgen-amd64-macos main.go
echo "Built: build/libgen-amd64-macos"
GOOS=linux GOARCH=amd64 go build -o build/libgen-amd64-linux main.go
echo "Built: build/libgen-amd64-linux"
sha256sum build/libgen-amd64.exe      | awk '{print $1}'  > build/sha256-libgen-amd64.exe.checksum
sha256sum build/libgen-386.exe        | awk '{print $1}' > build/sha256-libgen-386.exe.checksum
sha256sum build/libgen-amd64-macos    | awk '{print $1}'  > build/sha256-libgen-amd64-macos.checksum
sha256sum build/libgen-amd64-linux    | awk '{print $1}' > build/sha256-libgen-amd64-linux.checksum
