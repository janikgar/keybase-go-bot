mock:
	mockery --all

test: mock
	go test . -coverprofile cover.out

cover: test
	go tool cover -html cover.out

build: test
	goreleaser build --snapshot --rm-dist --single-target

release: test
	goreleaser build --snapshot --rm-dist

run:
	go run .