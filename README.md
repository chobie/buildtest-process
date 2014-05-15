Report Bot
==========

reinterpret buildtest-process program.

https://github.com/php/web-qa/blob/master/buildtest-process.php

How to build
============

```
git clone https://github.com/chobie/buildtest-process
cd buildtest-process
export GOPATH=`pwd`
go get code.google.com/p/goauth2/oauth
go get github.com/google/go-github/github
go build -o buildtest-process src/main.go

export HOST=127.0.0.1
export PORT=9999
export TOKEN=<YOUR-GITHUB-API-TOKEN>
export ORGANIZATION=php-git-bot
export REPOSITORY=test

# run as foreground
./buildtest-process -foreground

# run as a daemon
./buildtest-process
```