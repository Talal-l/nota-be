FROM golang:1.24.2 as build-env

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-s -w" -a -installsuffix cgo -o nota
ARG APP_VERSION
RUN ./nota gen-docs

FROM scratch
WORKDIR /app
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /app/service/db/migrations ./service/db/migrations
COPY --from=build-env /app/nota .
COPY --from=build-env /app/docs ./docs
CMD [ "/app/nota"]
