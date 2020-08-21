package main

// https://aaf.engineering/go-web-application-structure-part-2/
/*

	-- before release --
	TODO: update song titles after save
	TODO: setup deploy

	-- after release --
	TODO: test social login
	TODO: refactor GetDiff
	TODO: move all queries to repositories
	TODO: FRONT remove other user's profile completely, replace it with hover popup and make settings big fullscreen with message dialogs and etc

	DONE: search
	DONE: social login (vk, google)
	DONE: node view not updated? (actually, not getting last_seen_boris property for user)
	DONE: youtube titles
	DONE: /node POST
	DONE: /uploads
	DONE: /related
	DONE: add Responses and Requests
	DONE: /stats
	DONE: posting comments
	DONE: updating user profile
	DONE: updates: new messages not displayed in notifications because they should be Notification type, not message
*/

import (
	"github.com/muerwre/vault-golang/cmd"
)

func main() {
	cmd.Execute()
}
