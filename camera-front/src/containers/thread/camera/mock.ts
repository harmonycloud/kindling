export const testData: any = [
    {
        "name": "thread1",
        "events": [
            {
                "name": "event1",
                "start": 0,
                "time": 120
            }, {
                "name": "event2",
                "start": 120,
                "time": 120
            }, {
                "name": "event3",
                "start": 240,
                "time": 120
            }, {
                "name": "event4",
                "start": 360,
                "time": 120
            }, {
                "name": "event5",
                "start": 480,
                "time": 120
            }
        ],
        "logEvents": [
            {
                "time": 100,
                "log": 'this is log, oh no~~~~'
            }, {
                "time": 300,
                "log": 'this is log, oh no~~~~'
            }
        ]
    }, 
    {
        "name": "thread2",
        "events": [
            {
                "name": "event1",
                "start": 0,
                "time": 120
            }, {
                "name": "event2",
                "start": 120,
                "time": 120
            }, {
                "name": "event3",
                "start": 240,
                "time": 120
            }, {
                "name": "event4",
                "start": 360,
                "time": 120
            }, {
                "name": "event5",
                "start": 480,
                "time": 120
            }
        ]
    },
    {
        "name": "thread3",
        "events": [
            {
                "name": "event1",
                "start": 0,
                "time": 120
            }, {
                "name": "event2",
                "start": 120,
                "time": 0.1
            }, {
                "name": "event3",
                "start": 120.1,
                "time": 0.2
            }, {
                "name": "event4",
                "start": 120.3,
                "time": 40
            }, {
                "name": "event5",
                "start": 160.3,
                "time": 80.7
            }
        ]
    }, 
    {
        "name": "thread4",
        "events": [
            {
                "name": "event1",
                "start": 0,
                "time": 120
            }, {
                "name": "event2",
                "start": 120,
                "time": 0.05
            }, {
                "name": "event3",
                "start": 120.05,
                "time": 0.05
            }, {
                "name": "event4",
                "start": 120.1,
                "time": 40
            }, {
                "name": "event5",
                "start": 160.3,
                "time": 0.01
            }, {
                "name": "event6",
                "start": 160.31,
                "time": 80.7
            }
        ]
    }
]

// Only locks.
export const testData2: any = [{
    "pid": 25002,
    "tid": 25026,
    "threadName": "mysql-cj-abandoned-connection-cleanup",
    // Extracted from the log 
    "transactionId": "",
    // The start and end of this data segment. It is typically 1 second.
    "startTime": 1658136019874520600,
    "endTime": 1658136031663474700,
    // cpu on/off events list
    "cpuEvents": [
        {
            "startTime": 1658136015440450800,
            "endTime": 1658136020440660200,
            // Not used this version
            "runqLatency": "0,",
            // The latency of each piece of 'timeType'.
            "typeSpecs": "121160575,5000047316,",
            // 0: ON, 1: DISK, 2: NET, 3: FUTEX, 4: IDLE, 5: OHTER
            "timeType": "0,3,",
            // file/net format: type@operation@fdinfo@size(ts*)|
            // Use '|' as the separator
            // e.g. file@write@/root/a@8096|net@read@10.10.101.220:30002@size@ts|
            // futex format: futex@objAddress|*|
            "onInfo": "",
            "offInfo": "futex@addr140134989381716|",
            // format: len@logs|len2@logs2|0@||
            "log": ""
        },
        {
            "startTime": 1658136025440863700,
            "endTime": 1658136030441082000,
            "typeSpecs": "166664,5000049765,",
            "runqLatency": "0,",
            "timeType": "0,3,",
            "onInfo": "",
            "offInfo": "futex@addr140134989381716|",
            "log": ""
        }
    ],
    "javaFutexEvents": [
        {
            "startTime": 1658136015440608300,
            "endTime": 1658136020440688400,
            // kd@startTimestamp!endTimestamp!tid!lockAddress!LockEventType!threadName!waitForLockDuration!waitForTid!
            "dataValue": "kd@1658136015440608252!1658136020440688311!25026!cffad910!MonitorWait!mysql-cj-abandoned-connection-cleanup!5000080059!-1! "
        },
        {
            "startTime": 1658136025441031000,
            "endTime": 1658136030441104100,
            "dataValue": "kd@1658136025441030909!1658136030441104105!25026!cffad910!MonitorWait!mysql-cj-abandoned-connection-cleanup!5000073196!-1! "
        }
    ],
    // Not used.
    "IsSend": 0
}, {
    "pid": 25003,
    "tid": 26169,
    "threadName": "XNIO-1 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "522369878,1222785738,412178702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,3,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [
        {
            "startTime": 1658136016663432400,
            "endTime": 1658136031663474700,
            "dataValue": "kd@1658136016663432339!1658136031663474576!26169!d0390ae8!MonitorEnter!XNIO-1 task-10!15000042237!25090! "
        }
    ],
    "IsSend": 0
}, {
    "pid": 25004,
    "tid": 26170,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "1221278702,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,3,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25005,
    "tid": 26171,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25006,
    "tid": 26172,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25007,
    "tid": 26173,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25008,
    "tid": 26174,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 26004,
    "tid": 26175,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 27004,
    "tid": 26176,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 28004,
    "tid": 26277,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 29004,
    "tid": 26370,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25014,
    "tid": 26180,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}, {
    "pid": 25024,
    "tid": 26179,
    "threadName": "XNIO-2 task-10",
    "transactionId": "",
    "startTime": 1658136016662097400,
    "endTime": 1658136031663449600,
    "cpuEvents": [
        {
            "startTime": 1658136016662097400,
            "endTime": 1658136031663449600,
            "typeSpecs": "369878,1232785738,1221278702,5000016005,",
            "runqLatency": "0,0,",
            "timeType": "0,2,0,3,",
            "onInfo": "net@write@192.168.0.11:52130->192.168.0.9:9999@142||",
            "offInfo": "futex@addr140131632811972|futex@addr140131632811348|",
            "log": "96@17:20:16.662 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Finish SyncLock |98@17:20:16.663 [XNIO-1 task-10] INFO com.harmonycloud.stuck.web.WaitController - Start SyncLock... |"
        }
    ],
    "javaFutexEvents": [],
    "IsSend": 0
}]