FROM golang:1.27.1 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/defectdojo-mcp ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/defectdojo-mcp /defectdojo-mcp
USER nonroot
ENTRYPOINT ["/defectdojo-mcp"]
