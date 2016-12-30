# Setup
```bash
# $GOPATH should be properly set
# see https://golang.org/doc/code.html for info

~ $ mkdir -p $GOPATH/src/github.com/gigaroby/
~ $ cd $GOPATH/src/github.com/gigaroby
~ $ git clone git@github.com:gigaroby/quick-a.git
# install project deps
~ $ go get github.com/gigaroby/quick-a
# install go to js compiler
~ $ go get -u github.com/gopherjs/gopherjs

~ $ cd quick-a
~ $ go build  # builds quick-a binary
~ $ cd html-root
~ $ gopherjs build -m github.com/gigaroby/quick-a/frontend/  # build js

~ $ cd ..
~ $ ./quick-a 


```