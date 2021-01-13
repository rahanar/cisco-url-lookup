FROM golang:1.15-alpine AS builder

WORKDIR /url-lookup/
COPY cmd/main.go /url-lookup/
RUN CGO_ENABLED=0 go build -o /bin/webserver

FROM scratch
COPY --from=builder /bin/webserver /bin/webserver
ENTRYPOINT ["/bin/webserver"]