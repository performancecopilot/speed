test:
	go test -cover

cover: coverage
coverage:
	go test -coverprofile=c.out
	go tool cover -html=c.out

bench: benchmark
benchmark:
	go test -bench .