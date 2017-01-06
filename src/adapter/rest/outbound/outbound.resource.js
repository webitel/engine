/**
 * Created by Igor Navrotskyj on 17.09.2015.
 */

'use strict';

var outboundService = require(__appRoot + '/services/outboundQueue');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    //api.post('/api/v2/outbound/:name', postItem);
};

function postItem(req, res, next) {
    var data = req.body;

    if (req.query['domain']) {
        data['domain'] = req.query['domain'];
    };

    data['name'] = req.params['name'];

    outboundService.create(req.webitelUser, data,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}