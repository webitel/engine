/**
 * Created by i.navrotskyj on 11.03.2016.
 */
'use strict';

var EventEmitter2 = require('eventemitter2').EventEmitter2,
    plainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    Collection = require('./lib/collection');


class AutoDialer extends EventEmitter2 {
    constructor () {
        super();
    }

    run() {

    }
};

class Campaign extends EventEmitter2 {
    constructor () {
        super();
        this.runinng = false;
        this.agents = new Collection('id');
        this.domain = '10.10.10.144';
        this.init();
    };

    init () {
        // Load db
        // Load users
        async.waterfall([
            function initAgents(cb) {
                application.WConsole.listUser(null, this.domain, (err, agentsTxt)=>{
                    if (err)
                        return cb(err);
                    plainTableToJSON(agentsTxt, null, cb);
                })
            }
        ], (err, res) => {

        });
    };


}