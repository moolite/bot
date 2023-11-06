SOURCES = $(wildcard cmd/*/*.go, internal/*/*.go)

marrano-bot: ${SOURCES}
	go build -tags "sqlite_foreign_keys"\
		-v ./cmd/marrano-bot

bot.db: bot.sql
	sqlite3 -init bot.sql $@ '.exit'

all:
test:
	go test ./internal/*/

clean:
	rm marrano-bot

.PHONY: all test clean
