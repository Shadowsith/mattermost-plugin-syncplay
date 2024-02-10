## mattermost-plugin-syncplay
Mattermost server plugin to send video links to syncplay server via syncplay webhook

### Build
Run `make`

### Installation
Please read the offical Mattermost information to use custom plugins:
https://developers.mattermost.com/integrate/plugins/using-and-managing-plugins/#custom-plugins

### Usage
Type `/syncplay` in Mattermost chat after plugin has been enabled

#### Examples
- `/syncplay test_room https://test.com/test.mp4`
- `/syncplay https://test.com/test.mp4` only if default room is set

### License
MIT