/**
 * Created by igor on 12.04.16.
 */

var calendarService = require(__appRoot + '/services/calendar'),
    CodeError = require(__appRoot + '/lib/error');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/calendars', list);
    api.get('/api/v2/calendars/:id', item);
    api.post('/api/v2/calendars', create);
    api.put('/api/v2/calendars/:id', update);
    api.delete('/api/v2/calendars/:id', remove);
};

function list(req, res, next) {
    var options = {
        domain: req.query['domain']
    };

    calendarService.list(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
};

function item (req, res, next) {
    var options = {
        id: req.params.id,
        domain: req.query['domain']
    };

    calendarService.item(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        };

        if (!result)
            return res
                .status(401)
                .json({
                    "status": "error",
                    "info": "Not found."
                });

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
};
function create (req, res, next) {
    var options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    calendarService.create(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
};

function update (req, res, next) {
    var options = {
        "id": req.params.id,
        "data": req.body
    };

    if (req.query['domain'])
        options.data.domain = req.query['domain'];

    calendarService.update(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
};

function remove (req, res, next) {
    var options = {
        "id": req.params.id,
        "domain": req.query['domain']
    };

    calendarService.remove(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
};