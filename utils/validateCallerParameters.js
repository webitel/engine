/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

// todo permission.
module.exports = function (caller, domain) {
    if (!caller) {
        return null;
    };

    return caller['domain'] || domain;
};