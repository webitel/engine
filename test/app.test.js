/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

var request = require('supertest');
var assert = require('assert');
var uuid = require('node-uuid');

describe('Routing', function() {
    var url = 'http://10.10.10.25:10022';
    var wsServer = 'ws://10.10.10.25:10022';
    var ROOT_PASSWORD = 'ROOT_PASSWORD';
    var userCredentials = {};
    var testConfig = {
        domain: 'asd',
        user: {
            number: '100',
            password: '100',
            role: 'admin'
        },
        cc: {
            queue: "AUTO_TEST_QUEUE"
        },
        callNumber: '00'
    };
    before(function(done) {
        //new MongoClient(config.get('mongodb:uri'), function () {
        done();
        //});

    });

    describe('REST', function() {
        it('Вход пользователя ROOT', function(done) {
            var profile = {
                username: 'root',
                password: 'ROOT_PASSWORD'
            };
            request(url)
                .post('/login')
                .send(profile)
                .end(function(err, res) {
                    if (err) {
                        throw err;
                    };
                    userCredentials = res.body;
                    if (userCredentials && userCredentials.key) {
                        done(null, res.body);
                    } else {
                        throw res.body.info
                    }
                    //console.dir(userCredentials)
                });
        });

        it('Получить состояние сервера', function (done) {
            request(url)
                .get('/api/v2/status')
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .expect(200, function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (typeof res.body.freeSWITCH === 'string') {
                        done();
                    } else {
                        throw 'Undef response'
                    }
                });
        });

        it('Создать домен', function (done) {
            var _r = {
                "domain_name": testConfig.domain,
                "customer_id": testConfig.domain
            };
            request(url)
                .post('/api/v2/domains')
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .send(_r)
                .end(function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (res.body.status === 'OK') {
                        done();
                    } else {
                        throw res.body.info
                    }
                });
        });

        it('Создать пользователя', function (done) {
            var _r = {
                "login": testConfig.user.number,
                "role": testConfig.user.role,
                "domain": testConfig.domain,
                "password": testConfig.user.password
            };
            request(url)
                .post('/api/v2/accounts')
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .send(_r)
                .end(function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (res.body.status === 'OK') {
                        done();
                    } else {
                        throw res.body.info
                    }
                });
        });

        it('Получить список пользователей', function (done) {
            request(url)
                .get('/api/v2/accounts?domain=' + testConfig.domain)
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .expect(200, function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (res.body.status === 'OK') {
                        done();
                    } else {
                        throw res.body.info
                    }
                });

        });

        /* it('Удалить пользователя', function (done) {
         request(url)
         .delete('/api/v2/accounts')
         .set('x-key', userCredentials.key)
         .set('x-access-token', userCredentials.token)
         .expect('Content-Type', /json/)
         .send(_r)
         .end(function (err, res) {
         if (err) {
         throw err;
         };
         if (res.body.status === 'OK') {
         done();
         } else {
         throw res.body.info
         }
         });
         });
         */

        /*  it('Звонок на номер ' + testConfig.callNumber, function (done) {
         var _r = {
         calledId: testConfig.callNumber,
         callerId: testConfig.user.number + '@' + testConfig.domain
         };
         request(url)
         .post('/api/v2/channels')
         .set('x-key', userCredentials.key)
         .set('x-access-token', userCredentials.token)
         .expect('Content-Type', /json/)
         .send(_r)
         .end(function (err, res) {
         if (err) {
         throw err;
         };
         if (res.body.status === 'OK') {
         done();
         } else {
         throw res.body.info
         }
         });
         }); */

        it('Удалить домен', function (done) {
            request(url)
                .del('/api/v2/domains/' + testConfig.domain)
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .expect(200, function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (res.body.status === 'OK') {
                        done();
                    } else {
                        throw 'Undef response'
                    }
                });
        });

        it('Выход пользователя ROOT', function (done) {
            request(url)
                .post('/logout')
                .set('x-key', userCredentials.key)
                .set('x-access-token', userCredentials.token)
                .expect('Content-Type', /json/)
                .expect(200, function (err, res) {
                    if (err) {
                        throw err;
                    };
                    if (res.body.status === 'OK') {
                        done();
                    } else {
                        throw 'Undef response'
                    }
                });
        });


        // TODO CC
        describe('MOD CC', function () {
            it('Вход пользователя ROOT', function(done) {
                var profile = {
                    username: 'root',
                    password: ROOT_PASSWORD
                };
                request(url)
                    .post('/login')
                    .send(profile)
                    .end(function(err, res) {
                        if (err) {
                            throw err;
                        };
                        userCredentials = res.body;
                        if (userCredentials && userCredentials.key) {
                            done(null, res.body);
                        } else {
                            throw res.body.info
                        }
                        //console.dir(userCredentials)
                    });
            });

            it('Создать домен', function (done) {
                var _r = {
                    "domain_name": testConfig.domain,
                    "customer_id": testConfig.domain
                };
                request(url)
                    .post('/api/v2/domains')
                    .set('x-key', userCredentials.key)
                    .set('x-access-token', userCredentials.token)
                    .expect('Content-Type', /json/)
                    .send(_r)
                    .end(function (err, res) {
                        if (err) {
                            throw err;
                        };
                        if (res.body.status === 'OK') {
                            done();
                        } else {
                            throw res.body.info
                        }
                    });
            });

            it('Создать очередь.', function (done) {
                var _r = {
                    name: testConfig.cc.queue
                };
                request(url)
                    .post('/api/v2/callcenter/queues?domain=' + testConfig.domain)
                    .set('x-key', userCredentials.key)
                    .set('x-access-token', userCredentials.token)
                    .send(_r)
                    .expect('Content-Type', /json/)
                    .expect(200, function (err, res) {
                        if (err) {
                            return done(err);
                        };
                        if (res.body.status == 'OK') {
                            done();
                        } else {
                            done(res.body);
                        }
                    });
            });

            it('Удалить очередь.', function (done) {
                request(url)
                    .del('/api/v2/callcenter/queues/' + testConfig.cc.queue + '?domain=' + testConfig.domain)
                    .set('x-key', userCredentials.key)
                    .set('x-access-token', userCredentials.token)
                    .expect('Content-Type', /json/)
                    .expect(200, function (err, res) {
                        if (err) {
                            done(err);
                        };
                        if (res.body.status == 'OK') {
                            done();
                        } else {
                            done(res.body);
                        }
                    });
            });

            it('Удалить домен', function (done) {
                request(url)
                    .del('/api/v2/domains/' + testConfig.domain)
                    .set('x-key', userCredentials.key)
                    .set('x-access-token', userCredentials.token)
                    .expect('Content-Type', /json/)
                    .expect(200, function (err, res) {
                        if (err) {
                            throw err;
                        };
                        if (res.body.status === 'OK') {
                            done();
                        } else {
                            throw 'Undef response'
                        }
                    });
            });

            it('Выход пользователя ROOT', function (done) {
                request(url)
                    .post('/logout')
                    .set('x-key', userCredentials.key)
                    .set('x-access-token', userCredentials.token)
                    .expect('Content-Type', /json/)
                    .expect(200, function (err, res) {
                        if (err) {
                            throw err;
                        };
                        if (res.body.status === 'OK') {
                            done();
                        } else {
                            throw 'Undef response'
                        }
                    });
            });
        });
    });

    describe('WSS', function () {
        var WebSocket = require('ws'),
            ws,
            apiQ = {}
            ;

        function exec (id, obj, cb) {
            apiQ[id] = cb;
            var cmd = {
                'exec-uuid': id,
                'exec-func': obj.func,
                'exec-args': obj.args
            };
            ws.send(JSON.stringify(cmd));
        };
        function onMessage (res) {
            //console.dir(res);
            var _res = JSON.parse(res);
            if (!_res['exec-uuid'])
                return;
            var fn = apiQ[_res['exec-uuid']];

            if (!fn)
                return;

            fn(_res);
            delete apiQ[_res['exec-uuid']];
        };

        before(function (done) {
            ws = new WebSocket(wsServer);
            ws.on('open', function open() {
                done();
            });

            ws.on('message', onMessage);

        });

        it('Подключиться', function (done) {
            var _login = {
                'func': 'auth',
                'args': {
                    'account': 'root',
                    'secret': 'ROOT_PASSWORD'
                }
            };
            exec(uuid.v4(), _login, function (res) {
                if (res['exec-complete'] === '+OK') {
                    done();
                } else {
                    done(res['exec-response']);
                };

            });
        });

        it ('Список доменов', function (done) {
            var _e = {
                'func': 'api domain list',
                'args': {
                }
            };
            exec(uuid.v4(), _e, function (res) {
                if (res['exec-complete'] === '+OK') {
                    done();
                } else {
                    done(res['exec-response']);
                };
            });
        });

        it ('Создать домен', function (done) {
            var _e = {
                'func': 'api domain create',
                'args': {
                    'name': testConfig.domain,
                    'customerId': testConfig.domain
                }
            };
            exec(uuid.v4(), _e, function (res) {
                if (res['exec-complete'] === '+OK') {
                    done();
                } else {
                    done(res['exec-response']);
                };
            });
        });

        it ('Создать пользователя', function (done) {
            var _e = {
                'func': 'api account create',
                'args': {
                    'role': 'admin',
                    'param': ''.concat(testConfig.user.number, ':', testConfig.user.password, '@',
                        testConfig.domain)
                }
            };
            exec(uuid.v4(), _e, function (res) {
                if (res['exec-complete'] === '+OK') {
                    done();
                } else {
                    done(res['exec-response']);
                };
            });
        });

        it ('Удалить домен', function (done) {
            var _e = {
                'func': 'api domain remove',
                'args': {
                    'name': testConfig.domain
                }
            };
            exec(uuid.v4(), _e, function (res) {
                if (res['exec-complete'] === '+OK') {
                    done();
                } else {
                    done(res['exec-response']);
                };
            });
        });

        it('Проверка стека команд', function (done) {
            if (Object.keys(apiQ).length === 0) {
                done()
            } else {
                done(Object.keys(apiQ).length);
            }
        });
    });
});