SOURCES = $(wildcard cmd/*/*.go) $(wildcard internal/*/*.go)

marrano-bot: ${SOURCES}
	go build \
		-tags "sqlite_foreign_keys" \
		-v ./cmd/marrano-bot

w: watch
watch:
	fd|entr make marrano-bot

test:
	go test ./internal/*/

clean:
	rm marrano-bot

all: marrano-bot

.PHONY: all test clean w watch
