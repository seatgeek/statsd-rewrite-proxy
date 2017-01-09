APP_ENV ?= development
APP_NAME ?= statsd-rewrite-proxy
APP_VERSION ?= latest

.PHONY: install
install:
	go get github.com/kardianos/govendor

.PHONY: build
build: install
	govendor sync
	go install

.PHONY: deploy-build
deploy-build: deploy-docker-build deploy-docker-push

.PHONY: deploy-docker-build
deploy-docker-build:
	docker run \
		--rm \
		-v ${PWD}:/go/src/github.com/bownty/statsd-rewrite-proxy \
		--net=host \
		golang:1.7-wheezy \
		bash -c "cd /go/src/github.com/bownty/statsd-rewrite-proxy ; make deploy-build-internal"

.PHONY: deploy-docker-push
deploy-docker-push: deploy-docker-build
	curl \
		-X POST \
		-H "Authorization: ${GOOGLE_CLOUD_AUTH_KEY}" \
		"https://www.googleapis.com/upload/storage/v1/b/bownty-deploy-artifacts/o?uploadType=media&name=${APP_NAME}/${APP_ENV}/${APP_VERSION}/statsd-rewrite-proxy" \
		--data-binary @statsd-rewrite-proxy
	rm statsd-rewrite-proxy

.PHONY: deploy-build-internal
deploy-build-internal: install build
	mv ${GOPATH}/bin/statsd-rewrite-proxy .