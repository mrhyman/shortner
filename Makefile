.PHONY: lint test
lint:
	go vet -vettool=./statictest.exe ./...

test:
	shortenertest.exe -test.v -test.run=^TestIteration4$ -binary-path=cmd/shortener/shortener

mc:
	migrate create -ext sql -dir ./migrations -seq ${NAME}

bench:
	go test -bench=. -benchmem -run=^$$ ./internal/handler/ 2>/dev/null
