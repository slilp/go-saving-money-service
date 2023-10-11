run:
	go run main.go
generate: 
	go generate	./...
test:
	go	test	./...	-cover
build:
	go	build	main.go
