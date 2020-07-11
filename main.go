package main

// https://aaf.engineering/go-web-application-structure-part-2/
/*

	-- not working ---

	TODO: /related
	TODO: /uploads
	TODO: /node POST
	TODO: youtube titles
	TODO: update song titles after save

	TODO: refactor GetDiff
	TODO: moved all queries to repositories

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
