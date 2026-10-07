# Production image: the SPA embedded in a static Go binary. Migrations run separately:
#   docker compose run --rm app migrate up
FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
COPY app/openapi.json /src/app/openapi.json
RUN npm run build

FROM golang:1.26-alpine AS app
WORKDIR /src/app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
COPY --from=web /src/web/dist ./internal/platform/webui/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=app /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
CMD ["serve"]
