build:
	docker build . -t ports4u
run:
	docker run --rm --cap-add=NET_ADMIN --cap-add=NET_RAW --name testme -it ports4u
all: build run
	