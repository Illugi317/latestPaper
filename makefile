all: clean linux 

clean:
	-rm -rf ./bin

linux:
	go build -o bin/LatestPaper main.go
