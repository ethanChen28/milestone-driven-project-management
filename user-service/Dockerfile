FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . ./
RUN go build -o /out/user-service ./cmd/user-service

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/user-service /app/user-service
EXPOSE 8090
CMD ["/app/user-service"]
