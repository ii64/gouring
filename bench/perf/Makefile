alL: run

N := 10000000
PERF_OPTS := -n $(N) -noti 5000000

GCFLAGS := $(GCFLAGS)
#GCFLAGS += -m=2
#GCFLAGS += -l=4

build:
	go build -gcflags="$(GCFLAGS)" .

run: build
	./perf $(PERF_OPTS) -pprofCpu pprof-nonsqpoll.cpu
	./perf -sqpoll $(PERF_OPTS) -pprofCpu pprof-sqpoll.cpu

pprof:
	go tool pprof -http=:9001 $(P)