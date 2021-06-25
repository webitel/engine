/**
 * Created by i.navrotskyj on 23.02.2016.
 */
'use strict';

module.exports = (req) => {
    return req.headers['x-forwarded-for'] ||
        req.connection.remoteAddress ||
        req.socket.remoteAddress ||
        (req.connection.socket && req.connection.socket.remoteAddress);
}
