ifneq (oneshell, $(findstring oneshell, $(.FEATURES)))
  $(error GNU make 3.82 or later is required)
endif

all: dependencies ${GOPATH}/bin/hercules

test: all
	go test gopkg.in/src-d/hercules.v3

dependencies: ${GOPATH}/src/gopkg.in/bblfsh/client-go.v2

.ONESHELL:
${GOPATH}/src/gopkg.in/bblfsh/client-go.v2:
	go get -v gopkg.in/bblfsh/client-go.v2/... || true
	cd $$GOPATH/src/gopkg.in/bblfsh/client-go.v2
	make dependencies

.ONESHELL:
${GOPATH}/bin/hercules: *.go cmd/hercules/*.go rbtree/*.go stdout/*.go toposort/*.go pb/*.go
	cd ${GOPATH}/src/gopkg.in/src-d/hercules.v3
	go get -ldflags "-X gopkg.in/src-d/hercules.v3.GIT_HASH=$$(git rev-parse HEAD)" gopkg.in/src-d/hercules.v3/cmd/hercules
	${GOPATH}/bin/hercules -version
