FROM golang:1.26.1-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/app .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && adduser -D -H app
COPY --from=build /out/app /app
USER app

EXPOSE 8080
CMD ["/app"]