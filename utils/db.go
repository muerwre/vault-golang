package utils

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/constants"
)

func WhereIsFlowNode(d *gorm.DB) *gorm.DB {
	return d.Where(
		"deleted_at IS NULL AND is_promoted = 1 AND is_public = 1 AND type IN (?)",
		structs.Values(constants.FLOW_NODE_TYPES),
	)
}
