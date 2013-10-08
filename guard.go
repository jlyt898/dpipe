/*
           main
             |
             | goN
             |
     -----------------------
    |       |       |       |
   log1    log2    ...     logN

*/
package main

import (
    "github.com/funkygao/alser/parser"
    "github.com/funkygao/gofmt"
    "path/filepath"
    "runtime"
    "sync"
    "time"
)

func guard(jsonConfig jsonConfig) {
    parser.SetLogger(logger)
    parser.SetVerbose(options.verbose)
    parser.SetDebug(options.debug)

    var lines int = 0
    var workerN int = 0
    var wg = new(sync.WaitGroup)
    chLines := make(chan int)
    for _, item := range jsonConfig {
        paths, err := filepath.Glob(item.Pattern)
        if err != nil {
            panic(err)
        }

        for _, logfile := range paths {
            workerN++
            wg.Add(1)
            go run_worker(logfile, item, wg, chLines)
        }
    }

    if options.tick > 0 {
        ticker = time.NewTicker(time.Second * time.Duration(options.tick))
        go runTicker(&lines)
    }

    logger.Println(workerN, "workers started")

    // wait for all workers finish
    go func() {
        wg.Wait()
        logger.Println("all", workerN, " workers finished")
        close(chLines)
    }()

    // collect how many lines scanned
    for l := range chLines {
        lines += l
    }

    logger.Println("all lines scaned:", lines)
}

func runTicker(lines *int) {
    ms := new(runtime.MemStats)
    for _ = range ticker.C {
        runtime.ReadMemStats(ms)
        logger.Printf("goroutine: %d, mem: %s, lines: %d\n",
            runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc), *lines)
    }

}