dpipe
=====


         _       _           
        | |     (_)           
      __| |_ __  _ _ __   ___  
     / _  | '_ \| | '_ \ / _ \
    | (_| | |_) | | |_) |  __/ 
     \__,_| .__/|_| .__/ \___|
          | |     | |
          |_|     |_|
            

Distributed Data Pipeline

Performing "in-flight" processing of collected data, real time streaming analysis and alarming, and delivering the results to any number of destinations for further analysis.

[![Build Status](https://travis-ci.org/funkygao/dpipe.png?branch=master)](https://travis-ci.org/funkygao/dpipe)

### Install

    go get github.com/funkygao/dpipe

### Plugins

*   slide window based streaming biz alarm
*   cardinality statistics(for MAU alike counters where storing the data for statistics is prohibitive)
    
    In fact, if the data is stored only for the purpose of statistical calculations, incremental updates make storage unnecessary.
*   write events to ElasticSearch 
*   ElasticSearch buffering(lessen uneccessary load of ES, e,g. dau, pv, hits)
*   behaviour db, dimensional funnel analysis(user based action series)
*   batch processing of historical logs
*   self monitoring
*   to be more...

### Architecture

#### Overview

        +---------+     +---------+     +---------+     +---------+
        | server1 |     | server2 |     | server3 |     | serverN |
        |---------|     |---------|     |---------|     |---------|
        |syslog-ng|     |syslog-ng|     |syslog-ng|     |syslog-ng|
        |---------|     |---------|     |---------|     |---------|
        |collector|     |collector|     |collector|     |collector|
        +---------+     +---------+     +---------+     +---------+
            |               |               |               |
             -----------------------------------------------
                                    |
                                    | HTTP POST
                                    V
                            +-----------------+
                            |   ALS Server    |
                            |-----------------| 
                            |     dpiped      |
                            +-----------------+
                                        |
                                        |<----------------------------------------------------------------------+
                                        |                                                                       |
                                        | Input-Decode-Filter-Output                                            |
                                        V                                                                       |
                                   +----------------------------------------------------------------+           |
                                   |                   |           |           |           |        |           |
                              realtime analysis     indexer     archive    BehaviorDB      S3   hierarchy       |
                                   |                   |           |           |           |    deployment      |
            +----------------------|                   |           |           |           |        |           |
            |statistics            |alarming           |           |           |           |        |           |
       +----------+       +-----------------+          |           |           |           |        |           |
       |          |       |    |     |      |   ElasticSearch    HDFS      LevelDB/sky   RedShift  dpipe        |
     quantile   hyper     |    |   color    |          |           |           |           |        |           |
    histogram  loglog   beep email console etc         |           |           |           |        |           |
      topN        |       |    |     |      |          |           |      Dimensional      |        +-----------+
       |          |       +-----------------+       Kibana3        |    FunnelAnalysis   tableau
       +----------+                |                   |           |           |           |
            |                      |                   |           |           |           |
          PM/dev                dev/ops               PM          ops         dev         PM



### Implementation

#### Overview



                                             -- predict ----
                   (slide win)              |               |
    Input -> Filter(transform) -> Output -> |-- store ------| -> visualization
                   (cleaness)               |               | 
                   (decorator)              |-- explore ----|
                   (buffering)              |               |
                   (streaming)               -- alarm ------


#### PipelinePack

Main pipeline data structure containing a AlsMessage and other metadata

##### buffer size of PipelinePack

* EngineConfig
  - PoolSize
* Runner
  - PluginChanSize
* Router
  - PluginChanSize


##### data flow

                            -------<-------- 
                            |               |
                            V               | generate pool
           EngineConfig.inputRecycleChan    | recycling
                |           |               |
                | is         ------->------- 
                |
        InputRunner.inChan
                |
                |     +--------------------------------------------------------+
        consume |     |                     Router.inChan                      |
                V     +--------------------------------------------------------+
              Input         ^           |               |                   ^
                |           |           | put           | put               |
                |           |           |               |                   |
                 ----->-----       Matcher.inChan   Matcher.inChan          |
                   inject               |               |                   |
                                        | put           | put               |
                                        V               V                   |
                               OutputRunner.inChan   FilterRunner.inChan    |
                                        |               |                   |
                                        | consume       | consume           | inject
                                        V               V                   |
                                     Output           +------------------------+
                                                      |         Filter         |
                                                      +------------------------+
    
   
##### shutdown


        engine SIGINT
          |
        http.Stop
          |
        all Input.Stop()
          |
        router ----- close filterRunner.inChan --- Filter stopped
          |     |
          |      --- close outRunner.inChan ------ Output stopped
          |
        router ----- wait for FO runner finish --- close router.inChan 
          |
        done
