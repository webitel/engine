/**
 * Created by Admin on 04.08.2015.
 */
'use strict';

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.all('/*', function(req, res, next) {
        // CORS headers
        res.header("X-Powered-By", "Webitel");
        var origin = (req.headers.origin || "*");
        res.header("Access-Control-Allow-Origin", "*"); // restrict it to the required domain
        res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,PATH,DELETE,OPTIONS');

        res.header('Access-Control-Allow-Headers', 'Content-type,Accept,X-Access-Token,X-Key');
        if (req.method == 'OPTIONS') {
            res.status(200).end();
        } else {
            next();
        }
    });
}