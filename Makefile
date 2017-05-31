all:
	go get -d ./... && go build -v -o torrentfs cmd/main.go

clean:
	rm -f torrentfs
