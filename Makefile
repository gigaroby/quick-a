.PHONY: help all clean run frontend deps

all: clean quick-a frontend

quick-a:
	go build github.com/gigaroby/quick-a

frontend:
	+$(MAKE) -C html-root frontend.js

clean:
	rm -f quick-a
	# http://stackoverflow.com/questions/1139271/makefiles-with-source-files-in-different-directories
	+$(MAKE) -C html-root clean

run: all
	./quick-a

deps:
	go get -u github.com/gopherjs/gopherjs
	go get -u honnef.co/go/js/dom
	go get github.com/gigaroby/quick-a/frontend
