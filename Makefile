APP       := news-svc
BUILD_DIR := bin
BINARY    := $(BUILD_DIR)/$(APP)
TAG       := latest
IMAGE     := $(APP):$(TAG)

all: build run

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) ./

run: build
	@$(BINARY)

test:
	go test ./...

fmt:
	go fmt ./...

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -d -p ${SERVER_PORT}:${SERVER_PORT} \
		-e MONGO_USER=$(MONGO_USER) \
		-e MONGO_PASSWORD=$(MONGO_PASSWORD) \
		-e MONGO_HOST=mongo \
		-e MONGO_PORT=27017 \
		-e MONGO_NAME=$(MONGO_NAME) \
		-e SERVER_PORT=$(SERVER_PORT) \
		-e SERVER_IS_DEV=$(SERVER_IS_DEV) \
		$(IMAGE)

compose-up:
	docker compose up --build

compose-down:
	docker compose down

clean:
	rm -rf $(BUILD_DIR)
