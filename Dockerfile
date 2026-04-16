FROM node:20-alpine AS ui
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /src/web/dist web/dist
RUN CGO_ENABLED=0 go build -o /pillar ./cmd/pillar
RUN CGO_ENABLED=0 go build -o /pillarctl ./cmd/pillarctl
RUN CGO_ENABLED=0 go build -o /migrate ./scripts/migrate.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /pillar /usr/local/bin/pillar
COPY --from=build /pillarctl /usr/local/bin/pillarctl
COPY --from=build /migrate /usr/local/bin/migrate
COPY --from=build /src/internal/storage/postgres/migrations /migrations

ENV PILLAR_HTTP_ADDR=":8080"
ENV PILLAR_GRPC_ADDR=":9090"
EXPOSE 8080 9090

ENTRYPOINT ["pillar"]
