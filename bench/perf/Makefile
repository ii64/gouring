alL: run

N := 100000
PERF_OPTS := -n $(N) -noti $(N)

build:
	go build .

run: build
	./perf $(PERF_OPTS) -pprofCpu pprof-nonsqpoll.cpu
	./perf -sqpoll $(PERF_OPTS) -pprofCpu pprof-sqpoll.cpu

pprof:
	go tool pprof -http=:9001 $(P)