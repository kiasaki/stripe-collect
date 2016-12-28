run: build
	./stripe-collect

build:
	go build -o stripe-collect .
