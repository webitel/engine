/**
 * Created by i.navrotskyj on 12.02.2016.
 */
'use strict';

var crm = require('./eslEvents/crm'),
    log = require(__appRoot + '/lib/log')(module);

module.exports = handleWConsoleEvent;

function handleWConsoleEvent (application) {
    application.WConsole.on('webitel::event::event::**', crm);
};