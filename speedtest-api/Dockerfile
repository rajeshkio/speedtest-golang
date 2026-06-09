FROM golang:1.26.4-alpine
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/app ./...
EXPOSE 8010
CMD ["app"]
