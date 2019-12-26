package validation

import (
	"errors"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

var NODE_VALIDATORS = map[string]func(n *models.Node) error{
	models.FLOW_NODE_TYPES.IMAGE: ImageNodeValidator,
	models.FLOW_NODE_TYPES.AUDIO: AudioNodeValidator,
	models.FLOW_NODE_TYPES.TEXT:  TextNodeValidator,
	models.FLOW_NODE_TYPES.VIDEO: VideoNodeValidator,
}

// ImageNodeValidator validates node of type image
func ImageNodeValidator(n *models.Node) error {
	if n.FirstFileOfType(models.FILE_TYPES.IMAGE) < 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	return nil
}

// TextNodeValidator validates node of type text
func TextNodeValidator(n *models.Node) error {
	if len(n.Blocks) == 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	if n.FirstBlockOfType("text") < 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	return nil
}

// VideoNodeValidator validates node of type video
func VideoNodeValidator(n *models.Node) error {
	if len(n.Blocks) == 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	if n.FirstBlockOfType("video") < 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	return nil
}

// AudioNodeValidator validates node of type audio
func AudioNodeValidator(n *models.Node) error {
	if len(n.Blocks) == 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	if n.FirstFileOfType(models.FILE_TYPES.AUDIO) < 0 {
		return errors.New(codes.TOO_SHIRT)
	}

	return nil
}
