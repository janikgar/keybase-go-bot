mock:
	mockery --all

test: mock
	go test . -coverprofile cover.out

cover: test
	go tool cover -html cover.out

build:
	go build -o keybasebot.exe .

run:
	go run .