run: build
	bash -c "source .env; ./stripe-collect"

build:
	go build -o stripe-collect .
