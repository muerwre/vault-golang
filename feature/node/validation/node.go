package validation

import (
	"errors"
	constants2 "github.com/muerwre/vault-golang/feature/node/constants"
	constants3 "github.com/muerwre/vault-golang/feature/upload/constants"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

var NodeValidators = map[string]func(n *models.Node) error{
	constants2.FLOW_NODE_TYPES.IMAGE: ImageNodeValidator,
	constants2.FLOW_NODE_TYPES.AUDIO: AudioNodeValidator,
	constants2.FLOW_NODE_TYPES.TEXT:  TextNodeValidator,
	constants2.FLOW_NODE_TYPES.VIDEO: VideoNodeValidator,
}

// ImageNodeValidator validates node of type image
func ImageNodeValidator(n *models.Node) error {
	if n.FirstFileOfType(constants3.FileTypeImage) < 0 {
		return errors.New(codes.TooShirt)
	}

	return nil
}

// TextNodeValidator validates node of type text
func TextNodeValidator(n *models.Node) error {
	if len(n.Blocks) == 0 {
		return errors.New(codes.TooShirt)
	}

	if n.FirstBlockOfType(models.BlockTypeText) < 0 {
		return errors.New(codes.TooShirt)
	}

	return nil
}

// VideoNodeValidator validates node of type video
func VideoNodeValidator(n *models.Node) error {
	if len(n.Blocks) == 0 {
		return errors.New(codes.TooShirt)
	}

	if n.FirstBlockOfType(models.BlockTypeVideo) < 0 {
		return errors.New(codes.TooShirt)
	}

	return nil
}

// AudioNodeValidator validates node of type audio
func AudioNodeValidator(n *models.Node) error {
	if n.FirstFileOfType(constants3.FileTypeAudio) < 0 {
		return errors.New(codes.TooShirt)
	}

	return nil
}
