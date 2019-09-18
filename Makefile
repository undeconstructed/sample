
IMAGE_NAME=sample-1

run:
	go run github.com/undeconstructed/sample/sample test

gen:
	go generate

app:
	mkdir -p _build
	CGO_ENABLED=0 go build -o _build/app github.com/undeconstructed/sample/sample

clean:
	-rm -r _build

image: app
	bash make_image.sh $(IMAGE_NAME)
