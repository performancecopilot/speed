build:
	go build ./...
	go install ./...

clean:
	git clean -Xf

clean_string:
	rm *_string.go
	rm mmvdump/*_string.go

gen:
	go generate

deps:
	go get -u golang.org/x/tools/cmd/stringer
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.15.0
	bin/golangci-lint --version

lint:
	bin/golangci-lint run

test:
	go test -v ./...

race:
	go test -v -race ./...

cover: coverage
coverage:
	go test -v -coverprofile=speed.coverage
	go tool cover -html=speed.coverage

	go test -v -coverprofile=bytebuffer.coverage ./bytebuffer/
	go tool cover -html=bytebuffer.coverage

	go test -v -coverprofile=mmvdump.coverage ./mmvdump/
	go tool cover -html=mmvdump.coverage

bench: benchmark
benchmark:
	go test -v -bench ./...
