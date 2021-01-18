FROM golang:1.15-alpine AS builder

WORKDIR /url-lookup/
# copying go.mod and go.sum
COPY go.* ./
RUN go mod download

# build the app
COPY main.go /url-lookup/
COPY /url/* /url-lookup/url/
RUN CGO_ENABLED=0 go build -o /bin/webserver

FROM scratch
COPY --from=builder /bin/webserver /bin/webserver
EXPOSE 8000
ENTRYPOINT ["/bin/webserver"]
