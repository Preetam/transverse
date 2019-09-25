FROM golang:alpine AS build-go

COPY ./vendor /go/src/github.com/Preetam/transverse/vendor
COPY ./web /go/src/github.com/Preetam/transverse/web
COPY ./metadata /go/src/github.com/Preetam/transverse/metadata
COPY ./internal /go/src/github.com/Preetam/transverse/internal
RUN cd /go/src/github.com/Preetam/transverse/web && go build
RUN cd /go/src/github.com/Preetam/transverse/metadata && go build

FROM node AS build-ui

COPY ui /ui

RUN cd ui && npm i
RUN cd ui && npm test
RUN cd ui && make all && npm run build

FROM alpine

RUN mkdir -p /bin/transverse/web/static

COPY --from=build-go /go/src/github.com/Preetam/transverse/web/web /bin/transverse/web
COPY --from=build-go /go/src/github.com/Preetam/transverse/metadata/metadata /bin/transverse/metadata
COPY --from=build-go /go/src/github.com/Preetam/transverse/web/templates/ /bin/transverse/web/templates/
COPY --from=build-ui /ui/static/ /bin/transverse/web/static/

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT /entrypoint.sh
