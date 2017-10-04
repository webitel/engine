/**
 * Created by i.navrotskyj on 09.12.2015.
 */
'use strict';

const configureService = require(__appRoot + '/services/configure');
const tcpDumpService = require(__appRoot + '/services/tcpDump');
const getRequest = require(__appRoot + '/utils/helper').getRequest;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.put('/api/v2/system/reload/xml', reloadXml);
    api.put('/api/v2/system/reload/:modName', reloadFsModule);
    api.put('/api/v2/system/cache/clear', cache);

    api.get('/api/v2/system/tcp_dump', listTcpDump);
    api.post('/api/v2/system/tcp_dump', createTcpDump);
    api.get('/api/v2/system/tcp_dump/:id', getTcpDump);
    api.delete('/api/v2/system/tcp_dump/:id', removeTcpDump);
    api.put('/api/v2/system/tcp_dump/:id', updateTcpDump);

    //  V1
    api.get('/api/v1/reloadxml', reloadXml);
}

function listTcpDump(req, res, next) {
    tcpDumpService.list(req.webitelUser, getRequest(req), (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function createTcpDump(req, res, next) {
    tcpDumpService.create(req.webitelUser, req.body, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}

function getTcpDump(req, res, next) {
    const option = {
        id: req.params.id
    };

    tcpDumpService.get(req.webitelUser, option, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}

function removeTcpDump(req, res, next) {
    const option = {
        id: req.params.id
    };

    tcpDumpService.remove(req.webitelUser, option, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}

function updateTcpDump(req, res, next) {
    const option = req.body;
    option.id = req.params.id;

    tcpDumpService.update(req.webitelUser, option, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}


function reloadXml (req, res, next) {
    configureService.reloadXml(req.webitelUser, (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function reloadFsModule(req, res, next) {
    configureService.reloadMod(req.webitelUser, req.params.modName,  (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function cache(req, res, next) {
    configureService.cache(req.webitelUser, {},  (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}