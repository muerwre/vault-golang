package constants

import "github.com/muerwre/vault-golang/utils"

const BorisNodeId = 696
const MaxCommentLength = 1024 * 2
const MaxNodeTitleLength = 256

const (
	NodeTypeImage string = "image"
	NodeTypeVideo string = "video"
	NodeTypeAudio string = "audio"
	NodeTypeText  string = "text"
	NodeTypeBoris string = "boris"
)

const (
	NodeFlowDisplaySingle     string = "single"
	NodeFlowDisplayVertical   string = "vertical"
	NodeFlowDisplayHorizontal string = "horizontal"
	NodeFlowDisplayQuadro     string = "quadro"
)

var FlowNodeTypes = &utils.EnumStringArray{NodeTypeImage, NodeTypeVideo, NodeTypeText, NodeTypeAudio}
var LabNodeTypes = &utils.EnumStringArray{NodeTypeText}
var NodeFlowDisplay = &utils.EnumStringArray{NodeFlowDisplaySingle, NodeFlowDisplayVertical, NodeFlowDisplayHorizontal, NodeFlowDisplayQuadro}
