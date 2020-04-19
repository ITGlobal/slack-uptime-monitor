package main

import (
	"fmt"
	"github.com/slack-go/slack"
	"log"
)

// Notifier sends notifications
type Notifier struct {
	slack *slack.Client
}

// NewNotifier creates new Notifier object
func NewNotifier() (*Notifier, error) {
	api := slack.New(SlackToken)
	profile, err := api.GetUserProfile("", false)
	if err != nil {
		log.Printf("SLACK: unable to connect: %v", err)
		return nil, err
	}

	log.Printf("SLACK: connected as \"%s\"", profile.DisplayName)
	return &Notifier{api}, nil
}

// Notify sends a "host is up/down" notification
func (n *Notifier) Notify(r *HealthcheckResult) error {
	if r.OK {
		return n.notifyHostUp(r)
	}

	return n.notifyHostDown(r)
}

func (n *Notifier) notifyHostDown(r *HealthcheckResult) error {
	title := fmt.Sprintf(":no_entry: \"%s\" is down", r.Healthcheck.Name)
	text := fmt.Sprintf("`%s`", r.Message)
	emoji := ":no_entry:"
	color := "#E72222"
	return n.notifyCore(r, title, text, emoji, color)
}

func (n *Notifier) notifyHostUp(r *HealthcheckResult) error {
	title := fmt.Sprintf(":ok: \"%s\" is up", r.Healthcheck.Name)
	text := ""
	emoji := ":ok:"
	color := "#22E722"
	return n.notifyCore(r, title, text, emoji, color)
}

func (n *Notifier) notifyCore(r *HealthcheckResult, title, text, emoji, color string) error {
	if r.Healthcheck.Notify.SlackMentions != nil && len(r.Healthcheck.Notify.SlackMentions) > 0 {
		if len(text) > 0 {
			text += "\n"
		}

		for i, m := range r.Healthcheck.Notify.SlackMentions {
			if i > 0 {
				text += ", "
			}

			text += fmt.Sprintf("<@%s>", m)
		}
	}

	options := make([]slack.MsgOption, 0)

	if len(text) > 0 {
		a := slack.Attachment{
			Text:     text,
			Title:    title,
			Fallback: title,
			Color:    color,
		}
		options = append(options, slack.MsgOptionAttachments(a))
	}

	if emoji != "" {
		options = append(options, slack.MsgOptionIconEmoji(emoji))
	}

	if SlackUsername != "" {
		options = append(options, slack.MsgOptionUsername(SlackUsername))
	}

	for _, to := range r.Healthcheck.Notify.SlackChannels {
		_, ts, _, err := n.slack.SendMessage(to, options...)
		if err != nil {
			log.Printf("SLACK: unable to send message to \"%s\": %v", to, err)
			return err
		}

		log.Printf("SLACK: message \"%s\" has been sent to \"%s\"", ts, to)
	}
	return nil
}
