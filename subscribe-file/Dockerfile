FROM golang:alpine
WORKDIR /go/src/app

# Build Golang outisde container
COPY main .
COPY activity.log .

# Build Golang inside container
# RUN COPY . .
# RUN go mod download -x
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build main.go

ENTRYPOINT ["/go/src/app/main"]
