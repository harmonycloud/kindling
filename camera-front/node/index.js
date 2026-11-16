'use strict';

const express = require('express');
const app = express();
const http = require('http');
const serveStatic = require('serve-static');
const bodyParser = require('body-parser');
const _ = require('lodash');
const async = require('async');
const path = require('path');
const compression = require('compression');
const rateLimit = require('express-rate-limit');
const history = require('connect-history-api-fallback');
const { createProxyMiddleware } = require('http-proxy-middleware');

const settings = require('./settings');
const port = settings.port;

const profileAddress = 'http://' + settings.profileConfig.host + ':' + settings.profileConfig.port;
const profileProxy = createProxyMiddleware('/profile', {
    target: profileAddress,
    secure: false,
    changeOrigin: true,
});
app.use(profileProxy);

const limiter = rateLimit({
    windowMs: settings.ratelimit.windowMs,
    max: settings.ratelimit.max
});
app.use(limiter);

app.use(
    bodyParser.urlencoded({
        extended: true,
    })
);
app.use(bodyParser.json());
app.use(compression());
// app.use(history());

app.use(express.static('../dist'));
console.log(__dirname);
// app.set('views', path.join(__dirname, 'dist'));
// app.use(serveStatic(path.join(__dirname, 'dist')));

app.get("/test", function(req, res, next) {
    // res.render("test.html");
    res.status(200).json({
        test: true
    });
});

app.use('/file', require('./routers/file.js'));
app.use('/esserver', require('./routers/esServer.js'));

var httpServer = http.createServer(app);
// var httpsServer = https.createServer(credentials, app);
httpServer.listen(port, function() {
    console.log('Express http server listening on port ' + port);
});