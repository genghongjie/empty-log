CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o empty-log empty-log.go

docker build -t yinjianxia/empty-log:0.1 .

docker push yinjianxia/empty-log:0.1