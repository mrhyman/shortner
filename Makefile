.PHONY: lint test
lint:
	go vet -vettool=./statictest.exe ./...

test:
	shortenertest.exe -test.v -test.run=^TestIteration4$ -binary-path=cmd/shortener/shortener

mc:
	migrate create -ext sql -dir ./migrations -seq ${NAME}

bench:
	go test -bench=. -benchmem -run=^$$ ./internal/handler/ 2>/dev/null

bench-base:
	@echo "🧠 Creating base.pprof memory profile..."
	go test -bench=. -benchmem \
		-memprofile=./profiles/base.pprof \
		-run=^$$ ./internal/handler/
	@echo "✅ Base memory profile saved to ./profiles/base.pprof"
	@go tool pprof -top -nodecount=10 ./profiles/base.pprof

bench-result:
	@echo "🧠 Creating result.pprof memory profile..."
	go test -bench=. -benchmem \
		-memprofile=./profiles/result.pprof \
		-run=^$$ ./internal/handler/
	@echo "✅ Result memory profile saved to ./profiles/result.pprof"
	@go tool pprof -top -nodecount=10 ./profiles/result.pprof

bench-diff:
	@echo "🧠 Creating result.pprof memory profile..."
	go test -bench=. -benchmem \
		-memprofile=./profiles/result.pprof \
		-run=^$$ ./internal/handler/
	@echo "✅ Result memory profile saved to ./profiles/result.pprof"
	@go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof 
