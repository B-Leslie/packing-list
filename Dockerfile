# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 keeps the binary fully static (modernc sqlite is pure Go).
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/packing-list .
RUN go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed
RUN mkdir /out/data

FROM gcr.io/distroless/static:nonroot AS run
WORKDIR /app
COPY --from=build /out/packing-list /app/packing-list
COPY --from=build /out/seed /app/seed
COPY --from=build --chown=nonroot:nonroot /out/data /data
USER nonroot:nonroot
ENV DATA_DIR=/data
ENV PORT=8080
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/packing-list"]
