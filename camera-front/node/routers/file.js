
'use strict';
const router = require('express').Router();
const _ = require('lodash');
const fs = require('fs');
const sanitize = require("sanitize-filename");

const settings = require('../settings');
const path = require('path');
const basicFloder = settings.traceFilePath;

router.get("/getFolders", function(req, res, next) {
    let list = [], result = [];
    fs.readdir(basicFloder, function(err, files) {
        if (err) {
            console.error("Error: %s", err);
            res.status(500).json({
                success: false,
                message: err
            })
        }
        _.forEach(files, file => {
            let fileStats = fs.statSync(path.join(basicFloder, file));
            if (fileStats.isDirectory()) {
                let fileNameList = file.split('_');
                if (fileNameList.length === 4) {
                    list.push(fileNameList);
                }
            }
        });
        let workloadlist = _.groupBy(list, item => item[0]);
        _.forEach(workloadlist, (list, workload) => {
            let workloadObj = {
                title: workload,
                value: workload,
                selectable: false,
                children: []
            };
            let podList = _.groupBy(list, item => item[1]);
            _.forEach(podList, (list2, pod) => {
                let podObj = {
                    title: pod,
                    value: `${workload}_${pod}`,
                    selectable: false,
                    children: []
                };
                _.forEach(list2, item => {
                    podObj.children.push({
                        title: `${item[2]}(${item[3]})`,
                        value: _.join(item, '_')
                    });
                });
                workloadObj.children.push(podObj);
            });
            result.push(workloadObj);
        });
        // console.log(result);
        res.status(200).json({
            success: true,
            data: result
        });
    });
}); 

router.get("/getAllTraceFileList", function(req, res, next) {
    let folderName = sanitize(req.query.folderName);
    const filePath = path.join(basicFloder, folderName);
    let list = [];
	try {
        let files = fs.readdirSync(filePath);
        _.forEach(files, function(file) {
            let fileStats = fs.statSync(path.join(filePath, file));
            if (fileStats.isFile) {
                let fileNameList = file.split('_');
                let contentKeyBuffer = new Buffer.from(fileNameList[2], 'base64')
                let contentKey = contentKeyBuffer.toString();
                let newList = [].concat(fileNameList);
                newList[2] = contentKey;
                let showFileName = _.join(newList, '_');
                list.push({
                    fileName: file,
                    showFileName,
                    contentKey,
                    ctime: new Date(fileStats.ctime).getTime()
                });
            }
        });
        list = _.sortBy(list, 'ctime');
    } catch (err) {
        console.error(err);
    }
    // console.log(list);
	res.status(200).json({
		"success": true,
        "data": list
	});
});

router.get('/getTraceFile', function(req, res, next) {
    let fileName = sanitize(req.query.fileName);
    let folderName = sanitize(req.query.folderName);
    const filePath = path.join(basicFloder, folderName, fileName);
    console.log('filePath: %s', filePath);

    let output = '';
    const readStream = fs.createReadStream(filePath);

    readStream.on('data', function(chunk) {
        output += chunk.toString('utf8');
    });

    readStream.on('end', function() {
        console.log('finished reading');
        // write to file here.
        let result = output;
        let resList = result.split('------\n');
        let traceData = JSON.parse(_.head(resList));
        let cpuEventStrs = _.slice(resList, 1);
        let cpuEventsList = [];
        let cpuEvents = [];
        
        _.forEach(cpuEventStrs, (str) => {
            try {
                cpuEventsList.push(JSON.parse(str))
            } catch (error) {
                console.error('error: %s', error);
            }
        });
        cpuEvents = _.map(cpuEventsList, 'labels');
        _.forEach(cpuEvents, item => {
            try {
                item.cpuEvents = JSON.parse(item.cpuEvents);
                item.javaFutexEvents = JSON.parse(item.javaFutexEvents);
                item.transactionIds = JSON.parse(item.transactionIds);
                item.spans = JSON.parse(item.spans);
                item.innerCalls = JSON.parse(item.innerCalls);
            } catch (error) {
                console.error('error: %s', error, item);
            }
        });
        
        let finalResult = {
            trace: traceData,
            cpuEvents: cpuEvents
        };

        res.status(200).json({
            "success": true,
            "data": finalResult
        });
    });

    // fs.readFile(filePath, function(err, buffer) {
    //     if (err) {
    //         console.error("Error: ", err);
    //         res.status(500).json({
    //             success: false,
    //             message: err
    //         });
    //     }
    //     let result = buffer.toString('utf-8');
    //     let resList = result.split('------');
    //     let traceData = JSON.parse(_.head(resList));
    //     let cpuEventStrs = _.slice(resList, 1);
    //     let cpuEventsList = [];
    //     let cpuEvents = [];
        
    //     _.forEach(cpuEventStrs, (str) => {
    //         try {
    //             cpuEventsList.push(JSON.parse(str))
    //         } catch (error) {
    //             console.error('1', error);
    //         }
    //     });
    //     cpuEvents = _.map(cpuEventsList, 'labels');
    //     _.forEach(cpuEvents, item => {
    //         try {
    //             item.cpuEvents = JSON.parse(item.cpuEvents);
    //             item.javaFutexEvents = JSON.parse(item.javaFutexEvents);
    //             item.transactionIds = JSON.parse(item.transactionIds);
    //         } catch (error) {
    //             console.error('2', error, item);
    //         }
    //     });
        
    //     let finalResult = {
    //         trace: traceData,
    //         cpuEvents: cpuEvents
    //     };

    //     res.status(200).json({
    //         "success": true,
    //         "data": finalResult
    //     });
    // });
});

module.exports = router;