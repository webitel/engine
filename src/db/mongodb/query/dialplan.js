/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    defaultCollectionName = conf.get('mongodb:collectionDefault'),
    publicCollectionName = conf.get('mongodb:collectionPublic'),
    extensionCollectionName = conf.get('mongodb:collectionExtension'),
    domainVariableCollectionName = conf.get('mongodb:collectionDomainVar'),
    ObjectID = require('mongodb').ObjectID
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    var Dialplan = {
        /**
         * Internal extension
         */

        /**
         *
         * @param number
         * @param domain
         * @param cb
         * @returns {Promise}
         */
        existsExtension: function (number, domain, cb) {
            let _numberArray;
            // TODO �������� �� �������
            if (number instanceof Array) {
                _numberArray = number;
            } else {
                _numberArray = [number];
            };

            return db
                .collection(extensionCollectionName)
                .findOne({
                        "destination_number": {
                            "$in": _numberArray
                        },
                        "domain": domain
                    },
                    function (err, res) {
                        if (err)
                            return cb(err);

                        return cb(
                            null,
                            res ? true : false
                        );
                    }
                );
        },
        updateOrInsertExtension: function (userId, domain, userExtension, cb) {
            try {
                var collection = db.collection(extensionCollectionName);
                collection.update(
                    {
                        "userRef": userId + '@' + domain
                    },
                    userExtension,
                    {upsert: true},
                    cb
                );
            } catch (e) {
                cb(e);
            };
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {*}
         */
        getExtensions: function (domain, cb) {
            return Dialplan.getDataFromCollection({
                domain: domain,
                collectionName: extensionCollectionName
            }, cb);
        },

        /**
         *
         * @param _id
         * @param data
         * @param cb
         * @returns {number}
         */
        updateExtension: function (_id, data, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(extensionCollectionName)
                .findAndModify({"_id": new ObjectID(_id)}, [], data, function (err, result) {
                    if (err) return cb(err);

                    if (result) {
                        cb(null, {
                            "status": result.ok == 1 ? "OK" : "error",
                            "data": result.value
                        })
                    }
                });

            return 1;
        },

        /**
         *
         * @param username
         * @param domain
         * @param cb
         * @returns {Promise}
         */
        removeExtensionFromUser: function (username, domain, cb) {
            return db
                .collection(extensionCollectionName)
                .remove({
                    "userRef": username,
                    "domain": domain
                }, cb);
        },

        /**
         * 
         * @param dialplan
         * @param cb
         * @returns {number}
         */
        createExtension: function (dialplan, cb) {
            db
                .collection(extensionCollectionName)
                .insert(dialplan, function (err, res) {
                    var result = res && res['ops'];
                    if (result instanceof Array) {
                        result = result[0];
                    };
                    cb(err, result);
                });

            return 1;
        },

        /**
         * Default dialplan.
         **/

        /**
         *
         * @param dialplan
         * @param cb
         * @returns {number}
         */
        createDefault: function (dialplan, cb) {
            db
                .collection(defaultCollectionName)
                .insert(dialplan, function (err, res) {
                    var result = res && res['ops'];
                    if (result instanceof Array) {
                        result = result[0];
                    };
                    cb(err, result);
                });

            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {*}
         */
        getDefault: function (domain, cb) {
            return Dialplan.getDataFromCollection({
                domain: domain,
                collectionName: defaultCollectionName
            }, cb);
        },

        /**
         *
         * @param _id
         * @param cb
         * @returns {number}
         */
        removeDefault: function (_id, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(defaultCollectionName)
                .remove({"_id": new ObjectID(_id)}, function (err, res) {
                    if (err) return cb(err);

                    var result = res && res['result'];
                    if (result) {
                        result = {
                            "status": (result['ok'] == 1) ? "OK" : "error",
                            "info": result["n"]
                        };
                    };

                    cb(err, result || res);
                });

            return 1;
        },

        /**
         *
         * @param _id
         * @param data
         * @param cb
         * @returns {number}
         */
        updateDefault: function (_id, data, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(defaultCollectionName)
                .findAndModify({"_id": new ObjectID(_id)}, [], data, function (err, result) {
                    if (err) return cb(err);

                    if (result)
                        return cb(null, result.value);
                });

            return 1;
        },

        /**
         *
         * @param domain
         * @param option
         * @param cb
         * @returns {number}
         */
        incOrderDefault: function (domain, option, cb) {
            db
                .collection(defaultCollectionName)
                .update({
                        "domain": domain,
                        "order": {
                            "$gt": option['start']
                        }
                    },
                    {
                        $inc: {
                            "order": option['inc']
                        }
                    },
                    {
                        multi: true
                    },
                    cb
            );

            return 1;
        },

        /**
         *
         * @param _id
         * @param option
         * @param cb
         * @returns {number}
         */
        setOrderDefault: function (_id, option, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(defaultCollectionName)
                .update({"_id": new ObjectID(_id)}, {"$set": {"order": option['order']}}, cb);

            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        removeDefaultByDomain: function (domain, cb) {
            db
                .collection(defaultCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        },

        /**
         * Public dialplan
         */

        /**
         *
         * @param dialplan
         * @param cb
         * @returns {number}
         */
        createPublic: function (dialplan, cb) {
            db
                .collection(publicCollectionName)
                .insert(dialplan, function (err, res) {
                    if (err) return cb(err);

                    var result = res && res['ops'];
                    if (result instanceof Array && result.length == 1) {
                        result = result[0];
                    };
                    cb(null, result);
                });

            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        getPublic: function (domain, cb) {
            db
                .collection(publicCollectionName)
                .find({"domain": domain})
                .toArray(cb);

            return 1;
        },

        /**
         *
         * @param _id
         * @param cb
         * @returns {number}
         */
        removePublic: function (_id, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(publicCollectionName)
                .remove({"_id": new ObjectID(_id)}, function (err, res) {
                    if (err) return cb(err);

                    var result = res && res['result'];
                    if (result) {
                        result = {
                            "status": (result['ok'] == 1) ? "OK" : "error",
                            "info": result["n"]
                        };
                    };

                    cb(err, result || res);
                });

            return 1;
        },

        /**
         *
         * @param _id
         * @param data
         * @param cb
         * @returns {number}
         */
        updatePublic: function (_id, data, cb) {
            if (!validateId(_id, cb)) return 0;
            db
                .collection(publicCollectionName)
                .findAndModify({"_id": new ObjectID(_id)}, [], data, function (err, result) {
                    if (err) return cb(err);

                    if (result)
                        return cb(null, result.value);
                });

            return 1;
        },

        /*
         *
         */
        getDataFromCollection: function (option, cb) {
            var domain = option['domain'],
                collectionName = option['collectionName'];
            db
                .collection(collectionName)
                .find({"domain": domain})
                .sort({"order": 1})
                .toArray(cb);

            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        removePublicByDomain: function (domain, cb) {
            db
                .collection(publicCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        },

        /**
         *  Domain variables
         */

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        getDomainVariable: function (domain, cb) {
            db
                .collection(domainVariableCollectionName)
                .find({"domain": domain})
                .toArray(cb);
            return 1;
        },

        /**
         *
         * @param domain
         * @param variables
         * @param cb
         * @returns {number}
         */
        insertOrUpdateDomainVariable: function (domain, variables, cb) {
            var doc = {
                "variables": variables,
                "domain": domain
            };
            db
                .collection(domainVariableCollectionName)
                .update({
                        "domain": domain
                    },
                    doc,
                    {upsert: true},
                    function(err, res) {
                        if (err) return cb(err);
                        if (!res || !res.result) {
                            return cb(null, {
                                "status": "OK"
                            });
                        };

                        cb(null, {
                            "status": res.result.ok == 1 ? "OK" : "error",
                            "data": res.result
                        });
                    }
            );
            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        removeVariablesByDomain: function (domain, cb) {
            db
                .collection(domainVariableCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    };

    return Dialplan;
};

function validateId(id, cb) {
    if (ObjectID.isValid(id)) {
        return true;
    };
    cb(new CodeError(400, "Bad ObjectID."));
    return false;
};