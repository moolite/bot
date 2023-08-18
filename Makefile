SOURCES = $(shell find src/ -type f)

deps-lock.json: deps.edn
	CLJNIX_ADD_NIX_STORE=true nix run "github:jlesquembre/clj-nix#deps-lock"

bot.jar: deps.edn $(SOURCES)
	clj -X:build uber

bot.db: bot.sql
	sqlite3 -init bot.sql $@ '.exit'


clean:
	rm -f bot.db bot.jar
.PHONY: clean
