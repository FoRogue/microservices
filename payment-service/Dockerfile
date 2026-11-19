FROM golang:1.27-alpine AS build
WORKDIR /src
COPY --from=eventing /go.mod /go.sum /eventing/
COPY go.mod go.sum ./
RUN go mod download
COPY --from=eventing / /eventing/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/payment-service .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/payment-service .
EXPOSE 8083
ENTRYPOINT ["./payment-service"]
