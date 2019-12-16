package models

type Message struct {
	*CommentLike

	From *User `json:"from"`
	To   *User `json:"to"`
}
