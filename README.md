# arboga-bot

Get stock information from Systembolagets API about a specific item found in a lsit of stores, send stock changes default every fifth minute to a Telegram channel.

Requires redis to run, on a Debian/Ubuntu system a simple "sudo apt install redis" should be enough.

The following environmental variables need to be set in order to use the program:

Telegram API key and channel:
```
AB_TGAPIKEY="key"
AB_TGCHANNEL="channel"

AB_SBAPIKEY="key"
```

## Running

```
go run arboga-bot.go
```

### DEBUG mode
```
go run arboga-bot.go -debug true
```