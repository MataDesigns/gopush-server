package gopushserver

import (
	"./models"
)

// InitWorkers for initialize all workers.
func InitWorkers(workerNum int64, queueNum int64) {
	LogAccess.Debug("worker number is ", workerNum, ", queue number is ", queueNum)
	QueueNotification = make(chan PushNotification, queueNum)
	for i := int64(0); i < workerNum; i++ {
		go startWorker()
	}
}

// SendNotification is send message to iOS or Android
func SendNotification(msg PushNotification) {
	switch msg.Platform {
	case PlatFormIos:
		PushToIOS(msg)
	case PlatFormAndroid:
		PushToAndroid(msg)
	}
}

func startWorker() {
	for {
		notification := <-QueueNotification
		SendNotification(notification)
	}
}

// queueNotification add notification to queue list.
func queueNotification(req RequestPush) (int, []LogPushEntry) {
	var app models.Application
	var count int
	log := make([]LogPushEntry, 0, count)
	if err := models.GetApplication(req.ApplicationID, &app); err != nil {
		return count, log
	}
	// wg := sync.WaitGroup{}
	newNotification := []*PushNotification{}
	for i := range req.Notifications {
		notification := &req.Notifications[i]
		notification.ApplicationID = app.ID
		newNotification = append(newNotification, notification)
	}

	for _, notification := range newNotification {
		// if PushConf.Core.Sync {
		// 	notification.wg = &wg
		// 	notification.log = &log
		// 	notification.AddWaitCount()
		// }
		QueueNotification <- *notification
		count += len(notification.Tokens)
		// Count topic message
		if notification.To != "" {
			count++
		}
	}

	// if PushConf.Core.Sync {
	// 	wg.Wait()
	// }

	// StatStorage.AddTotalCount(int64(count))

	return count, log
}
