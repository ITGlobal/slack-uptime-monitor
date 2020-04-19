slack-uptime-monitor
====================

![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/itglobal/slack-uptime-monitor?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/itglobal/slack-uptime-monitor?style=flat-square)
![GitHub](https://img.shields.io/github/license/itglobal/slack-uptime-monitor?style=flat-square)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/itglobal/slack-uptime-monitor?style=flat-square)

Small app that monitors specified HTTP endpoints and reports anything to Slack.

Install
-------

Create `docker-compose.yaml` file:

```yaml
version: "2.4"
services:
  slack_uptime_monitor:
    image: itglobal/slack-uptime-monitor
    container_name: slack_uptime_monitor
    # Map .env file that contains sensistive settings
    env_file: .env
    # Map data directory ./var
    volumes:
      - ./var:/app/var
    # Expose 5000 port for self-monitoring (see below)
    # ports:
    #   - 5000:5000
```

Configure
---------

Create file `.env`:

```shell
SLACK_TOKEN="your-slack-access-token"
SLACK_USERNAME="your-slack-bot-name" # optional, default is "UptimeMonitor"
```

Define healthchecks
-------------------

Create file `./var/config.yaml`:

```yaml
# Endpoints to check
healthchecks:
    # Simple definition - will use auto-generated name and default notification settings
    - url: https://hostname.com/

    # Definition with custom name and default notification settings
    - url: https://hostname.com/
      name: "hostname.com_2"

    # Definition with custom name and custom notification settings
    - url: https://hostname.com/
      name: "hostname.com_3"
      notify:
          # Will send notifications to default channel and will mention @username
          - slack_mention:
                - "username"

    # Definition with custom name and more custom notification settings
    - url: https://hostname.com/
      name: "hostname.com_4"
      notify:
          # Will send notifications to "@username
          - slack: "@username"

# Default notification settings
# Will send notifications to #channel and will mention @username
notify:
    - slack: "#channel"
      slack_mention:
          - "username"
```

Start
-----

```shell
docker-compose up -d
```

Self-monitoring
---------------

**SlackUptimeMonitor** contains a built-in healthcheck.
It's available at `GET http://localhost:5000`:

* `200 OK` means heathchecks are being performed
* anything else means that something has gone wrong

License
-------

[MIT](LICENSE)