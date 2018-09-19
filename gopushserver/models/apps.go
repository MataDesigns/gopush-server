package models

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"../../storage"
	"../../storage/buntdb"
)

type Application struct {
	ID       string          `json:"id"`
	Name     string          `json:"name" binding:"required"`
	APNS     APNSApplication `json:"apns"`
	FCM      FCMApplication  `json:"fcm"`
	MaxRetry int             `json:"max_retry"`
}

type APNSApplication struct {
	KeyPath    string `json:"key_path"`
	Password   string `json:"password"`
	Production bool   `json:"production"`
	KeyID      string `json:"key_id"`
	TeamID     string `json:"team_id"`
}

type FCMApplication struct {
	APIKey string `json:"api_key"`
}

// GetApplicationStorage Get application storage buntdb
func GetApplicationStorage() storage.Storage {
	if gin.Mode() == gin.ReleaseMode {
		return buntdb.New("/db/apps.db")
	}
	return buntdb.New("./db/apps.db")
}

func GetApplication(key string, app *Application) error {
	var appJSON string
	appStore := GetApplicationStorage()
	if err := appStore.Get(key, &appJSON); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(appJSON), &app); err != nil {
		return err
	}
	return nil
}
