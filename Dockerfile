FROM alpine:latest

ADD ./bin/internallb-webhook-admission-controller /internallb-webhook-admission-controller
CMD ["/internallb-webhook-admission-controller","--alsologtostderr","-v=4","2>&1"]
