package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"

	fcm "github.com/appleboy/go-fcm"
)

// D provide string array
type D map[string]interface{}

const (
	// ApnsPriorityLow will tell APNs to send the push message at a time that takes
	// into account power considerations for the device. Notifications with this
	// priority might be grouped and delivered in bursts. They are throttled, and
	// in some cases are not delivered.
	ApnsPriorityLow = 5

	// ApnsPriorityHigh will tell APNs to send the push message immediately.
	// Notifications with this priority must trigger an alert, sound, or badge on
	// the target device. It is an error to use this priority for a push
	// notification that contains only the content-available key.
	ApnsPriorityHigh = 10
)

// Alert is APNs payload
type Alert struct {
	Action       string   `json:"action,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	Body         string   `json:"body,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	Title        string   `json:"title,omitempty"`
	Subtitle     string   `json:"subtitle,omitempty"`
	TitleLocArgs []string `json:"title-loc-args,omitempty"`
	TitleLocKey  string   `json:"title-loc-key,omitempty"`
}

// RequestPush support multiple notification request.
type RequestPush struct {
	Notifications []PushNotification `json:"notifications" binding:"required"`
	ApplicationID string             `json:"app_id" binding:"required"`
}

// PushNotification is single notification request
type PushNotification struct {
	ApplicationID string
	// Common
	Tokens           []string    `json:"tokens"`
	Platform         int         `json:"platform" binding:"required"`
	Message          string      `json:"message,omitempty"`
	Title            string      `json:"title,omitempty"`
	Priority         string      `json:"priority,omitempty"`
	ContentAvailable bool        `json:"content_available,omitempty"`
	MutableContent   bool        `json:"mutable_content,omitempty"`
	Sound            interface{} `json:"sound,omitempty"`
	Data             D           `json:"data,omitempty"`
	Retry            int         `json:"retry,omitempty"`
	wg               *sync.WaitGroup
	log              *[]LogPushEntry

	// Android
	APIKey                string           `json:"api_key,omitempty"`
	To                    string           `json:"to,omitempty"`
	CollapseKey           string           `json:"collapse_key,omitempty"`
	DelayWhileIdle        bool             `json:"delay_while_idle,omitempty"`
	TimeToLive            *uint            `json:"time_to_live,omitempty"`
	RestrictedPackageName string           `json:"restricted_package_name,omitempty"`
	DryRun                bool             `json:"dry_run,omitempty"`
	Condition             string           `json:"condition,omitempty"`
	Notification          fcm.Notification `json:"notification,omitempty"`

	// iOS
	Expiration  int64    `json:"expiration,omitempty"`
	ApnsID      string   `json:"apns_id,omitempty"`
	CollapseID  string   `json:"collapse_id,omitempty"`
	Topic       string   `json:"topic,omitempty"`
	Badge       *int     `json:"badge,omitempty"`
	Category    string   `json:"category,omitempty"`
	ThreadID    string   `json:"thread-id,omitempty"`
	URLArgs     []string `json:"url-args,omitempty"`
	Alert       Alert    `json:"alert,omitempty"`
	Production  bool     `json:"production,omitempty"`
	Development bool     `json:"development,omitempty"`
	SoundName   string   `json:"name,omitempty"`
	SoundVolume float32  `json:"volume,omitempty"`
}

// WaitDone decrements the WaitGroup counter.
func (p *PushNotification) WaitDone() {
	if p.wg != nil {
		p.wg.Done()
	}
}

// AddWaitCount increments the WaitGroup counter.
func (p *PushNotification) AddWaitCount() {
	if p.wg != nil {
		p.wg.Add(1)
	}
}

// AddLog record fail log of notification
func (p *PushNotification) AddLog(log LogPushEntry) {
	if p.log != nil {
		*p.log = append(*p.log, log)
	}
}

// IsTopic check if message format is topic for FCM
// ref: https://firebase.google.com/docs/cloud-messaging/send-message#topic-http-post-request
func (p *PushNotification) IsTopic() bool {
	return (p.Platform == PlatFormAndroid && p.To != "" && strings.HasPrefix(p.To, "/topics/")) ||
		p.Condition != ""
}

// CheckMessage for check request message
func CheckMessage(req PushNotification) error {
	var msg string

	// ignore send topic mesaage from FCM
	if !req.IsTopic() && len(req.Tokens) == 0 && len(req.To) == 0 {
		msg = "the message must specify at least one registration ID"
		LogAccess.Debug(msg)
		return errors.New(msg)
	}

	if len(req.Tokens) == PlatFormIos && len(req.Tokens[0]) == 0 {
		msg = "the token must not be empty"
		LogAccess.Debug(msg)
		return errors.New(msg)
	}

	if req.Platform == PlatFormAndroid && len(req.Tokens) > 1000 {
		msg = "the message may specify at most 1000 registration IDs"
		LogAccess.Debug(msg)
		return errors.New(msg)
	}

	// ref: https://firebase.google.com/docs/cloud-messaging/http-server-ref
	if req.Platform == PlatFormAndroid && req.TimeToLive != nil && (*req.TimeToLive < uint(0) || uint(2419200) < *req.TimeToLive) {
		msg = "the message's TimeToLive field must be an integer " +
			"between 0 and 2419200 (4 weeks)"
		LogAccess.Debug(msg)
		return errors.New(msg)
	}

	return nil
}

// SetProxy only working for FCM server.
func SetProxy(proxy string) error {

	proxyURL, err := url.ParseRequestURI(proxy)

	if err != nil {
		return err
	}

	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	LogAccess.Debug("Set http proxy as " + proxy)

	return nil
}
