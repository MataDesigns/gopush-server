package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/matadesigns/gopushserver/models"
	"github.com/matadesigns/gopushserver/storage"
)

func getApplicationsHandler(c *gin.Context) {
	appStore := models.GetApplicationStorage()
	appValues := []string{}
	appStore.GetAll(&appValues)
	apps := []models.Application{}
	for _, appValue := range appValues {
		var app models.Application
		if err := json.Unmarshal([]byte(appValue), &app); err != nil {
			continue
		}
		apps = append(apps, app)
	}

	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

func createApplicationHandler(c *gin.Context) {
	var app models.Application

	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app.ID = uuid.New().String()

	appStore := models.GetApplicationStorage()
	appJSON, _ := json.Marshal(app)
	appStore.Set(app.ID, string(appJSON))

	c.JSON(http.StatusCreated, gin.H{
		"application": app,
	})
}

func updateApplicationHandlder(c *gin.Context) {
	var app models.Application

	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appStore := models.GetApplicationStorage()
	appJSON, _ := json.Marshal(app)
	appStore.Set(app.ID, string(appJSON))

	c.JSON(http.StatusCreated, gin.H{
		"application": app,
	})
}

func createDeviceHandler(c *gin.Context) {
	var form models.RequestAddDevice

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appID := form.ApplicationID
	var app models.Application
	if err := models.GetApplication(appID, &app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := form.Device
	var deviceStore storage.Storage
	switch form.Device.Platform {
	case 1:
		deviceStore = models.GetDeviceStorage(app.ID, "ios")
	case 2:
		deviceStore = models.GetDeviceStorage(app.ID, "android")
	}
	deviceJSON, _ := json.Marshal(device)
	deviceStore.Set(device.ID, string(deviceJSON))

	c.JSON(http.StatusCreated, gin.H{
		"device": device,
	})
}

func getCountHandler(c *gin.Context) {
	appID := c.DefaultQuery("app_id", "")
	var app models.Application
	if err := models.GetApplication(appID, &app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceIosIds := []string{}
	deviceAndroidIds := []string{}
	models.GetDeviceStorage(appID, "ios").GetAllKeys(&deviceIosIds)
	models.GetDeviceStorage(appID, "android").GetAllKeys(&deviceAndroidIds)

	c.JSON(http.StatusOK, gin.H{
		"total_device_count":   len(deviceIosIds) + len(deviceAndroidIds),
		"ios_device_count":     len(deviceIosIds),
		"android_device_count": len(deviceAndroidIds),
	})

}

func pushHandler(c *gin.Context) {
	var form RequestPush

	if err := c.ShouldBindWith(&form, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(form.Notifications) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notifications field is empty."})
		return
	}

	// if int64(len(form.Notifications)) > PushConf.Core.MaxNotification {
	// 	msg = fmt.Sprintf("Number of notifications(%d) over limit(%d)", len(form.Notifications), PushConf.Core.MaxNotification)
	// 	LogAccess.Debug(msg)
	// 	abortWithError(c, http.StatusBadRequest, msg)
	// 	return
	// }

	for i := 0; i < len(form.Notifications); i++ {
		notification := form.Notifications[i]
		var store storage.Storage
		switch notification.Platform {
		case 1:
			store = models.GetDeviceStorage(form.ApplicationID, "ios")
		case 2:
			store = models.GetDeviceStorage(form.ApplicationID, "android")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown Platform"})
			return
		}
		deviceIds := []string{}
		store.GetAllKeys(&deviceIds)
		form.Notifications[i].Tokens = deviceIds
	}

	counts, logs := queueNotification(form)

	c.JSON(http.StatusOK, gin.H{
		"success": "ok",
		"counts":  counts,
		"logs":    logs,
	})
}

func heartbeatHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func GetRouterEngine() *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(LogMiddleware())
	r.Use(cors.Default())

	api := r.Group("/api")

	api.GET("/apps", getApplicationsHandler)
	api.POST("/apps", createApplicationHandler)
	api.PUT("/apps", updateApplicationHandlder)
	api.POST("/devices", createDeviceHandler)
	api.GET("/devices/count", getCountHandler)
	api.POST("/push", pushHandler)

	r.GET("/health", heartbeatHandler)

	return r
}
