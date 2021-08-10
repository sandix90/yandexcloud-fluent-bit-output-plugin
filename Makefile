build:
	mkdir -p "./bin"
	go build -buildmode=c-shared -o ./bin/out_yandexcloud.so yandex_cloud_logging.go

all:
	go test -timeout=120s -v -cover ./...
	mkdir -p "./bin"
	go build -buildmode=c-shared -o ./bin/out_yandexcloud.so yandex_cloud_logging.go

.PHONY: test
test:
	go test -timeout=120s -v -cover ./...

clean:
	rm -rf *.so *.h *~