package utils

import (
	"github.com/jinzhu/gorm"
	constants2 "github.com/muerwre/vault-golang/internal/feature/node/constants"
)

func WhereIsFlowNode(d *gorm.DB) *gorm.DB {
	return d.Where(
		"deleted_at IS NULL AND is_promoted = 1 AND is_public = 1 AND type IN (?)",
		constants2.FlowNodeTypes.AsCondition(),
	)
}

func WhereIsLabNode(d *gorm.DB) *gorm.DB {
	return d.Where(
		// "deleted_at IS NULL AND is_promoted = 0 AND is_public = 1 AND type IN (?)",
		"deleted_at IS NULL AND is_promoted = 1 AND is_public = 1 AND type IN"+
			" (?)",
		constants2.LabNodeTypes.AsCondition(),
	)
}
