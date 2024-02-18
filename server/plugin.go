package main

import (
	"fmt"
	"strconv"
	"strings"

	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type SyncplayPlugin struct {
	plugin.MattermostPlugin
}

type PluginSettings struct {
	Url           string `json:"url"`
	Port          int    `json:"port"`
	DefaultRoom   string `json:"default_room"`
	ChatResponse  bool   `json:"chat_response"`
	EnableBotUser bool   `json:"enable_bot_user"`
	BotUser       string `json:"bot_user"`
}

func (p *SyncplayPlugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cmdParams := strings.Split(args.Command, " ")

	if strings.TrimSpace(cmdParams[0]) == "/syncplay" {
		settings := p.getSettings()

		if len(cmdParams) < 3 && settings.DefaultRoom == "" {
			p.SendErrorMessage(args, nil, "Syncplay command need minimum two arguments!")
			return &model.CommandResponse{}, nil
		} else if len(cmdParams) < 2 {
			p.SendErrorMessage(args, nil, "Syncplay command need minimum the url argument!")
			return &model.CommandResponse{}, nil
		}

		if settings.Url == "" || settings.Port == 0 {
			p.SendErrorMessage(args, nil, "Syncplay plugin: url or port setting not set!")
			return &model.CommandResponse{}, nil
		}

		room := ""
		videoUrl := ""
		if settings.DefaultRoom != "" {
			room = settings.DefaultRoom
			videoUrl = cmdParams[1]
		} else {
			room = cmdParams[1]
			videoUrl = cmdParams[2]
		}

		formData := url.Values{}
		formData.Set("room", room)
		formData.Set("url", videoUrl)

		requestUrl := fmt.Sprintf("%s:%d", settings.Url, settings.Port)

		client := &http.Client{}
		req, err := http.NewRequest("POST", requestUrl, strings.NewReader(formData.Encode()))
		if err != nil {
			p.SendErrorMessage(args, err, "request generation error")
			return &model.CommandResponse{}, nil
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			p.SendErrorMessage(args, err, "request sending error")
			return &model.CommandResponse{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			p.SendErrorMessage(args, nil, "webhook http error, status code: "+strconv.Itoa(resp.StatusCode))
			return &model.CommandResponse{}, nil
		}

		userId := args.UserId
		username, err := p.getUsernameByUserId(args.UserId)
		if err != nil {
			username = "some user"
		}
		if settings.EnableBotUser {
			if settings.BotUser != "" {
				userId, err = p.getBotUserIdByName(settings.BotUser)
				if err != nil {
					userId = args.UserId
				}
			}
		}

		if settings.ChatResponse {

			post := &model.Post{
				UserId:    userId,
				ChannelId: args.ChannelId,
				Message:   fmt.Sprintf("Url %s successfully inserted in room %s by @%s", videoUrl, room, username),
			}

			if _, err := p.API.CreatePost(post); err != nil {
				return nil, model.NewAppError("ExecuteCommand", "Unable to post message", nil, err.Error(), http.StatusInternalServerError)
			}
		}

		return &model.CommandResponse{}, nil
	}

	return nil, model.NewAppError("ExecuteCommand", "Command not recognized "+args.Command, nil, "", http.StatusBadRequest)
}

func (p *SyncplayPlugin) SendErrorMessage(args *model.CommandArgs, err error, msg string) {
	error_msg := "unkown"
	if err != nil {
		error_msg = err.Error()
	}

	p.API.SendEphemeralPost(args.UserId, &model.Post{
		ChannelId: args.ChannelId,
		Message:   "Syncplay plugin error: " + msg + " error message " + error_msg,
	})
}

func (p *SyncplayPlugin) OnActivate() error {
	settings := p.getSettings()

	if settings.DefaultRoom != "" {
		if err := p.API.RegisterCommand(&model.Command{
			Trigger:          "syncplay",
			AutoComplete:     true,
			AutoCompleteDesc: "Send URL to Syncplay Room. Usage: /syncplay room_name video_url OR /syncplay video_url when default room is set",
			AutoCompleteHint: "E.g. /syncplay https://test.com/abc.mp4 OR /syncplay testroom https://test.com/abc.mp4",
		}); err != nil {
			return err
		}
	} else {
		if err := p.API.RegisterCommand(&model.Command{
			Trigger:          "syncplay",
			AutoComplete:     true,
			AutoCompleteDesc: "Send URL to Syncplay Room. Usage: /syncplay room_name video_url",
			AutoCompleteHint: "E.g. /syncplay test https://test.com/abc.mp4",
		}); err != nil {
			return err
		}
	}
	return nil
}

func (p *SyncplayPlugin) getBotUserIdByName(botName string) (string, error) {
	// Sucht nach Benutzern mit dem gegebenen Benutzernamen
	users, appErr := p.API.GetUsersByUsernames([]string{botName})
	if appErr != nil {
		return "", appErr
	}

	if len(users) == 0 {
		return "", nil
	}

	if !users[0].IsBot {
		return "", nil
	}

	return users[0].Id, nil
}

func (p *SyncplayPlugin) getUsernameByUserId(userId string) (string, error) {
	// Abrufen der Benutzerinformationen anhand der UserId
	user, appErr := p.API.GetUser(userId)
	if appErr != nil {
		return "", appErr
	}

	if user == nil {
		return "", nil
	}

	return user.Username, nil
}

func (p *SyncplayPlugin) getSettings() PluginSettings {
	pluginSettings, ok := p.API.GetConfig().PluginSettings.Plugins["syncplay"]
	if !ok {
		return p.getDefaultSettings()
	}

	settings := PluginSettings{
		Url:           p.getStrVal(pluginSettings["url"]),
		Port:          p.getIntVal(pluginSettings["port"]),
		DefaultRoom:   p.getStrVal(pluginSettings["default_room"]),
		ChatResponse:  p.getBoolVal(pluginSettings["chat_response"]),
		EnableBotUser: p.getBoolVal(pluginSettings["enable_bot_user"]),
		BotUser:       p.getStrVal(pluginSettings["bot_user"]),
	}

	return settings
}

func (p *SyncplayPlugin) getDefaultSettings() PluginSettings {
	return PluginSettings{
		Url:           "",
		Port:          0,
		DefaultRoom:   "",
		ChatResponse:  true,
		EnableBotUser: false,
		BotUser:       "",
	}
}

func (p *SyncplayPlugin) getStrVal(v interface{}) string {
	val, ok := v.(string)
	if !ok {
		val = ""
	}
	return val
}

func (p *SyncplayPlugin) getBoolVal(v interface{}) bool {
	val, ok := v.(bool)
	if !ok {
		val = true
	}
	return val
}

func (p *SyncplayPlugin) getIntVal(v interface{}) int {
	val, ok := v.(int)
	if !ok {
		val = 8080
	}
	return val
}

func main() {
	plugin.ClientMain(&SyncplayPlugin{})
}
