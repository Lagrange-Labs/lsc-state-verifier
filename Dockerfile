FROM golang:1.21-alpine AS build

RUN apk add --no-cache --update gcc g++ make

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:edge
COPY config.toml /app/config.toml
COPY --from=build /src/dist/lsc-state-verifier /app/lsc-state-verifier
CMD ["/bin/sh", "-c", "/app/lsc-state-verifier db"]