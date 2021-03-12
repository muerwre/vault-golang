package constants

import "github.com/muerwre/vault-golang/pkg"

const BorisNodeId = 696
const MaxCommentLength = 4096 * 2
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

var FlowNodeTypes = &pkg.EnumStringArray{NodeTypeImage, NodeTypeVideo,
	NodeTypeText, NodeTypeAudio}
var LabNodeTypes = &pkg.EnumStringArray{NodeTypeImage, NodeTypeVideo,
	NodeTypeText, NodeTypeAudio}
var NodeFlowDisplay = &pkg.EnumStringArray{NodeFlowDisplaySingle, NodeFlowDisplayVertical, NodeFlowDisplayHorizontal, NodeFlowDisplayQuadro}
