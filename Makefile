build: deps
	go build ./...
	go install ./...
	go generate
	go build ./...
	go install ./...

deps:
	go get golang.org/x/tools/cmd/stringer

test:
	go test -cover ./...

cover: coverage
coverage:
	go test -coverprofile=speed.coverage
	go tool cover -html=speed.coverage

	go test -coverprofile=bytebuffer.coverage ./bytebuffer/
	go tool cover -html=bytebuffer.coverage

bench: benchmark
benchmark:
	go test -bench ./...
