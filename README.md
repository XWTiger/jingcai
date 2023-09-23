swag init --parseDependency ./main.go
go build -o jingcai .\main.go

swag init --parseDependency --parseInternal --parseGoList=false --parseDepth=1 .\main.go

CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build .\main.go

CGO_ENABLED=0 GOOS=windows  GOARCH=amd64  go  build  main.go