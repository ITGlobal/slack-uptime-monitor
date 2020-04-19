package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config contains application configuration
type Config struct {
	Healthchecks []*Healthcheck
}

// Healthcheck defines a single healthcheck
type Healthcheck struct {
	ID     string
	URL    *url.URL
	Name   string
	Notify *NotificationConfig
}

// NotificationConfig defines targets for notifications
type NotificationConfig struct {
	SlackChannels []string
	SlackMentions []string
}

// LoadConfig reads config from a YAML file and parses it
func LoadConfig() (*Config, error) {
	yaml, err := LoadConfigYAML(ConfigPath)
	if err != nil {
		return nil, err
	}

	config, err := yaml.ParseConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ConfigYAML is a YAML model for config file (root)
type ConfigYAML struct {
	Healthchecks []*HealthcheckConfigYAML  `yaml:"healthchecks"`
	Notify       []*NotificationConfigYAML `yaml:"notify"`
	path         string                    `yaml:"-"`
}

// HealthcheckConfigYAML is a YAML model for config file
// ("healthchecks.*")
type HealthcheckConfigYAML struct {
	URL    string                    `yaml:"url"`
	Name   string                    `yaml:"name"`
	Notify []*NotificationConfigYAML `yaml:"notify"`
}

// NotificationConfigYAML is a YAML model for config file
// ("notify" and "healthchecks.*.notify")
type NotificationConfigYAML struct {
	SlackChannel  *string  `yaml:"slack"`
	SlackMentions []string `yaml:"slack_mention"`
}

// LoadConfigYAML reads config from a YAML file
func LoadConfigYAML(path string) (*ConfigYAML, error) {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("unable to load config file \"%s\": %v", path, err)
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewBuffer(buffer))
	config := &ConfigYAML{}
	err = decoder.Decode(config)
	if err != nil {
		log.Printf("unable to parse config file \"%s\": %v", path, err)
		return nil, err
	}

	config.path = path
	return config, nil
}

// ParseConfig creates a Config from its YAML representation
func (c *ConfigYAML) ParseConfig() (*Config, error) {
	config := &Config{
		Healthchecks: make([]*Healthcheck, len(c.Healthchecks)),
	}

	for i, h := range c.Healthchecks {
		hc, err := c.createHealthcheck(h, i)
		if err != nil {
			return nil, err
		}

		config.Healthchecks[i] = hc
	}

	return config, nil
}

func (c *ConfigYAML) createHealthcheck(h *HealthcheckConfigYAML, i int) (*Healthcheck, error) {
	hc := &Healthcheck{
		Notify: &NotificationConfig{
			SlackChannels: make([]string, 0),
			SlackMentions: make([]string, 0),
		},
	}

	// URL
	u, err := url.Parse(h.URL)
	if err != nil {
		log.Printf(
			"unable to parse config file \"%s\": field \"healthchecks[%d].url\" is malformed (%v)",
			c.path,
			i,
			err)
		return nil, err
	}
	hc.URL = u

	// Name
	hc.Name = h.Name
	if hc.Name == "" {
		hc.Name = u.Hostname()
	}

	// Notify
	if h.Notify != nil {
		for _, n := range h.Notify {
			if n.SlackChannel != nil {
				hc.Notify.SlackChannels = append(hc.Notify.SlackChannels, *n.SlackChannel)
			}

			if n.SlackMentions != nil {
				for _, m := range n.SlackMentions {
					hc.Notify.SlackMentions = append(hc.Notify.SlackMentions, m)
				}
			}
		}
	}

	if len(hc.Notify.SlackChannels) == 0 {
		if c.Notify != nil {
			for _, n := range c.Notify {
				if n.SlackChannel != nil {
					hc.Notify.SlackChannels = append(hc.Notify.SlackChannels, *n.SlackChannel)
				}
			}
		}

		if len(hc.Notify.SlackChannels) == 0 {
			log.Printf(
				"unable to parse config file \"%s\": no notifications are configured for \"healthchecks[%d]\"",
				c.path,
				i)
			return nil, err
		}
	}

	if len(hc.Notify.SlackMentions) == 0 {
		if c.Notify != nil {
			for _, n := range c.Notify {
				if n.SlackMentions != nil {
					for _, m := range n.SlackMentions {
						hc.Notify.SlackMentions = append(hc.Notify.SlackMentions, m)
					}
				}
			}
		}
	}

	// ID
	hasher := sha1.New()
	hasher.Write([]byte(strings.ToLower(hc.URL.String())))
	hc.ID = base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hc, nil
}
