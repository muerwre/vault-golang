package constants

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

type StringArray []string

var FlowNodeTypes = &StringArray{NodeTypeImage, NodeTypeVideo, NodeTypeText, NodeTypeAudio}
var LabNodeTypes = &StringArray{NodeTypeText}
var NodeFlowDisplay = &StringArray{NodeFlowDisplaySingle, NodeFlowDisplayVertical, NodeFlowDisplayHorizontal, NodeFlowDisplayQuadro}

func (f StringArray) Contains(t string) bool {
	for _, a := range f {
		if a == t {
			return true
		}
	}

	return false
}
