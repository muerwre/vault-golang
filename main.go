package main

// https://aaf.engineering/go-web-application-structure-part-2/

/*

	-- before release --
	TODO: empty tags are still posted
	TODO: inputing "Tag" adds "tag" duplicate to form
	TODO: add healthcheck endpoint
	TODO: refactor GetDiff

	-- after release --
	TODO: move all queries to repositories
	TODO: FRONT remove other user's profile completely, replace it with hover popup and make settings big fullscreen with message dialogs and etc

*/

import (
	"github.com/muerwre/vault-golang/cmd"
)

func main() {
	cmd.Execute()
}
