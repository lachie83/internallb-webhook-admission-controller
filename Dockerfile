FROM golang:1.8

WORKDIR /go/src
RUN mkdir -p github.com/lachie83/internallb-webhook-admission-controller
COPY . ./github.com/lachie83/internallb-webhook-admission-controller
RUN go install github.com/lachie83/internallb-webhook-admission-controller
CMD ["internallb-webhook-admission-controller","--alsologtostderr","-v=4","2>&1"]
