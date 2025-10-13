.PHONY: lint test
lint:
	go vet -vettool=./statictest.exe ./...

test:
	shortenertest.exe -test.v -test.run=^TestIteration4$ -binary-path=cmd/shortener/shortener