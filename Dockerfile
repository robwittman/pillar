FROM node:20-alpine AS ui
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /pillar ./cmd/pillar

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /pillar /usr/local/bin/pillar
EXPOSE 8080 9090
ENTRYPOINT ["pillar"]
