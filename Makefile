build:
	go build ./...
	go install ./...

clean:
	git clean -Xf
	rm *_string.go
	rm mmvdump/*_string.go

gen:
	go generate

deps:
	go get -u golang.org/x/tools/cmd/stringer
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	gometalinter ./... --vendor --deadline=10000s --dupl-threshold=100 --disable=interfacer	--disable=gas

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
