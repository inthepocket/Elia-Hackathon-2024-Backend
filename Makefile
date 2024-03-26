

build:
	docker build -f ./docker/Dockerfile -t happyhour_backend .

run:
	docker run happyhour_backend