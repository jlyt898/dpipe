package main

import (
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	mail "github.com/funkygao/alser/sendmail"
	"github.com/funkygao/gofmt"
	"path/filepath"
	"runtime"
	"time"
)

func runTicker(ticker *time.Ticker, lines *int) {
	startTime := time.Now()
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("ver:%s, goroutine:%d, mem:%s, workers:%d parsers:%d lines:%d, elapsed:%s\n",
			BuildID,
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			len(allWorkers), parser.ParsersCount(), *lines, time.Since(startTime))
	}
}

func runAlarmCollector(ch <-chan parser.Alarm) {
	// we don't know when to send alarm, we just send alarm one by one
	// alarm can span several lines
	// it's parsers' responsibility for flow control such as backoff
	for alarm := range ch {
		// TODO send email
		logger.Println(alarm)
	}
}

func notifyUnGuardedLogs(conf *config.Config) {
	const prefixLen = 5

	guardedLogs := make(map[string]bool)
	for _, g := range conf.Guards {
		var filePrefix string
		if options.tailmode {
			filePrefix = g.TailLogGlob
		} else {
			filePrefix = g.HistoryLogGlob
		}

		baseName := filepath.Base(filePrefix)
		guardedLogs[baseName[:prefixLen]] = true
	}

	// FIXME we assume that all the guarded logs are in the same dir
	var logfile string
	if options.tailmode {
		logfile = conf.Guards[0].TailLogGlob
	} else {
		logfile = conf.Guards[0].HistoryLogGlob
	}

	unGuardedLogs := make(map[string]bool)
	baseDir := filepath.Dir(logfile)
	allLogs, _ := filepath.Glob(baseDir + "/*")
	for _, path := range allLogs {
		baseName := filepath.Base(path)
		if _, present := guardedLogs[baseName[:prefixLen]]; !present {
			unGuardedLogs[path] = true
		}
	}

	if len(unGuardedLogs) > 0 {
		var mailBody = "Unguared logs:\n\n"
		for logfile, _ := range unGuardedLogs {
			mailBody += logfile + "\n"
		}

		mail.Sendmail(conf.String("unguarded.mail_to", ""), "Unguarded Logs", mailBody)
	}

}
