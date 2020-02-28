# build the server binary
FROM golang:1.10.0 AS builder
LABEL stage=server-intermediate
WORKDIR /go/src/github.com/infobloxopen/atlas-contacts-app
COPY . .

RUN go build -o bin/server ./cmd/server

# copy the server binary from builder stage; run the server binary
FROM alpine:latest AS runner

# Go programs require libc
RUN mkdir /lib64 && \
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /bin
COPY --from=builder /go/src/github.com/infobloxopen/atlas-contacts-app/bin/server .
COPY pkg/pb/contacts.swagger.json /go/src/github.com/infobloxopen/atlas-contacts-app/www/swagger.json
ENTRYPOINT ["server", "-swagger-dir", "/go/src/github.com/infobloxopen/atlas-contacts-app/www/swagger.json"]
