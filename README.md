# Jellyfin Webhook Plugin Episodes Merger

[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/LICENSE-AGPLv3%20Liscense-blue?style=flat-square)](./LICENSE)
[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/GitHub-Jellyfin%20Webhook%20Plugin%20Episodes%20Merger-blueviolet?style=flat-square&logo=github)](https://github.com/fernvenue/jellyfin-plugin-webhook-episodes-merger)
[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/GitLab-Jellyfin%20Webhook%20Plugin%20Episodes%20Merger-orange?style=flat-square&logo=gitlab)](https://gitlab.com/fernvenue/jellyfin-plugin-webhook-episodes-merger)

Merge the webhooks of Episodes based on queue, work with [jellyfin/jellyfin-plugin-webhook](https://github.com/jellyfin/jellyfin-plugin-webhook).

## Usage

This tool is a middleware that listens to a TCP port and receives requests from the Webhook Plugin. It completes batch pushes by creating queues, avoiding the need to notify each Episode individually. This ensures that notifications for Episodes are maintained without sending a large number of notifications due to a rapid update of a batch of Episodes.

Here is an example of pushing to Telegram. Download directly from Releases or build it yourself, and then run:

```
./jellyfin-plugin-webhook-episodes-merger --target-url "https://api.telegram.org/bot******/sendMessage" --additional-params '{"chat_id": "******","parse_mode": "html"}'
```

In the Webhook configuration page of Jellyfin, select **Add Generic Destination**, then you only need to check **Item Added** and tick **Episodes** in **Item Type**, and the **Template** must be:

```
{
  "SeriesId": "{{SeriesId}}",
  "SeriesName": "{{SeriesName}}",
  "SeasonNumber": {{SeasonNumber}},
  "EpisodeNumber": {{EpisodeNumber}},
  "EpisodeName": "{{Name}}"
}
```

And Webhook URL should be the address and port that `jellyfin-plugin-webhook-episodes-merger` listen to, by deafult it will be `http://[::1]:8520`.

Then you will receive a notification, like this:

```
📺 Episode update reminder: Series Season 1

Episode 1 EpisodeTitle...
Episode 2 EpisodeTitle...
Episode 3 EpisodeTitle...
...
```

### Custom Push Format (Or you would like to use another language to notify.)

If you don't want to use the default notification format, or if you want to use a language other than English for notifications, here, taking Chinese users as an example, we can use parameters like this:

```
--text-content "📺 <b>單集更新提醒:</b> <b>{{.SeriesName}}</b> <b>第 {{.SeasonNumber}} 季</b>\n" --episode-format "\n第 {{.EpisodeNumber}} 集"
```

Then we can receive notifications like this:

```
📺 單集更新提醒: 某劇 第 n 季

第 1 集 這一集名
第 2 集 這一集名
```

You can also define additional content to be added to the outgoing requests, fully customizing the received requests and outgoing requests. For specific details, please refer to the help message or the additional instructions in the documentation.

### Push to Telegram with Series Poster

Here we need to go to Jellyfin first, change the **Template** in **Item Type** like this: 

```
{
  "SeriesId": "{{SeriesId}}",
  "SeriesName": "{{SeriesName}}",
  "SeasonNumber": {{SeasonNumber}},
  "EpisodeNumber": {{EpisodeNumber}},
  "EpisodeName": "{{Name}}"
}
```

And we should run with:

```
./jellyfin-plugin-webhook-episodes-merger --target-url "https://api.telegram.org/bot******/sendPhoto" --text-key "caption" --additional-params "{\"chat_id\": \"******\", \"photo\": \"https://***/Items/{{.SeriesId}}/Images/Primary\", \"parse_mode\": \"html\"}"
```

Other parameters remain unchanged, and then you will receive notifications with Series images.

### Push to Telegram with Series Poster and Redirect Button

As with the previous part, we only need to make a small modification:

```
./jellyfin-plugin-webhook-episodes-merger --target-url "https://api.telegram.org/bot******/sendPhoto" --text-key "caption" --additional-params "{\"reply_markup\": {\"inline_keyboard\": [[{\"text\": \"Go Check it Out!\", \"url\": \"https://******/web/#/details?id={{.SeriesId}}&serverId=******\"}]]}, \"chat_id\": \"******\", \"photo\": \"https://***/Items/{{.SeriesId}}/Images/Primary\", \"parse_mode\": \"html\"}"
```

Then you will receive a notification with a jump button!

## Configuration Options

| Parameter            | Description                                                                 | Default Value                                               |
|----------------------|-----------------------------------------------------------------------------|-------------------------------------------------------------|
| `--wait-second`       | The wait time in seconds before merging the notifications.                   | 300                                                         |
| `--text-content`      | The template for the notification text. You can use variables like `{{.SeriesName}}`, `{{.SeasonNumber}}`, and `{{.EpisodeName}}`. | `📺 <b>Episode update reminder:</b> <b>{{.SeriesName}}</b> <b>Season {{.SeasonNumber}}</b>\n` |
| `--text-key`          | The key used for the notification text in the JSON payload, allowing flexibility in JSON structure. | `text`                                                      |
| `--episode-format`    | The format for each episode's notification. You can use variables like `{{.EpisodeNumber}}` and `{{.EpisodeName}}`. | `\nEpisode {{.EpisodeNumber}}`                               |
| `--target-url`        | The target URL to send the notification to.                                  | `""` (Must be specified)                                    |
| `--additional-params` | Additional parameters in JSON format. Supports variables like `{{.SeriesId}}`. Example: `{"chat_id": "******", "photo": "https://example.com/{{.SeriesId}}"}`. | `{}` (Valid JSON format)                                    |
| `--listen-address`    | The address to listen on. Defaults to `::1`.                                | `::1`                                                       |
| `--listen-port`       | The port to listen on. Defaults to `8520`.                                  | 8520                                                        |
