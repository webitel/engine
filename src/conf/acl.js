/**
 * Created by Igor Navrotskyj on 01.09.2015.
 */

'use strict';

/*
    c = create new element;
    u = update element;
    r = read element;
    d = delete element;
 */

module.exports = [{
    roles: 'root',
    allows: [{
        resources: [
            'blacklist',
            'rotes/default', 'rotes/public', 'rotes/extension', 'rotes/domain',
            'channels',
            'cc/tiers', 'cc/members', 'cc/queue',
            'book',
            'cdr',
            'outbound/list',
            'gateway', 'gateway/profile',
            'domain', 'domain/item',
            'account',
            'system/reload'
        ],
        permissions: '*'
    }]
}, {
    roles: 'admin',
    allows: [{
        resources: [
            'blacklist',
            'rotes/default', 'rotes/public', 'rotes/extension', 'rotes/domain',
            'channels',
            'cc/tiers', 'cc/members', 'cc/queue',
            'book',
            'cdr',
            'outbound/list',
            'gateway',
            'domain/item',
            'account'
        ],
        permissions: ['*']
    }]
}, {
    roles: 'user',
    allows: [{
        resources: 'blacklist',
        permissions: ['c', 'u', 'r']
    }, {
        resources: 'channels',
        permissions: ['*']
    }, {
        resources: 'book',
        permissions: ['c', 'u', 'r']
    }, {
        resources: 'cdr',
        permissions: ['r']
    }, {
        resources: 'cc/tiers',
        permissions: ['r']
    }]
}];