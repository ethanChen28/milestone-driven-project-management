FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN go test ./...
RUN go build -o /goal-manager ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /goal-manager /usr/local/bin/goal-manager
EXPOSE 8080
CMD ["goal-manager"]
