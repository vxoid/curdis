package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandler = func(*Bot, *Message)
type MessageDeleteHandler = func(*Bot, *Channel)
type CommandHandler = func(*Bot, *User, string, []WSArgument) string
type Bot struct {
	Token                string
	AppId                string
	client               http.Client
	messageHandler       MessageHandler
	messageDeleteHandler MessageDeleteHandler
	commandHandlers      map[string]CommandHandler
}
type Message struct {
	Content   string `json:"content"`
	ChannelId string `json:"channel_id"`
	Author    User   `json:"author"`
	Pinned    bool   `json:"pinned"`
	Id        string `json:"id"`
}
type sendMessage struct {
	Content string `json:"content"`
}
type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}
type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type uint8  `json:"type"`
}
type Command struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        uint8      `json:"type"`
	Id          string     `json:"id"`
	Options     []Argument `json:"options"`
}
type Argument struct {
	Type        uint8  `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}
type sendCommand struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        uint8      `json:"type"`
	Options     []Argument `json:"options"`
}
type Member struct {
	User User `json:"user"`
}
type WSResponse struct {
	Type uint8 `json:"type"`
	Data struct {
		Content string `json:"content"`
	} `json:"data"`
}
type WSInteraction struct {
	Token   string `json:"token"`
	Type    int    `json:"type"`
	Id      string `json:"id"`
	GuildId string `json:"guild_id"`
	Data    struct {
		Id      string       `json:"id"`
		Name    string       `json:"name"`
		Options []WSArgument `json:"options"`
	} `json:"data"`
	Member Member `json:"member"`
}
type WSArgument struct {
	Name  string       `json:"name"`
	Value *interface{} `json:"value,omitempty"`
}

func Auth(token string) (Bot, error) {
	client := http.Client{
		Timeout: 60 * time.Second,
	}
	request, err := http.NewRequest("GET", API+"/gateway/bot", nil)
	if err != nil {
		return Bot{}, err
	}

	request.Header.Set("Authorization", "Bot "+token)
	response, err := client.Do(request)
	if err != nil {
		return Bot{}, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return Bot{}, fmt.Errorf("authentication failed with status code %d", response.StatusCode)
	}

	bot := Bot{client: client, Token: token, commandHandlers: make(map[string]CommandHandler)}
	appId, err := bot.getAppId()
	if err != nil {
		return Bot{}, err
	}

	bot.AppId = appId
	return bot, err
}

func (bot *Bot) getAppId() (string, error) {
	request, err := http.NewRequest("GET", API+"/oauth2/applications/@me", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return "", fmt.Errorf("failed to get app id with status code %d", response.StatusCode)
	}

	var appInfo struct {
		Id string `json:"id"`
	}

	err = json.NewDecoder(response.Body).Decode(&appInfo)
	if err != nil {
		return "", err
	}

	return appInfo.Id, nil
}

func (bot *Bot) SendMessage(channel *Channel, content string) error {
	message := sendMessage{Content: content}

	payloadBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", API+"/channels/"+channel.Id+"/messages", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bot "+bot.Token)
	request.Header.Set("Content-Type", "application/json")

	response, err := bot.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to send message with status code %d", response.StatusCode)
	}

	return nil
}

func (bot *Bot) DeleteMessage(message *Message) error {
	request, err := http.NewRequest(http.MethodDelete, API+"/channels/"+message.ChannelId+"/messages/"+message.Id, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to delete message with status %s", response.Status)
	}

	return nil
}

func (bot *Bot) GetChannelMessages(channel *Channel) ([]Message, error) {
	request, err := http.NewRequest("GET", API+"/channels/"+channel.Id+"/messages", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get channel messages with status code %d", response.StatusCode)
	}

	var messages []Message
	err = json.NewDecoder(response.Body).Decode(&messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (bot *Bot) GetChannels(serverId string) ([]Channel, error) {
	request, err := http.NewRequest("GET", API+"/guilds/"+serverId+"/channels", nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get channels with status code %d", response.StatusCode)
	}

	var channels []Channel
	err = json.NewDecoder(response.Body).Decode(&channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (bot *Bot) AddGlobalCommand(command *Command) error {
	sendCommand := sendCommand{
		Name:        command.Name,
		Description: command.Description,
		Type:        CHAT_INPUT,
		Options:     command.Options,
	}

	payload, err := json.Marshal(&sendCommand)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", API+"/applications/"+bot.AppId+"/commands", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bot "+bot.Token)
	request.Header.Set("Content-Type", "application/json")

	response, err := bot.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to create global command with status code %s", response.Status)
	}

	return nil
}

func (bot *Bot) GetGlobalCommands() ([]Command, error) {
	request, err := http.NewRequest("GET", API+"/applications/"+bot.AppId+"/commands", nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get channel messages with status code %d", response.StatusCode)
	}

	var commands []Command
	err = json.NewDecoder(response.Body).Decode(&commands)
	if err != nil {
		return nil, err
	}

	return commands, nil
}

func (bot *Bot) DeleteGlobalCommand(command *Command) error {
	request, err := http.NewRequest(http.MethodDelete, API+"/applications/"+bot.AppId+"/commands/"+command.Id, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bot "+bot.Token)

	response, err := bot.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return fmt.Errorf("failed to delete message with status %s", response.Status)
	}

	return nil
}

func (bot *Bot) SetMessageDeleteHanler(handler MessageDeleteHandler) {
	bot.messageDeleteHandler = handler
}

func (bot *Bot) SetMessageHandler(handler MessageHandler) {
	bot.messageHandler = handler
}

func (bot *Bot) AddCommandHandler(command Command, handler CommandHandler) {
	bot.commandHandlers[command.Name] = handler
}

func (bot *Bot) HandleEvents() error {
	connection, _, err := websocket.DefaultDialer.Dial(WS_API, nil)
	if err != nil {
		return err
	}
	defer connection.Close()

	var idPayload struct {
		Code uint8 `json:"op"`
		Data struct {
			Token      string `json:"token"`
			Intents    uint   `json:"intents"`
			Properties struct {
				Os      string `json:"os"`
				Browser string `json:"browser"`
				Device  string `json:"device"`
			} `json:"properties"`
		} `json:"d"`
		Presence struct {
			Status string `json:"status"`
			Afk    bool   `json:"afk"`
		} `json:"presence"`
	}

	idPayload.Code = WS_INDENTIFY
	idPayload.Data.Token = bot.Token
	idPayload.Data.Intents = 513
	idPayload.Data.Properties.Os = "linux"
	idPayload.Data.Properties.Browser = "curdis"
	idPayload.Data.Properties.Device = "curdis"
	idPayload.Presence.Status = "online"
	idPayload.Presence.Afk = false

	payload, err := json.Marshal(idPayload)
	if err != nil {
		return err
	}

	err = connection.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		return err
	}

	for {
		_, messageBytes, err := connection.ReadMessage()
		if err != nil {
			return err
		}

		var event struct {
			Code uint8  `json:"op"`
			Type string `json:"t"`
		}
		err = json.Unmarshal(messageBytes, &event)
		if err != nil {
			fmt.Printf("Error parsing event data: %s\n", err)
			continue
		}

		switch event.Code {
		case 10:
			var event struct {
				Data struct {
					HeartbeatInterval uint64 `json:"heartbeat_interval"`
				} `json:"d"`
			}

			err = json.Unmarshal(messageBytes, &event)
			if err != nil {
				fmt.Printf("Error parsing event data: %s\n", err)
				continue
			}

			go func() {
				wait := time.Duration(event.Data.HeartbeatInterval) * time.Millisecond
				var payload struct {
					Code uint8  `json:"op"`
					Data uint64 `json:"d"`
				}
				payload.Code = 1

				for {
					payload.Data = uint64(time.Now().Unix())
					payloadBytes, err := json.Marshal(payload)
					if err != nil {
						return
					}

					time.Sleep(wait)

					err = connection.WriteMessage(websocket.TextMessage, payloadBytes)
					if err != nil {
						return
					}
				}
			}()

			continue
		}

		switch event.Type {
		case "INTERACTION_CREATE":
			var event_data struct {
				Interaction WSInteraction `json:"d"`
			}
			err = json.Unmarshal(messageBytes, &event_data)
			if err != nil {
				fmt.Printf("Error parsing event data: %s\n", err)
				continue
			}

			interaction := event_data.Interaction

			var WSresponse WSResponse
			WSresponse.Type = 4

			handler, found := bot.commandHandlers[interaction.Data.Name]
			if !found {
				WSresponse.Data.Content = "Command isn't handled yet"
			} else {
				WSresponse.Data.Content = handler(bot, &interaction.Member.User, interaction.GuildId, interaction.Data.Options)
			}

			payload, err = json.Marshal(WSresponse)
			if err != nil {
				return err
			}

			request, err := http.NewRequest(http.MethodPost, API+"/interactions/"+interaction.Id+"/"+interaction.Token+"/callback", bytes.NewBuffer(payload))
			if err != nil {
				return err
			}

			request.Header.Set("Content-Type", "application/json")

			response, err := bot.client.Do(request)
			if err != nil {
				return err
			}

			defer response.Body.Close()

			if response.StatusCode >= 300 {
				fmt.Printf("error sending interaction response with status code %s\n", response.Status)
			}
		case "MESSAGE_CREATE":
			if bot.messageHandler == nil {
				continue
			}
			var message Message

			var event_data struct {
				Message struct {
					Pinned    bool   `json:"bool"`
					Author    User   `json:"author"`
					Id        string `json:"id"`
					Content   string `json:"content"`
					ChannelId string `json:"channel_id"`
				} `json:"d"`
			}
			err = json.Unmarshal(messageBytes, &event_data)
			if err != nil {
				fmt.Printf("Error parsing event data: %s\n", err)
				continue
			}

			message.Pinned = event_data.Message.Pinned
			message.Author = event_data.Message.Author
			message.Id = event_data.Message.Id
			message.Content = event_data.Message.Content
			message.ChannelId = event_data.Message.ChannelId

			bot.messageHandler(bot, &message)
		case "MESSAGE_DELETE":
			if bot.messageDeleteHandler == nil {
				continue
			}

			var channel Channel

			var event_data struct {
				Data struct {
					ChannelId string `json:"channel_id"`
				} `json:"d"`
			}
			err = json.Unmarshal(messageBytes, &event_data)
			if err != nil {
				fmt.Printf("Error parsing event data: %s\n", err)
				continue
			}

			channel.Id = event_data.Data.ChannelId

			bot.messageDeleteHandler(bot, &channel)
		}
	}
}
