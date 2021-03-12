package constants

import "github.com/fatih/structs"

const BorisNodeId = 696
const MaxCommentLength = 4096 * 2
const MaxNodeTitleLength = 256

type FlowNodeTypes struct {
	IMAGE string
	VIDEO string
	TEXT  string
	AUDIO string
}

type NodeTypes struct {
	IMAGE string
	VIDEO string
	TEXT  string
	BORIS string
	AUDIO string
}

type NodeFlowDisplay struct {
	SINGLE     string
	VERTICAL   string
	HORIZONTAL string
	QUADRO     string
}

var FLOW_NODE_TYPES = FlowNodeTypes{
	IMAGE: "image",
	VIDEO: "video",
	TEXT:  "text",
	AUDIO: "audio",
}

var NODE_TYPES = NodeTypes{
	IMAGE: "image",
	VIDEO: "video",
	AUDIO: "audio",
	TEXT:  "text",
	BORIS: "boris",
}

var NODE_FLOW_DISPLAY = NodeFlowDisplay{
	SINGLE:     "single",
	VERTICAL:   "vertical",
	HORIZONTAL: "horizontal",
	QUADRO:     "quadro",
}

func (f NodeFlowDisplay) Contains(t string) bool {
	for _, a := range structs.Map(f) {
		if a == t {
			return true
		}
	}

	return false
}

func (f FlowNodeTypes) Contains(t string) bool {
	for _, a := range structs.Map(f) {
		if a == t {
			return true
		}
	}

	return false
}
