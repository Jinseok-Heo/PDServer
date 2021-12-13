MAIN=./cmd/server/main.go
NAME=plantdoctor

run: $(NAME)
build:
	source ./config/env-gen.sh && go build $(MAIN)

$(NAME):
	source ./config/env-gen.sh && go run $(MAIN)

docker:
	docker-compose up --build
