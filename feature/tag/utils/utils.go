package utils

import "github.com/muerwre/vault-golang/models"

func TagArrayContains(s []*models.Tag, el string) bool {
	for _, v := range s {
		if v.Title == el {
			return true
		}
	}

	return false
}
