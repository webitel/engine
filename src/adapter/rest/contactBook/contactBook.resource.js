/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

const bookService = require(__appRoot + '/services/contactBook');
const getRequest = require(__appRoot + '/utils/helper').getRequest;

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/api/v2/contacts/communications', createType);
    api.get('/api/v2/contacts/communications', listTypes);
    api.delete('/api/v2/contacts/communications/:id', removeType);
    api.put('/api/v2/contacts/communications/:id', updateType);

    api.get('/api/v2/contacts/yealink', testYealink);
    //api.get('/api/v2/contacts/v-card', testVCard);

    api.get('/api/v2/contacts', listBook);
    api.get('/api/v2/contacts/:id', itemBook);
    api.post('/api/v2/contacts', createBook);
    api.put('/api/v2/contacts/:id', updateItem);
    api.delete('/api/v2/contacts/:id', deleteItem);

    // api.post('/api/v2/contacts/searches', searches);


    // api/v2/contacts/:id/communications -user comm
    // api/v2/contacts/communications - all com
    // api/v2/contacts/tags - all tags
    // api/v2/contacts/tags/:name - all contact by tag

}

function testYealink(req, res, next) {
    application.PG.getQuery('contacts').importData.yeaLink(req.query.domain, (err, result) => {
        if (err)
            return next(err);

        res.set('Content-Type', 'text/xml');
        res.send(result)
    });
}

function testVCard(req, res, next) {
    application.PG.getQuery('contacts').importData.vCard(req.query.domain, (err, result) => {
        if (err)
            return next(err);

        res.set('Content-Type', 'text/v-card');
        res.send(result)
    });
}

function listBook (req, res, next) {
    bookService.list(req.webitelUser, getRequest(req), (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function itemBook (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    bookService.getById(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function createBook (req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    bookService.create(req.webitelUser, options, (err, result) => {
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

function updateItem (req, res, next) {
    const options = req.body;
    options.id = req.params.id;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    bookService.updateItem(req.webitelUser, options, (err, result) => {
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

function deleteItem (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    bookService.removeItem(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function createType(req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    bookService.types.create(req.webitelUser, options, (err, result) => {
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

function listTypes (req, res, next) {
    bookService.types.list(req.webitelUser, getRequest(req), (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function removeType(req, res, next) {
    const options = {
        domain: req.query.domain,
        id: req.params.id
    };
    bookService.types.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function updateType(req, res, next) {
    const options = req.body;
    options.domain = req.query.domain;
    options.id = req.params.id;

    bookService.types.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}