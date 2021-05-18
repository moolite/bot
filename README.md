# marrano-bot

A marrano bot

## Usage

    $ java -Dconfig="<path/to/config.edn>" -jar marrano-bot-0.1.0-SNAPSHOT-standalone.jar

## Options

``` edn
{:webhook "https://bot.example.net"
 :token "botABC:123467"}
```

- `:webhook` Bot's *base* webhook URL
- `:secret` bot's secret (used by the webhook)
- `:token` telegram's bot secret token

The bot will register the webhook `<:webhook>/t/<:secret>`

## License

Copyright © 2018 Lorenzo Giuliani

Release under MPL-2.0, see attached [LICENSE](LICENSE) file.
