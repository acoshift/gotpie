clean:
	rm gotpie

build:
	go build -o gotpie main.go

deps:
	go get

install:
	go install github.com/acoshift/gotpie
