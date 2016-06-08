/**
 * Created by i.navrotskyj on 15.03.2016.
 */
'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    log = require(__appRoot + '/lib/log')(module)
    ;

var Service = {
    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    list: function (caller, option, cb) {
        checkPermissions(caller, 'hook', 'r', function (err) {
            if (err)
                return cb(err);
            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            var domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };

            var db = application.DB._query.hook;
            return db.search(domain, option, cb);
        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    item: function (caller, option, cb) {
        checkPermissions(caller, 'hook', 'r', function (err) {
            if (err)
                return cb(err);
            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            var domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };

            var db = application.DB._query.hook;
            return db.item(option.id, domain, {}, cb);
        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    update: function (caller, option, cb) {
        checkPermissions(caller, 'hook', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || !option.id)
                return cb(new CodeError(400, "Bad request options"));

            let id = option.id,
                _domain = option.domain,
                doc = option.doc;

            var domain = validateCallerParameters(caller, _domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };

            if (!doc.event)
                return cb(new CodeError(400, 'Bad request: event name is required.'));

            // TODO add type email, sms ... and validate
            if (!doc.action || !doc.action.type || !doc.action.url || !doc.action.method)
                return cb(new CodeError(400, 'Bad request: action options is required.'));

            let _doc = {
                domain: domain,
                event: doc.event,
                enable: doc.enable,
                description: doc.description,
                action: doc.action,
                fields: doc.fields,
                map: doc.map,
                filter: doc.filter
            };

            var db = application.DB._query.hook;

            Service._unBindBroker(id, domain, (err) => {
                if (err)
                    return cb(err);

                return db.update(id, domain, _doc, (err, res) => {
                    if (err)
                        return cb(err);
                    application.Broker.bindHook(_doc.event);
                    return cb(null, res);
                });
            });
        });
    },
    
    _unBindBroker: function (id, domain, cb) {
        var db = application.DB._query.hook;
        db.item(id, domain, {"event": 1, "_id": 0}, (e, res) => {
            if (e) return cb(e);
            db.count({"event": res.event}, (e, count) => {
                if (e) return cb(e);

                if (count === 1)
                    application.Broker.unBindHook(res.event);

                return cb(null, count === 1);
            })
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: function (caller, option, cb) {
        checkPermissions(caller, 'hook', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let _domain = option.domain,
                doc = option.doc;

            var domain = validateCallerParameters(caller, _domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };

            if (!doc.event)
                return cb(new CodeError(400, 'Bad request: event name is required.'));

            // TODO add type email, sms ... and validate
            if (!doc.action || !doc.action.type || !doc.action.url || !doc.action.method)
                return cb(new CodeError(400, 'Bad request: action options is required.'));

            let _doc = {
                domain: domain,
                event: doc.event,
                enable: doc.enable,
                description: doc.description,
                action: doc.action,
                fields: doc.fields,
                map: doc.map,
                filter: doc.filter
            };

            var db = application.DB._query.hook;
            return db.create(_doc, (e, res) => {
                if (e) return cb(e);
                application.Broker.bindHook(_doc.event);
                return cb(null, res);
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    remove: function (caller, option, cb) {
        checkPermissions(caller, 'hook', 'd', function (err) {
            if (err)
                return cb(err);
            if (!option || !option.id)
                return cb(new CodeError(400, "Bad request options"));

            var domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };

            Service._unBindBroker(option.id, domain, (err) => {
                if (err)
                    return cb(err);

                let db = application.DB._query.hook;
                return db.remove(option.id, domain, cb);
            });
        });
    },
};

module.exports = Service;