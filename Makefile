build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/internallb-webhook-admission-controller .
	docker build -t internallb-webhook-admission-controller .

