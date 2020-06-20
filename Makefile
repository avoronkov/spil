.PHONY: all plugins

all: libraries plugins test
	go build

# Install go-bindata with `go get -u github.com/a-urth/go-bindata/...`
libraries:
	go-bindata --nometadata --pkg library -o library/library.go library/builtin/... library/std/...

plugins:
	cd plugins/io && go build --buildmode=plugin
	cd plugins/drawer && go build --buildmode=plugin

test:
	go test ./...
