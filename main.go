package main

// https://aaf.engineering/go-web-application-structure-part-2/
/*

	-- not working ---

	TODO: heroes displaying all nodes
	TODO: /uploads
	TODO: /stats
	TODO: /node POST
	TODO: youtube titles
	TODO: update song titles after save
	TODO: posting comments

	TODO: refactor GetDiff
	TODO: add Responses and Requests

	DONE: updating user profile
	DONE: updates: new messages not displayed in notifications because they should be Notification type, not message
*/

import (
	"github.com/muerwre/vault-golang/cmd"
)

func main() {
	cmd.Execute()
}
