package examples

import cloudlog "github.com/anexia-it/go-cloudlog"

func main() {

	var logger *cloudlog.CloudLog
	var err error

	// init CloudLog client
	logger, err = cloudlog.InitCloudLog("index", "ca.pem", "cert.pem", "cert.key")
	if err != nil {
		panic(err)
	}

	// push event
	err = logger.PushEvent("event")
	if err != nil {
		panic(err)
	}

	// push multiple events
	err = logger.PushEvents([]string{"event1", "event2"})
	if err != nil {
		panic(err)
	}

	// close connection
	err = logger.Close()
	if err != nil {
		panic(err)
	}

}
