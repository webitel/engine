/**
 * Created by Igor Navrotskyj on 22.09.2015.
 */

'use strict';
module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/location/:number', getLocationNumber);
};

function getLocationNumber (req, res, next) {
    var db = application.DB._query.location;
    var number = req.params['number'];
    if (number)
        number = number.replace(/[\D]/g, '');
    db.find(number, function (err, arr) {
        if (err)
            return next(err);

        return res
            .status(200)
            .json(arr);
    });
}