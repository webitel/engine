/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    emailCollectionName = conf.get('mongodb:collectionEmail');

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        get: function (option, cb) {
            var domain = option['domain'];
            db
                .collection(emailCollectionName)
                .findOne({"domain": domain}, cb);

            return 1;
        },
        
        set: function (settings, cb) {
            var domain = settings['domain'];

            db
                .collection(emailCollectionName)
                .update(
                    {"domain": domain},
                    settings,
                    {upsert: true},
                    function (err) {
                        if (err) {
                            return cb(err);
                        };

                        db
                            .collection(emailCollectionName)
                            .findOne({"domain": domain}, cb);
                });
        },
        
        update: function (settings, cb) {
            var domain = settings['domain'];

            db
                .collection(emailCollectionName)
                .update(
                    {"domain": domain},
                    settings,
                    function (err, res) {
                        if (err) {
                            return cb(err);
                        };

                        if (res && res.result && res.result.nModified === 0) {
                            return cb(new CodeError(404, 'Not found.'));
                        };

                        db
                            .collection(emailCollectionName)
                            .findOne({"domain": domain}, cb);
                });
        },

        remove: function (option, cb) {
            var domain = option['domain'];
            db
                .collection(emailCollectionName)
                .remove({"domain": domain}, cb);

            return 1;
        },
        
        removeByDomain: function (domain, cb) {
            db
                .collection(emailCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    };
};