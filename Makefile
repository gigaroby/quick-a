.PHONY: help all clean run frontend deps

all: clean quick-a frontend

quick-a:
	go build github.com/gigaroby/quick-a

frontend:
	$(MAKE) -C html-root frontend.js

clean:
	rm -f quick-a
	$(MAKE) -C html-root clean

run: all
	./quick-a

deps:
	go get -u github.com/gopherjs/gopherjs
	go get -u honnef.co/go/js/dom
