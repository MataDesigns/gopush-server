package models

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/matadesigns/gopushserver/storage"
	"github.com/matadesigns/gopushserver/storage/buntdb"
)

type RequestAddDevice struct {
	ApplicationID string `json:"app_id" binding:"required"`
	Device        Device `json:"device" binding:"required"`
}

type Device struct {
	ID       string `json:"id"`
	Platform int    `json:"platform" binding:"required"`
}

// GetDeviceStorage Get application devices storage buntdb
func GetDeviceStorage(appId string, platform string) storage.Storage {
	if gin.Mode() == gin.ReleaseMode {
		return buntdb.New(fmt.Sprintf("/db/%v-%v.db", appId, platform))
	}
	return buntdb.New(fmt.Sprintf("./db/%v-%v.db", appId, platform))
}
