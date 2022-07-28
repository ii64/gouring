test:
	go test -c -gcflags=all=-d=checkptr .
	./gouring.test
clean:
	rm -rf *.test