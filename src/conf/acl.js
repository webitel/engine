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
    'calendar',

    'rotes/default',
    'rotes/public',
    'rotes/extension',
    'rotes/domain',

    'channels',

    'cc/tiers',
    'cc/members',
    'cc/queue',

    'dialer',
    'dialer/members',
    'dialer/templates',

    'book',
    'hook',

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

    'license',

    'vmail',
    'widget',

    'metadata',

    'hotdesk'
];

module.exports = [
    {
        roles: 'root',
        allows: {
            'acl/roles': ["*"],
            'acl/resource': ["*"],

            'blacklist': ["*"],
            'calendar': ["*"],

            // TODO rename !!!
            'rotes/default': ["*"],
            'rotes/public': ["*"],
            'rotes/extension': ["*"],
            'rotes/domain': ["*"],

            'channels': ["*"],

            'cc/tiers': ["*"],
            'cc/members': ["*"],
            'cc/queue': ["*"],

            'dialer': ["*"],
            'dialer/members': ["*"],
            'dialer/templates': ["*"],

            'book': ["*"],
            'hook': ["*"],
            'vmail': ["*"],

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
            'license': ["*"],
            'widget': ["*"],
            'callback': ["*"],
            'callback/members': ["*"],

            'metadata': ["*"],

            'hotdesk': ["*"],

            'kibana': ["*"]
        }
    },
    {
        roles: 'user',
        allows: {
            'account': ['r', 'uo'],
            'acl/roles': ['r'],
            'blacklist': ['c', 'u', 'r'],
            'channels': ['*'],
            'book': ['c', 'u', 'r'],
            'vmail': ['ro'],
            'cdr': ['ro'],
            'cdr/files': ['ro'],
            'cc/tiers': ["r"],
            'metadata': ["r"],
            'hotdesk': ["*"],

            'kibana': ["r"]
        }
    },
    {
        roles: 'admin',
        parents: 'user',
        allows: {
            'blacklist': ["*"],
            'calendar': ["*"],

            'rotes/default': ["*"],
            'rotes/public': ["*"],
            'rotes/extension': ["*"],
            'rotes/domain': ["*"],

            'channels': ["*"],

            'cc/tiers': ["*"],
            'cc/members': ["*"],
            'cc/queue': ["*"],

            'dialer': ["*"],
            'dialer/members': ["*"],
            'dialer/templates': ["*"],

            'book': ["*"],
            'hook': ["*"],
            'vmail': ["*"],

            'cdr': ["*"],
            'cdr/files': ["*"],
            'cdr/media': ["*"],

            'outbound/list': ["*"],

            'gateway': ["*"],

            'domain/item': ["*"],
            'domain': ["ro", "uo"],

            'account': ["*"],

            'widget': ["*"],
            'callback': ["*"],
            'callback/members': ["*"],

            'metadata': ["*"],

            'hotdesk': ["*"],

            'kibana': ["*"]
        }
    }
];