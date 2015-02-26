package agent

import (
	"github.com/telemetryapp/gotelemetry"
	"github.com/telemetryapp/gotelemetry_agent/agent/config"
)

func ProcessNotificationRequest(configFile *config.ConfigFile, errorChannel chan error, completionChannel chan bool, notificationChannel string, notification gotelemetry.Notification) {
	errorChannel <- gotelemetry.NewLogError("Notification mode is on.")

	apiKey, err := configFile.Accounts()[0].GetAPIKey()

	if err != nil {
		errorChannel <- err
		completionChannel <- true

		return
	}

	credentials, err := gotelemetry.NewCredentials(apiKey)

	if err != nil {
		errorChannel <- err
		completionChannel <- true

		return
	}

	credentials.SetDebugChannel(&errorChannel)

	channel := gotelemetry.NewChannel(notificationChannel)

	if err := channel.SendNotification(credentials, notification); err != nil {
		errorChannel <- err
	} else {
		errorChannel <- gotelemetry.NewLogError("Notification sent successfully.")
	}

	completionChannel <- true
}
