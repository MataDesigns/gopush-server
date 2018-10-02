package server

import (
	"errors"
	"fmt"

	"github.com/appleboy/go-fcm"
	"github.com/matadesigns/gopushserver/models"
)

// InitFCMClient use for initialize FCM Client.
func InitFCMClient(key string, app models.Application) (*fcm.Client, error) {
	var err error
	fcmConfig := app.FCM
	if key == "" {
		return nil, errors.New("Missing Android API Key")
	}

	if key != fcmConfig.APIKey {
		return fcm.NewClient(key)
	}
	client, ok := FcmClients[app.ID]
	if FcmClients == nil {
		FcmClients = make(map[string]*fcm.Client)
	}
	if !ok {
		client, err = fcm.NewClient(key)
		FcmClients[app.ID] = client
		return FcmClients[app.ID], err
	}

	return client, nil
}

// GetAndroidNotification use for define Android notification.
// HTTP Connection Server Reference for Android
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func GetAndroidNotification(req PushNotification) *fcm.Message {
	notification := &fcm.Message{
		To:                    req.To,
		Condition:             req.Condition,
		CollapseKey:           req.CollapseKey,
		ContentAvailable:      req.ContentAvailable,
		MutableContent:        req.MutableContent,
		DelayWhileIdle:        req.DelayWhileIdle,
		TimeToLive:            req.TimeToLive,
		RestrictedPackageName: req.RestrictedPackageName,
		DryRun:                req.DryRun,
	}

	if len(req.Tokens) > 0 {
		notification.RegistrationIDs = req.Tokens
	}

	if len(req.Priority) > 0 && req.Priority == "high" {
		notification.Priority = "high"
	}

	// Add another field
	if len(req.Data) > 0 {
		notification.Data = make(map[string]interface{})
		for k, v := range req.Data {
			notification.Data[k] = v
		}
	}

	notification.Notification = &req.Notification

	// Set request message if body is empty
	if len(req.Message) > 0 {
		notification.Notification.Body = req.Message
	}

	if len(req.Title) > 0 {
		notification.Notification.Title = req.Title
	}

	if v, ok := req.Sound.(string); ok && len(v) > 0 {
		notification.Notification.Sound = v
	}

	return notification
}

// PushToAndroid provide send notification to Android server.
func PushToAndroid(req PushNotification) bool {
	LogAccess.Debug("Start push notification for Android")
	// if PushConf.Core.Sync {
	// 	defer req.WaitDone()
	// }
	var app models.Application
	if err := models.GetApplication(req.ApplicationID, &app); err != nil {
		return false
	}
	fmt.Println("PUSHING TO ANDROID")
	var (
		client     *fcm.Client
		retryCount = 0
		maxRetry   = app.MaxRetry
	)

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}

	// check message
	err := CheckMessage(req)

	if err != nil {
		LogError.Error("request error: " + err.Error())
		return false
	}

Retry:
	var isError = false

	notification := GetAndroidNotification(req)

	if req.APIKey != "" {
		client, err = InitFCMClient(req.APIKey, app)
	} else {
		client, err = InitFCMClient(app.FCM.APIKey, app)
	}

	if err != nil {
		// FCM server error
		LogError.Error("FCM server error: " + err.Error())
		return false
	}

	res, err := client.Send(notification)
	if err != nil {
		// Send Message error
		LogError.Error("FCM server send message error: " + err.Error())
		return false
	}

	if !req.IsTopic() {
		LogAccess.Debug(fmt.Sprintf("Android Success count: %d, Failure count: %d", res.Success, res.Failure))
	}

	// StatStorage.AddAndroidSuccess(int64(res.Success))
	// StatStorage.AddAndroidError(int64(res.Failure))

	var newTokens []string
	// result from Send messages to specific devices
	for k, result := range res.Results {
		to := ""
		if k < len(req.Tokens) {
			to = req.Tokens[k]
		} else {
			to = req.To
		}

		if result.Error != nil {
			isError = true
			newTokens = append(newTokens, to)
			LogPush(FailedPush, to, req, result.Error)
			// if PushConf.Core.Sync {
			// 	req.AddLog(getLogPushEntry(FailedPush, to, req, result.Error))
			// }
			continue
		}

		LogPush(SucceededPush, to, req, nil)
	}

	// result from Send messages to topics
	if req.IsTopic() {
		to := ""
		if req.To != "" {
			to = req.To
		} else {
			to = req.Condition
		}
		LogAccess.Debug("Send Topic Message: ", to)
		// Success
		if res.MessageID != 0 {
			LogPush(SucceededPush, to, req, nil)
		} else {
			isError = true
			// failure
			LogPush(FailedPush, to, req, res.Error)
			// if PushConf.Core.Sync {
			// 	req.AddLog(getLogPushEntry(FailedPush, to, req, res.Error))
			// }
		}
	}

	// Device Group HTTP Response
	if len(res.FailedRegistrationIDs) > 0 {
		isError = true
		for _, id := range res.FailedRegistrationIDs {
			newTokens = append(newTokens, id)
		}

		LogPush(FailedPush, notification.To, req, errors.New("device group: partial success or all fails"))
		// if PushConf.Core.Sync {
		// 	req.AddLog(getLogPushEntry(FailedPush, notification.To, req, errors.New("device group: partial success or all fails")))
		// }
	}

	if isError && retryCount < maxRetry {
		retryCount++

		// resend fail token
		req.Tokens = newTokens
		goto Retry
	}

	return isError
}
