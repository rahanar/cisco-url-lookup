FROM golang:1.15-alpine AS builder

WORKDIR /url-lookup/
# copying go.mod and go.sum
COPY go.* ./
RUN go mod download

# build the app
COPY main.go /url-lookup/
COPY url-database.json /url-lookup/url-database.json
COPY /url/* /url-lookup/url/
RUN CGO_ENABLED=0 go build 

FROM alpine
WORKDIR /url-lookup/
COPY --from=builder /url-lookup .
EXPOSE 8000
ENTRYPOINT ["./cisco-url-lookup"]
