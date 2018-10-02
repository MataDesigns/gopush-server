package server

import (
	fcm "github.com/appleboy/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sirupsen/logrus"
)

var (
	Address string
	Port    string
	// QueueNotification is chan type
	QueueNotification chan PushNotification
	ApnsClients       map[string]*apns2.Client
	FcmClients        map[string]*fcm.Client
	// LogAccess is log server request log
	LogAccess *logrus.Logger
	// LogError is log server error log
	LogError *logrus.Logger
)
