FROM golang:1.21-bookworm AS build

WORKDIR ../

COPY . .

RUN go mod download
RUN go build -o main .

EXPOSE 8080

ENTRYPOINT [ "./main" ]