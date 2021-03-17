all: app

app:
	GOOS=linux GOARCH=amd64 go build -o bugfender-integration-elasticsearch-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o bugfender-integration-elasticsearch-darwin-amd64

clean:
	-rm -f bugfender-integration-elasticsearch-linux-amd64 bugfender-integration-elasticsearch-darwin-amd64

