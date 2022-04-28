# Build
FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /imageresizer


# Deploy
FROM alpine

WORKDIR /app

COPY --from=build /imageresizer imageresizer

ENTRYPOINT ["/app/imageresizer"]