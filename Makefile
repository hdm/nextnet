ALL:
	@go get github.com/mitchellh/gox && \
	go get -u ./... && \
	go fmt ./... && \
	go vet ./... && \
	go build ./... && \
	go install ./...


.PHONY: ALL
