package main

// https://aaf.engineering/go-web-application-structure-part-2/

/*
	-- bugs --
	TODO: node recommendations displaying deleted nodes

	-- after release --
	TODO: try to fill metadata for images (if empty)
	TODO: add healthcheck endpoint
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
