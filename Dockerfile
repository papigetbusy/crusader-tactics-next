# The online server for Render (online step 3). It builds the game's pages and
# the server from this repository, then runs the server, which serves both.
FROM golang:1.24 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./adapters/server \
 && sh adapters/server/build-site.sh /out/site "online-$(date -u +%Y%m%d-%H%M)"

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/site /app/site
ENV SITE=/app/site
CMD ["/app/server"]
