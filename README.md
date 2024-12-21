# Jellyfin Webhook Plugin Episodes Merger

[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/LICENSE-AGPLv3%20Liscense-blue?style=flat-square)](./LICENSE)
[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/GitHub-Jellyfin%20Webhook%20Plugin%20Episodes%20Merger-blueviolet?style=flat-square&logo=github)](https://github.com/fernvenue/jellyfin-plugin-webhook-episodes-merger)
[![jellyfin-plugin-webhook-episodes-merger](https://img.shields.io/badge/GitLab-Jellyfin%20Webhook%20Plugin%20Episodes%20Merger-orange?style=flat-square&logo=gitlab)](https://gitlab.com/fernvenue/jellyfin-plugin-webhook-episodes-merger)

Merge the webhooks of Episodes based on queue.

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
  "EpisodeNumber": {{EpisodeNumber}}
}
```

And Webhook URL should be the address and port that `jellyfin-plugin-webhook-episodes-merger` listen to, by deafult it will be `http://[::1]:8520`.

Then you will receive a notification, like this:

```
📺 Episode update reminder: Series Season 1

Episode 1
Episode 2
Episode 3
...
```

### Custom Push Format

Here, taking Chinese users as an example, we can use parameters like this:

```
--text-content "📺 <b>单集更新提醒:</b> <b>{{.SeriesName}}</b> <b>第 {{.SeasonNumber}} 季</b>\n" --episode-format "\n第 {{.EpisodeNumber}} 集"
```

Then we can receive notifications like this:

```
📺 单集更新提醒: 某剧 第 n 季

第 1 集
第 2 集
```

You can also define additional content to be added to the outgoing requests, fully customizing the received requests and outgoing requests. For specific details, please refer to the help message or the additional instructions in the documentation.

## Configuration Options

| Parameter            | Description                                                                 | Default Value                                               |
|----------------------|-----------------------------------------------------------------------------|-------------------------------------------------------------|
| `--wait-second`       | The wait time in seconds before merging the notifications.                   | 300                                                         |
| `--text-content`      | The template for the notification text. You can use variables like `{{.SeriesName}}`. | `📺 <b>Episode update reminder:</b> <b>{{.SeriesName}}</b> <b>Season {{.SeasonNumber}}</b>\n` |
| `--episode-format`    | The format for each episode's notification. You can use variables like `{{.EpisodeNumber}}`. | `\nEpisode {{.EpisodeNumber}}`                               |
| `--target-url`        | The target URL to send the notification to.                                  | `""` (Must be specified)                                    |
| `--additional-params` | Additional parameters in JSON format, such as `chat_id` for Telegram.        | `{}` (Valid JSON format)                                    |
| `--content-header`    | The key for the notification text in the JSON payload.                      | `text`                                                      |
| `--listen-address`    | The address to listen on. Defaults to `::1`.                                | `::1`                                                       |
| `--listen-port`       | The port to listen on. Defaults to `8520`.    
