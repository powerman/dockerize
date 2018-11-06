.PHONY : dockerize clean fmt

TAG:=`git describe --abbrev=5 --dirty --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

all: dockerize

dockerize:
	echo "Building dockerize"
	go install -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist
	rm -f dockerize-*.tar.gz

dist: dist-clean
	mkdir -p dist/linux/amd64 && GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags "$(LDFLAGS)" -o dist/linux/amd64/dockerize
	mkdir -p dist/linux/386 && GOOS=linux GOARCH=386 go build -a -tags netgo -ldflags "$(LDFLAGS)" -o dist/linux/386/dockerize
	mkdir -p dist/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags "$(LDFLAGS)" -o dist/darwin/amd64/dockerize

release: dist
	tar -cvzf dockerize-linux-amd64-$(TAG).tar.gz -C dist/linux/amd64 dockerize
	tar -cvzf dockerize-linux-386-$(TAG).tar.gz -C dist/linux/386 dockerize
	tar -cvzf dockerize-darwin-amd64-$(TAG).tar.gz -C dist/darwin/amd64 dockerize
