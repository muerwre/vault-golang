package main

/*

	App structure I use:
	https://aaf.engineering/go-web-application-structure-part-2/

	Goods:
	https://awesomeopensource.com/project/FlowerWrong/awesome-gin - list of good things
	https://github.com/olahol/melody - websocket middleware I'll use
*/

/*

	-- after release --
	TODO: add healthcheck endpoint
	TODO: try to fill metadata for images (if empty)
	TODO: websockets https://github.com/gin-gonic/gin/issues/461
	TODO: refactor GetDiff
	TODO: move all queries to repositories
	TODO: FRONT remove other user's profile completely, replace it with hover popup and make settings big fullscreen with message dialogs and etc

*/

import (
	"github.com/muerwre/vault-golang/cmd"
)

func main() {
	cmd.Execute()
}
