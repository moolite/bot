# marrano-bot

A marrano bot

## Usage

    $ java -Dconfig="<path/to/config.edn>" -jar marrano-bot-0.1.0-SNAPSHOT-standalone.jar

## Options

``` edn
{:api "https://api.telegram.org/bot123456:12345678/"
 :hook {:url "https://bot.example.net/"
        :token "123456789"}}
```

- `:api`: telegram api URI, must be composed of *default api url* and *telegram token*
- `:hook :url`: base URI assigned to the bot
- `:hook :token`: token to concatenate to the bot's URI

## License

Copyright © 2018 Lorenzo Giuliani

Distributed under the Eclipse Public License either version 1.0 or (at
your option) any later version.
