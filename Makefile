version=fake

binaries:
		@go build -o bin/linux/x86_64/gophie && echo "Built bin for linux x86_64"
		@env GOOS=windows GOARCH=amd64; go build -o bin/windows/64-bit/gophie && echo "Built bin for windows 64-bit"
		@env GOOS=windows GOARCH=386; go build -o bin/windows/32-bit/gophie && echo "Built bin for windows 32-bit"


ifneq ($(findstring fake, $(version)), fake)
upgrade:
	 @echo ">>>  Recreating version_num.go"
	 @echo 'package cmd\n\nconst Version = "$(version)"' > cmd/version_num.go
	 @go run main.go version
	 @echo ">>>  Creating Git Tag v$(version)"
	 @git tag v$(version)
	 @echo ">>>  Pushing Tag to Remote"
	 @git push origin v$(version)
else
upgrade:
	@echo "Version not set - use syntax \`make upgrade version=0.x.x\`"
endif	 
	
