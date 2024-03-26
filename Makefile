

build:
	docker build -f ./docker/Dockerfile -t happyhour_backend .

run:
	docker run -p 80:80 happyhour_backend