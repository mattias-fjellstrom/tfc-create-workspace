FROM golang:1.20.2-alpine
WORKDIR /app
COPY ./ ./
RUN go build -o /bin/app main.go
ENTRYPOINT ["app"]