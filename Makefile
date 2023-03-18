SOURCES = $(shell find src/ -type f)

bot.jar: deps.edn $(SOURCES)
	clj -X:build uber

bot.db: bot.sql
	sqlite3 -init bot.sql $@ '.exit'


clean:
	rm -f bot.db bot.jar
.PHONY: clean
