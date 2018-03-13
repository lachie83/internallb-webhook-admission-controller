FROM golang:1.10-alpine as build

RUN apk --update add gcc libc-dev

ADD . /go/src/github.com/lachie83/internallb-webhook-admission-controller

WORKDIR /go/src/github.com/lachie83/internallb-webhook-admission-controller

RUN go build -buildmode=pie -ldflags "-linkmode external -extldflags -static -w" -o internallb-webhook-admission-controller

# The RUN line below is an alternative that also results in a static binary
# that can be run FROM scratch but results in a slightly-larger executable
# without ASLR.
#
# RUN CGO_ENABLED=0 go build -a -o internallb-webhook-admission-controller

FROM scratch

USER 1

EXPOSE 8443

COPY --from=build /go/src/github.com/lachie83/internallb-webhook-admission-controller/internallb-webhook-admission-controller /

CMD ["/internallb-webhook-admission-controller","--logtostderr","-v=4","2>&1"]
