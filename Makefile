all: confex

confex: *.go
	@go build 

clean:
	@rm -f confex