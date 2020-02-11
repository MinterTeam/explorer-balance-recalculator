FROM golang:1.13.7-alpine as builder

WORKDIR /app
COPY ./ /app
RUN apk add --no-cache make gcc musl-dev linux-headers
RUN go mod download
RUN go build -o ./builds/linux/recalculator ./cmd/recalculator.go

FROM alpine:3.7

COPY --from=builder /app/builds/linux/recalculator /usr/bin/recalculator
RUN addgroup minteruser && adduser -D -h /minter -G minteruser minteruser
USER minteruser
WORKDIR /minter
ENTRYPOINT ["/usr/bin/recalculator"]
CMD ["recalculator"]