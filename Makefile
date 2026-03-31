SOURCES = $(wildcard cmd/*/*.go) $(wildcard internal/*/*.go) $(wildcard pkg/*/*.go)

marrano-bot: ${SOURCES}
	go build \
		-tags "sqlite_foreign_keys" \
		-v ./cmd/marrano-bot

llm-chat: ${SOURCES}
	go build \
		-tags "fts5" \
		-v ./cmd/llm-chat

w: watch
watch:
	fd|entr make marrano-bot

test:
	go test -tags "fts5" -v ./...

tw: test-watch
test-watch:
	fd | entr make test

clean:
	rm marrano-bot llm-chat

all: marrano-bot llm-chat

.PHONY: all test clean w watch test-watch tw
