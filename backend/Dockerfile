FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mapup-api ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/mapup-api /mapup-api
EXPOSE 8080
ENTRYPOINT ["/mapup-api"]
