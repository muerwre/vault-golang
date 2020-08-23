FROM golang:1.15 AS build-env

RUN apt-get update && apt-get install -y libjpeg-dev -q

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

WORKDIR /usr/src/app/bin
COPY ./docker/vault/wait-for-it.sh .
COPY ./config.yaml .
COPY ./templates/. ./templates/.

EXPOSE $EXPOSE
# HEALTHCHECK --interval=1m --timeout=30s CMD curl -f http://app:$EXPOSE/api/route/list/all || kill -s 2 1
CMD ["./vault", "serve"]