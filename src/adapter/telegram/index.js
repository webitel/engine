/**
 * Created by igor on 27.06.16.
 */
// TODO add support workers
"use strict";

let TelegramBot = require('node-telegram-bot-api'),
    log = require(__appRoot + '/lib/log')(module),
    conf = require(__appRoot + '/conf'),
    enabled = conf.get('telegram:enabled'),
    authService = require(__appRoot + '/services/auth')
    ;

const HELP_MSG =
    `/login <username> <password> - login webitel bot;
/logout - Logout webitel;
/busy - Set status busy;
/ready - Set status ready;`;


class TelegramAdapter {
    constructor(token, app) {
        this.bot = new TelegramBot(token, {polling: true});
        this.bot.onText(/.*/, this.onText.bind(this));
        this.query = app.DB._query.telegram;
        this._routes = [];
        app.Broker.on('telegramEvent', this.sendNotification.bind(this))
    }

    sendNotification (e = {}) {
        let userName = e['Channel-Presence-ID'],
            message = e._body
            ;
        if (!userName || !message) {
            return log.warn(`bad event ${e}`);
        }

        this.query.getByUserId(userName, (err, res) => {
            if (err)
                return log.error(err);

            if (!res) {
                return log.debug(`Not found session ${userName}`);
            }
            log.trace(`Send ${userName} to chat ${res.chatId}`);
            this.sendMessage(res.chatId, message);
        })
    }

    onText(msg) {
        let found = false;
        this._routes.forEach((reg) => {
            const result = reg.regexp.exec(msg.text);
            if (result) {
                found = true;
                if (reg.skipAuth || msg.session) {
                    reg.handler(msg, result);
                } else {
                    // TODO
                    this.getSession(msg.from.id, (err, res) => {
                        if (err) {
                            this.sendMessage(msg.from.id, err.message);
                            return log.error(err);
                        }

                        if (!res) {
                            return this.sendMessage(msg.from.id, `Please login.`);
                        }
                        msg.session = res;
                        reg.handler(msg, result);
                    });
                }
            }
        });

        if (!found)
            return this.sendMessage(msg.from.id, `Error 404 :\nYou can control me by sending these commands:\n${HELP_MSG}`);
    }

    sendMessage(fromId, msg) {
        this.bot.sendMessage(fromId, msg);
    }

    editMessageText (text, messageId, chatId) {
        this.bot.editMessageText(text, {message_id: messageId, chat_id: chatId})
    }

    getSession (id , cb) {
        this.query.getSession(id, cb);
    }

    addRoute (regexp, handler, options = {}) {
        this._routes.push({
            regexp: regexp,
            handler: handler,
            skipAuth: options.skipAuth
        })
    }
}

module.exports = (app) => {
    if ('' + enabled !== 'true') return;
    const bot = new TelegramAdapter(conf.get('telegram:token'), app);
    let query = app.DB._query.telegram;
    bot.addRoute(/\login\s(.+)\s(.+)/, login, {skipAuth: true});
    bot.addRoute(/\help|\/start/, help, {skipAuth: true});
    bot.addRoute(/\logout/, logout);
    bot.addRoute(/\/busy\s?(.+)?/, busy);
    bot.addRoute(/\/ready/, ready);

    function help(msg) {
        let fromId = msg.from.id;
        return bot.sendMessage(fromId, HELP_MSG);
    }

    function login(msg, match) {
        let fromId = msg.from.id,
            username = match[1],
            password = match[2];

        if (!username || !password) {
            log.warn(`Bad credentials: ${fromId}`);
            return bot.sendMessage(fromId, 'Invalid credentials.');
        }
        // bot.editMessageText('test', msg.message_id, fromId);
        authService.login({
            username: username,
            password: password
        }, function (err, result) {
            if (err) {
                log.error(err.message);
                return bot.sendMessage(fromId, err.message);
            }

            let session = {
                createdOn: Date.now(),
                chatId: fromId,
                user: username,
                telegram: msg.from
            };

            query.create(session, (err, res) => {
                if (err && err.code == 11000) {
                    log.debug(err.message);
                    return bot.sendMessage(fromId, `Session active.`);
                }
                if (err) {
                    log.error(err.message);
                    return bot.sendMessage(fromId, err.message);
                }

                return bot.sendMessage(fromId, `Hello ${username}, subscribed.`);
            })
        });

    }

    function logout(msg, match) {
        let fromId = msg.from.id;
        query.removeSession(fromId, msg.session && msg.session.user, (e, c) => {
            if (e) {
                log.error(e.message);
                return bot.sendMessage(fromId, err.message);
            }
            return bot.sendMessage(fromId, `good bye.`);
        });
    }

    function busy(msg, match) {
        let status = match[1] || 'ONBREAK';
        let fromId = msg.from.id;
        app.WConsole.setAccountStatus(msg.session.user, status, (res) => {
            if (/^-/.test(res.body)) {
                log.error(res.body);
                return bot.sendMessage(fromId, res.body);
            }

            return bot.sendMessage(fromId, `Busy status: ${status}`);
        })
    }
    
    function ready(msg) {
        let fromId = msg.from.id;
        app.WConsole.setAccountStatus(msg.session.user, 'ONHOOK', (res) => {
            if (/^-/.test(res.body)) {
                log.error(res.body);
                return bot.sendMessage(fromId, res.body);
            }

            return bot.sendMessage(fromId, `On ready.`);
        })
    }
};