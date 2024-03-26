

build:
	docker build -f Dockerfile -t happyhour_backend .

run:
	docker run --env-file .env -p 80:80 happyhour_backend