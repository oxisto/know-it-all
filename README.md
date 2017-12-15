# know-it-all

A SlackBot that aims to "know it all". It integrates with different data sources, such as Google Maps, Wikipedia or even your TeamSpeak server.

## Building

```
docker build -t know-it-all .
```
## Running

You have to at least provide an authentication token for your Slack workspace, i.e. by setting the SLACK_API_TOKEN environment variable.

```
docker run -e SLACK_API_TOKEN=XXXXXXX know-it-all
```
