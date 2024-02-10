all:
	make prepare
	make linux
	make macos
	make windows
	make pack

prepare:
	rm -f mattermost-syncplay-plugin.tar.gz
	rm -rf mattermost-syncplay-plugin
	mkdir -p mattermost-syncplay-plugin
	mkdir -p mattermost-syncplay-plugin/server

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mattermost-syncplay-plugin/server/plugin-linux-amd64 server/plugin.go

macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o mattermost-syncplay-plugin/server/plugin-darwin-amd64 server/plugin.go

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o mattermost-syncplay-plugin/server/plugin-windows-amd64 server/plugin.go

webapp:
	mkdir -p dist
	npm install
	./node_modules/.bin/webpack --mode=production

pack:
	cp plugin.json mattermost-syncplay-plugin/
	tar -czvf mattermost-syncplay-plugin.tar.gz mattermost-syncplay-plugin
