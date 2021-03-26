.PHONY: build
build:
	CGO_ENABLED=0 GO111MODULE=on go build \
		-mod=vendor \
		-buildmode=exe \
		-o cloner \
		github.com/impochi/cloner/cmd/cloner

.PHONY: build-docker
docker-build:
	docker build -t imranpochi/cloner .

.PHONY: test
test:
	go test -mod=vendor -buildmode=exe ./...
