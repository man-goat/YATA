FROM golang:alpine as builder

WORKDIR /bloingo
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOARCH=arm64 go build -o /yata

FROM scratch as final
WORKDIR /
COPY --from=builder /yata /yata

EXPOSE 80
ENTRYPOINT ["/yata"]