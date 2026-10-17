FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /promise-server ./cmd/promise-server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /promise-server /promise-server
VOLUME ["/data"]
ENV PROMISE_LISTEN_ADDR=:8080
ENV PROMISE_DATA_DIR=/data
EXPOSE 8080
ENTRYPOINT ["/promise-server"]
