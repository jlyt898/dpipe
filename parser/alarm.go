package parser

import (
	"fmt"
	mail "github.com/funkygao/alser/sendmail"
	"time"
)

// TODO
type Alarm interface {
	String() string
}

func runSendAlarmsWatchdog() {
	mailBody := ""
	bodyLines := 0

	for {
		select {
		case line, ok := <-chParserAlarm:
			if !ok {
				// chParserAlarm closed, this should never happen
				break
			}

			if debug {
				logger.Printf("got alarm: %s\n", line)
			}

			mailBody += line + "\n"
			bodyLines += 1

		case <-time.After(time.Second * 120):
			if mailBody != "" {
				go mail.Sendmail("peng.gao@funplusgame.com", fmt.Sprintf("%s %d", "ALS ", bodyLines), mailBody)
				logger.Printf("alarm sent=> %s\n", "peng.gao@funplusgame.com")

				mailBody = ""
				bodyLines = 0
			}

		}
	}
}
