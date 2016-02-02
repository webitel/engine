/**
 * Created by Igor Navrotskyj on 01.09.2015.
 */

'use strict';

/*
    c = create new element;
    u = update element;
    r = read element;
    d = delete element;
    <x>o - only user element;
 */

const RESOURCES = [
    'acl/roles',
    'acl/resource',

    'blacklist',

    'rotes/default',
    'rotes/public',
    'rotes/extension',
    'rotes/domain',

    'channels',

    'cc/tiers',
    'cc/members',
    'cc/queue',

    'book',

    'cdr',
    'cdr/files',
    'cdr/media',

    'outbound/list',

    'gateway',
    'gateway/profile',

    'domain',
    'domain/item',

    'account',

    'system/reload',

    'license'
];

module.exports = [
    {
        roles: 'root',
        allows: {
            'acl/roles': ["*"],
            'acl/resource': ["*"],

            'blacklist': ["*"],

            'rotes/default': ["*"],
            'rotes/public': ["*"],
            'rotes/extension': ["*"],
            'rotes/domain': ["*"],

            'channels': ["*"],

            'cc/tiers': ["*"],
            'cc/members': ["*"],
            'cc/queue': ["*"],

            'book': ["*"],

            'cdr': ["*"],
            'cdr/files': ["*"],
            'cdr/media': ["*"],

            'outbound/list': ["*"],

            'gateway': ["*"],
            'gateway/profile': ["*"],

            'domain': ["*"],
            'domain/item': ["*"],

            'account': ["*"],

            'system/reload': ["*"],
            'license': ["*"]
        }
    },
    {
        roles: 'user',
        allows: {
            'account': ['r', 'uo'],
            'blacklist': ['c', 'u', 'r'],
            'channels': ['*'],
            'book': ['c', 'u', 'r'],
            'cdr': ['ro'],
            'cdr/files': ['ro'],
            'cc/tiers': ["r"]
        }
    },
    {
        roles: 'admin',
        parents: 'user',
        allows: {
            'blacklist': ["*"],

            'rotes/default': ["*"],
            'rotes/public': ["*"],
            'rotes/extension': ["*"],
            'rotes/domain': ["*"],

            'channels': ["*"],

            'cc/tiers': ["*"],
            'cc/members': ["*"],
            'cc/queue': ["*"],

            'book': ["*"],

            'cdr': ["*"],
            'cdr/files': ["*"],
            'cdr/media': ["*"],

            'outbound/list': ["*"],

            'gateway': ["*"],

            'domain/item': ["*"],

            'account': ["*"]
        }
    }
];