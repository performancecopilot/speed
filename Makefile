build: clean
	go build ./...
	go install ./...
	go generate
	go build ./...
	go install ./...

clean:
	git clean -Xf

deps:
	go get -u golang.org/x/tools/cmd/stringer
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	gometalinter ./... --vendor --deadline=10000s --dupl-threshold=150 --disable=gas	

test: 
	go test ./...

cover: coverage
coverage:
	go test -coverprofile=speed.coverage
	go tool cover -html=speed.coverage

	go test -coverprofile=bytebuffer.coverage ./bytebuffer/
	go tool cover -html=bytebuffer.coverage

bench: benchmark
benchmark:
	go test -bench ./...
