FROM golang:1.12 AS builder
ADD . /go/src/gitlab.com/asciishell/tfs-go-auction
WORKDIR /go/src/gitlab.com/asciishell/tfs-go-auction
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/tfs-go-auction ./cmd/auth-api/

FROM alpine:3.9
RUN apk add --update-cache ca-certificates
COPY --from=builder /go/src/gitlab.com/asciishell/tfs-go-auction/bin/tfs-go-auction /usr/local/bin/tfs-go-auction
COPY --from=builder /go/src/gitlab.com/asciishell/tfs-go-auction/internal/template/html/ /usr/local/etc/tfs-go-auction/internal/template/html/
WORKDIR /usr/local/etc/tfs-go-auction
ENTRYPOINT ["/usr/local/bin/tfs-go-auction"]
