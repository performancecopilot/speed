build:
	go build ./...
	go install ./...

clean:
	git clean -Xf
	rm *_string.go

gen:
	go generate

deps:
	go get -u golang.org/x/tools/cmd/stringer
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	gometalinter ./... --vendor --deadline=10000s --dupl-threshold=100 --disable=interfacer	--disable=gas

test: 
	go test ./...

race:
	go test -race ./...

cover: coverage
coverage:
	go test -coverprofile=speed.coverage
	go tool cover -html=speed.coverage

	go test -coverprofile=bytebuffer.coverage ./bytebuffer/
	go tool cover -html=bytebuffer.coverage

bench: benchmark
benchmark:
	go test -bench ./...
