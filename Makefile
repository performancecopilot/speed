build: deps
	go generate
	go build

deps:
	go get golang.org/x/tools/cmd/stringer

test:
	go test -cover

cover: coverage
coverage:
	go test -coverprofile=c.out
	go tool cover -html=c.out

bench: benchmark
benchmark:
	go test -bench .
