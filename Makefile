
IMAGE_NAME=sample-1

run:
	go run github.com/undeconstructed/sample/sample test

grpc:
	protoc -I common/ common/store.proto --go_out=plugins=grpc:common

app:
	mkdir -p _build
	CGO_ENABLED=0 go build -o _build/app github.com/undeconstructed/sample/sample

clean:
	-rm -r _build

image: app
	bash make_image.sh $(IMAGE_NAME)
