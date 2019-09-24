FROM golang:alpine AS build-go

COPY ./vendor /go/src/github.com/Preetam/transverse/vendor
COPY ./web /go/src/github.com/Preetam/transverse/web
COPY ./metadata /go/src/github.com/Preetam/transverse/metadata
COPY ./internal /go/src/github.com/Preetam/transverse/internal
RUN cd /go/src/github.com/Preetam/transverse/web && go build
RUN cd /go/src/github.com/Preetam/transverse/metadata && go build

FROM node AS build-node

COPY web /web

RUN cd web && npm i
RUN cd web && npm test
RUN cd web && make all && npm run build

FROM alpine

RUN mkdir -p /bin/transverse/web/static

COPY --from=build-go /go/src/github.com/Preetam/transverse/web/web /bin/transverse/web
COPY --from=build-go /go/src/github.com/Preetam/transverse/metadata/metadata /bin/transverse/metadata
COPY --from=build-node /web/static/ /bin/transverse/web/static/
COPY --from=build-node /web/templates/ /bin/transverse/web/templates/

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT /entrypoint.sh
