package main

// https://aaf.engineering/go-web-application-structure-part-2/
/*

	-- not working ---

	TODO: social login (vk, google)
	TODO: node view not updated?
	TODO: update song titles after save
	TODO: search

	TODO: refactor GetDiff
	TODO: move all queries to repositories

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
