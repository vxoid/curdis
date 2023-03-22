# Curdis v0.1.0
This is a library wrapper around discord API using http and websocket

## Update Logs

## Example
```go
bot, err := curdis.Auth(TOKEN)
if err != nil {
    panic(err)
}

clearCommand := curdis.Command{
    Name:        "clear",
    Description: "Deletes all messages in channel",
    Options: []curdis.Argument{
    {
        Name:        "channel",
        Description: "Channel to clear (if not specified clearing whole server)",
        Required:    false,
        Type:        curdis.ARG_CHANNEL,
    },
    },
}
err = bot.AddGlobalCommand(&clearCommand)
if err != nil {
    panic(err)
}
bot.AddCommandHandler(clearCommand, clearHandler)

defer func() {
    // This funcion clears all commands when bot is stopped
    commands, err := bot.GetGlobalCommands()
    if err != nil {
        panic(err)
    }

    for _, command := range commands {
        bot.DeleteGlobalCommand(&command)
    }
}()
go panic(bot.HandleEvents())
for {
}

func clearHandler(bot *curdis.Bot, user *curdis.User, guildId string, args []curdis.WSArgument) string {
	if len(args) > 0 {
		channelId := (*args[0].Value).(string)

		channel := curdis.Channel{
			Id: channelId,
		}
		err := clearChannel(bot, &channel)
		if err != nil {
			return "Error clearing channel 😡🌋: " + err.Error()
		}
		return "🫰Cleared channel✨"
	} else {
		var notcleared []curdis.Channel

		channels, err := bot.GetChannels(guildId)
		if err != nil {
			return "Got error: " + err.Error()
		}

		for _, channel := range channels {
			err := clearChannel(bot, &channel)
			if err != nil {
				notcleared = append(notcleared, channel)
			}
		}

		message := "🫰Cleared all channels "

		if len(notcleared) > 0 {
			message = fmt.Sprintf("%s instead of [ ", message)

			for _, channel := range notcleared {
				message = fmt.Sprintf("%s\"%s\" ", message, channel.Name)
			}

			message = fmt.Sprintf("%s]", message)
		}

		message = fmt.Sprintf("%s✨", message)

		return message
	}
}

func clearChannel(bot *curdis.Bot, channel *curdis.Channel) error {
	messages, err := bot.GetChannelMessages(channel)
	if err != nil {
		return err
	}

	for _, message := range messages {
		err = bot.DeleteMessage(&message)
		if err != nil {
			return err
		}
	}

	return nil
}
```