binaries:
		@go build -o bin/linux/x86_64/gophie && echo "Built bin for linux x86_64"
		@env GOOS=windows GOARCH=amd64; go build -o bin/windows/64-bit/gophie && echo "Built bin for windows 64-bit"
		@env GOOS=windows GOARCH=386; go build -o bin/windows/32-bit/gophie && echo "Built bin for windows 32-bit"
