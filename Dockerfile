FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gnosis-workspace .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /out/gnosis-workspace /gnosis-workspace
COPY --from=build /src/migrations /migrations

EXPOSE 4000 50051

USER nonroot:nonroot

ENTRYPOINT ["/gnosis-workspace"]