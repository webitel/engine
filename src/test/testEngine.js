/**
 * Created by Igor Navrotskyj on 02.09.2015.
 *
 * Update the package index:
 *    sudo apt-get update
 * Install sip-tester deb package:
 *    sudo apt-get install sip-tester
 *
 * sipp pre.webitel.com:5060 -i 10.10.10.135 -s sipp -d 1s -l 15 -aa -mi 10.10.10.135 -rtp_echo -nd -r 10
 */

'use strict';

var assert = require('chai').assert;
var request = require('supertest');
var uuid = require('node-uuid');
request = request('http://localhost:10022');

var conf = require('../conf');
var rootName = 'root';
var rootPassword = conf.get('webitelServer:secret');
var rootCredentials = {};
var domain = '10.10.10.144';


describe("REST API", function () {
    describe("AUTH", function () {
        it ('Root auth', function (done) {
            request
                .post('/login')
                .expect('Content-Type', /json/)
                .send({
                    "username": rootName,
                    "password": rootPassword
                })
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    assert.equal(res.body.username, rootName);
                    assert.ok(res.body.key, 'Bad key response');
                    assert.ok(res.body.token, 'Bad token response');
                    rootCredentials = res.body;
                    done();
                });
        });
    });
    
    describe('Root exec API', function () {
        /** Domain
         *
         */

        describe('Domain', function () {

            var defData = {
                "domain_name": "testEngine2",
                "customer_id": "testEngine2",
                "parameters": ["ASD=1"],
                "variables": ["XCV=2"]
            };

            it('POST [/api/v2/domains]', function (done) {
                request
                    .post('/api/v2/domains')
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(defData)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.ok(result['info'], 'Bad response info attribute.');
                        assert.equal(result['status'], 'OK');
                        done();
                    });
            });

            it('GET [/api/v2/domains]', function (done) {
                request
                    .get('/api/v2/domains?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        var item = results['info'][defData.domain_name];
                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        assert.ok(item, "Bad response domain.");
                        done();
                    });
            });

            it('GET [/api/v2/domains/:name]', function (done) {
                request
                    .get('/api/v2/domains/' + defData.domain_name + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        var item = results['info'];

                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        assert.equal(item['ASD'], "1");
                        assert.equal(item['variable_XCV'], "2");
                        done();
                    });
            });

            it('PUT [/api/v2/domains/:name]', function (done) {
                var updateDef = ['default_lang=ru'];
                request
                    .put('/api/v2/domains/' + defData.domain_name + '/var')
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(updateDef)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var body = res.body;
                        assert.ok(body, 'Bad response.');
                        assert.equal(body['status'], "OK");
                        assert.isObject(body['info'], "Bad response info");
                        done();
                    });
            });

            it('DELETE [/api/v2/domains/:name]', function (done) {
                request
                    .delete('/api/v2/domains/' + defData.domain_name + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        done();
                    });
            });

        });

        /**
         * Accounts
         */
        
        describe('Account', function () {
            var defData = {
                "login": "10030",
                "domain": domain,
                "role": "admin",
                "parameters": ["webitel-extension=10030", "foo=bar"]
            };

            it('POST [/api/v2/accounts]', function (done) {
                request
                    .post('/api/v2/accounts?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(defData)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.ok(result['info'], 'Bad response info attribute.');
                        assert.equal(result['status'], 'OK');
                        done();
                    });
            });

            it('GET [/api/v2/accounts]', function (done) {
                request
                    .get('/api/v2/accounts?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        var item = results['info'];

                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        assert.isObject(item[defData.login], "Bad created user");
                        done();
                    });
            });

            it('GET [/api/v2/accounts/:name]', function (done) {
                request
                    .get('/api/v2/accounts/' + defData.login + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        var item = results['info'];

                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        assert.equal(item['foo'], 'bar');
                        assert.equal(item['variable_account_role'], defData.role);
                        assert.equal(item['variable_w_domain'], defData.domain);
                        done();
                    });
            });

            it('PUT [/api/v2/accounts/:name]', function (done) {
                var updateDef = {
                    "variables": ["test2=3"],
                    "parameters": ["asd=asd", "webitel-extensions=10090"]
                };

                request
                    .put('/api/v2/accounts/' + defData.login + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(updateDef)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var body = res.body;
                        assert.ok(body, 'Bad response.');
                        assert.equal(body['status'], "OK");
                        assert.isObject(body['info'], "Bad response info");
                        done();
                    });
            });

            it('DELETE [/api/v2/accounts/:name]', function (done) {
                request
                    .delete('/api/v2/accounts/' + defData.login + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var results = res.body;
                        assert.equal(results['status'], "OK");
                        assert.ok(results['info'], "Bad response info.");
                        done();
                    });
            });
        });

        /**
         * Dialplan
         */
        describe('Dial plan', function () {

            describe('Public', function () {
                var publicDialplan = {
                    "domain": domain,
                    "destination_number": ["testEngine"],
                    "fs_timezone": "Europe/Andorra",
                    "name": "testEngine",
                    "timezone": "+60",
                    "timezonename": "GMT(+01:00) Europe/Andorra",
                    "callflow": [{
                        "if": {
                            "expression": "1==1",
                            "then": [{
                                "hangup": "USER_BUSY"
                            }]
                        }
                    }]
                };
                var publicId;
                it('POST [/api/v2/routes/public]', function (done) {
                    request
                        .post('/api/v2/routes/public')
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .send(publicDialplan)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var result = res.body;
                            var dialplan = result['data'];
                            assert.ok(result['info'], 'Bad response info attribute.');
                            publicId = result['info'];
                            assert.equal(result['status'], "OK");

                            assert.equal(dialplan['domain'], publicDialplan['domain']);
                            assert.equal(dialplan['version'], 2);
                            assert.equal(dialplan['fs_timezone'], publicDialplan['fs_timezone']);
                            assert.equal(dialplan['timezone'], publicDialplan['timezone']);
                            assert.equal(dialplan['timezonename'], publicDialplan['timezonename']);
                            assert.equal(dialplan['name'], publicDialplan['name']);
                            assert.equal(dialplan['destination_number'][0], publicDialplan['destination_number'][0]);
                            assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                            done();
                        });
                });
                
                it('GET [/api/v2/routes/public]', function (done) {
                    request
                        .get('/api/v2/routes/public?domain=' + domain)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var results = res.body,
                                dialplan;

                            for (var key in results) {
                                dialplan = results[key];
                                if (dialplan['_id'] != publicId) continue;

                                assert.equal(dialplan['domain'], publicDialplan['domain']);
                                assert.equal(dialplan['version'], 2);
                                assert.equal(dialplan['fs_timezone'], publicDialplan['fs_timezone']);
                                assert.equal(dialplan['timezone'], publicDialplan['timezone']);
                                assert.equal(dialplan['timezonename'], publicDialplan['timezonename']);
                                assert.equal(dialplan['name'], publicDialplan['name']);
                                assert.equal(dialplan['destination_number'][0], publicDialplan['destination_number'][0]);
                                assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                                return done();
                            };
                            done(new Error("Not found " + publicId));
                        });
                });

                it('PUT [/api/v2/routes/public/:ID]', function (done) {
                    var updateDef = JSON.parse(JSON.stringify(publicDialplan));
                    updateDef['name'] = 'updatedNameTestEngine';
                    request
                        .put('/api/v2/routes/public/' + publicId)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .send(updateDef)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var dialplan = res.body;
                            assert.ok(dialplan, 'Bad response.');
                            assert.equal(dialplan['domain'], publicDialplan['domain']);
                            assert.equal(dialplan['version'], 2);
                            assert.equal(dialplan['fs_timezone'], publicDialplan['fs_timezone']);
                            assert.equal(dialplan['timezone'], publicDialplan['timezone']);
                            assert.equal(dialplan['timezonename'], publicDialplan['timezonename']);
                            assert.equal(dialplan['name'], publicDialplan['name']);
                            assert.equal(dialplan['destination_number'][0], publicDialplan['destination_number'][0]);
                            assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                            done();
                        });
                });

                it('DELETE [/api/v2/routes/public/:ID]', function (done) {
                    request
                        .delete('/api/v2/routes/public/' + publicId)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var result = res.body;
                            assert.equal(result['status'], "OK");
                            assert.equal(result['info'], 1);
                            done();
                        });
                });
            });

            describe('Default', function () {
                var defaultDialplan = {
                    "domain": domain,
                    "destination_number": "testEngine",
                    "fs_timezone": "Europe/Andorra",
                    "name": "testEngine",
                    "timezone": "+60",
                    "timezonename": "GMT(+01:00) Europe/Andorra",
                    "callflow": [{
                        "if": {
                            "expression": "1==1",
                            "then": [{
                                "hangup": "USER_BUSY"
                            }]
                        }
                    }]
                };
                var defaultId;
                it('POST [/api/v2/routes/default]', function (done) {
                    request
                        .post('/api/v2/routes/default')
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .send(defaultDialplan)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var result = res.body;
                            var dialplan = result['data'];
                            assert.ok(result['info'], 'Bad response info attribute.');
                            defaultId = result['info'];
                            assert.equal(result['status'], "OK");

                            assert.equal(dialplan['domain'], defaultDialplan['domain']);
                            assert.equal(dialplan['version'], 2);
                            assert.equal(dialplan['fs_timezone'], defaultDialplan['fs_timezone']);
                            assert.equal(dialplan['timezone'], defaultDialplan['timezone']);
                            assert.equal(dialplan['timezonename'], defaultDialplan['timezonename']);
                            assert.equal(dialplan['name'], defaultDialplan['name']);
                            assert.equal(dialplan['destination_number'], defaultDialplan['destination_number']);
                            assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                            done();
                        });
                });

                it('GET [/api/v2/routes/default]', function (done) {
                    request
                        .get('/api/v2/routes/default?domain=' + domain)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var results = res.body,
                                dialplan;

                            for (var key in results) {
                                dialplan = results[key];
                                if (dialplan['_id'] != defaultId) continue;

                                assert.equal(dialplan['domain'], defaultDialplan['domain']);
                                assert.equal(dialplan['version'], 2);
                                assert.equal(dialplan['fs_timezone'], defaultDialplan['fs_timezone']);
                                assert.equal(dialplan['timezone'], defaultDialplan['timezone']);
                                assert.equal(dialplan['timezonename'], defaultDialplan['timezonename']);
                                assert.equal(dialplan['name'], defaultDialplan['name']);
                                assert.equal(dialplan['destination_number'], defaultDialplan['destination_number']);
                                assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                                return done();
                            };
                            done(new Error("Not found " + defaultId));
                        });
                });

                it('PUT [/api/v2/routes/default/:ID]', function (done) {
                    var updateDef = JSON.parse(JSON.stringify(defaultDialplan));
                    updateDef['name'] = 'updatedNameTestEngine';
                    request
                        .put('/api/v2/routes/default/' + defaultId)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .send(updateDef)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var dialplan = res.body;
                            assert.equal(dialplan['domain'], defaultDialplan['domain']);
                            assert.equal(dialplan['version'], 2);
                            assert.equal(dialplan['fs_timezone'], defaultDialplan['fs_timezone']);
                            assert.equal(dialplan['timezone'], defaultDialplan['timezone']);
                            assert.equal(dialplan['timezonename'], defaultDialplan['timezonename']);
                            assert.equal(dialplan['name'], defaultDialplan['name']);
                            assert.equal(dialplan['destination_number'], defaultDialplan['destination_number']);
                            assert.isArray(dialplan['callflow'], 'Bad response callflow attribute');
                            done();
                        });
                });

                it('DELETE [/api/v2/routes/default/:ID]', function (done) {
                    request
                        .delete('/api/v2/routes/default/' + defaultId)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var result = res.body;
                            assert.equal(result['status'], "OK");
                            assert.equal(result['info'], 1);
                            done();
                        });
                });
            });

            // TODO Extensions

            describe('Domain variables', function () {
                var defaultVar = {
                    "my_super_var": "test",
                    "my_super_var_2": "test2"
                };

                it('POST [/api/v2/routes/variables]', function (done) {
                    request
                        .post('/api/v2/routes/variables?domain=' + domain)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .send(defaultVar)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var result = res.body;
                            assert.equal(result['status'], "OK");
                            assert.equal(result.data['nModified'], 1);
                            done();
                        });
                });

                it('GET [/api/v2/routes/variables]', function (done) {
                    request
                        .get('/api/v2/routes/variables?domain=' + domain)
                        .expect('Content-Type', /json/)
                        .set('x-key', rootCredentials.key)
                        .set('x-access-token', rootCredentials.token)
                        .expect(200)
                        .end(function (err, res) {
                            if (err) return done(err);
                            var variables = res.body[0].variables;
                            for (var key in defaultVar) {
                                assert.equal(variables[key], defaultVar[key]);
                            };
                            done();
                        });
                });
            });
        });

        /**
         * Call centre.
         */
        
        describe('Call centre.', function () {
            var postQueue = {
                "name": "testEngine",
                "domain": "505050",
                "params": ["strategy=longest-idle-agent", "description='Test Engine'"]
            };
            var putData = {
                "description": 'test\bOK'
            };

            it('GET queues [/api/v2/callcenter/queues]', function (done) {
                request
                    .get('/api/v2/callcenter/queues?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isArray(result['info'], "Bad response info");
                        done();
                    });
            });

            it('POST queues [/api/v2/callcenter/queues]', function (done) {

                request
                    .post('/api/v2/callcenter/queues?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(postQueue)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result['info'], "Bad response info");
                        done();
                    });
            });

            it('GET queues [/api/v2/callcenter/queues/:name]', function (done) {
                request
                    .get('/api/v2/callcenter/queues/' + postQueue.name + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result['info'], "Bad response info");
                        assert.equal(result['info']['description'], "Test Engine");
                        assert.equal(result['info']['strategy'], "longest-idle-agent");
                        done();
                    });
            });

            it('PUT queues [/api/v2/callcenter/queues/:name]', function (done) {
                request
                    .put('/api/v2/callcenter/queues/' + postQueue.name + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(putData)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var response = res.body;
                        assert.ok(response, 'Bad response.');
                        assert.equal(response['status'], "OK");
                        assert.ok(response['info'], "Bad request info");
                        assert.equal(response['info'].description, putData.description);
                        done();
                    });
            });

            it('POST tier [/api/v2/callcenter/queues/:queue/tiers]', function (done) {
                var body = {
                    "agent": "100",
                    "level": 2,
                    "position": 1
                };
                request
                    .post('/api/v2/callcenter/queues/' + postQueue.name + '/tiers?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(body)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result['info'], "Bad response info");
                        done();
                    });
            });

            it('PATH tier lvl [/api/v2/callcenter/queues/:queue/tiers/:agent/level]', function (done) {
                var body = {
                    "level": "0"
                };
                request
                    .put('/api/v2/callcenter/queues/' + postQueue.name + '/tiers/100/level' +  '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(body)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var response = res.body;
                        assert.ok(response, 'Bad response.');
                        assert.equal(response['status'], "OK");
                        assert.ok(response['info'], "Bad request info");
                        done();
                    });
            });

            it('PATH tier position [/api/v2/callcenter/queues/:queue/tiers/:agent/position]', function (done) {
                var body = {
                    "position": "5"
                };
                request
                    .put('/api/v2/callcenter/queues/' + postQueue.name + '/tiers/100/position' +  '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(body)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var response = res.body;
                        assert.ok(response, 'Bad response.');
                        assert.equal(response['status'], "OK");
                        assert.ok(response['info'], "Bad request info");
                        done();
                    });
            });

            it('DELETE tier [/api/v2/callcenter/queues/:queue/tiers/:agent]', function (done) {
                request
                    .delete('/api/v2/callcenter/queues/' + postQueue.name + '/tiers/100' +  '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var response = res.body;
                        assert.ok(response, 'Bad response.');
                        assert.equal(response['status'], "OK");
                        assert.ok(response['info'], "Bad request info");
                        done();
                    });
            });

            it('PATH queues [/api/v2/callcenter/queues/:name/:state]', function (done) {
                request
                    .put('/api/v2/callcenter/queues/' + postQueue.name + '/enable?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var response = res.body;
                        assert.ok(response, 'Bad response.');
                        assert.equal(response['status'], "OK");
                        assert.ok(response['info'], "Bad request info");
                        done();
                    });
            });

            it('GET tiers [/api/v2/callcenter/queues/:queue/tiers]', function (done) {
                // TODO create queue;
                var queue = 'test';
                request
                    .get('/api/v2/callcenter/queues/' + queue +'/tiers?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isArray(result['info'], "Bar response info");
                        done();
                    });
            });

            it('GET members [/api/v2/callcenter/queues/:queue/members]', function (done) {
                // TODO create queue;
                var queue = 'test';
                request
                    .get('/api/v2/callcenter/queues/' + queue +'/members?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isArray(result['info'], "Bar response info");
                        done();
                    });
            });

            it('GET members count [/api/v2/callcenter/queues/:queue/members/count]', function (done) {
                // TODO create queue;
                var queue = 'test';
                request
                    .get('/api/v2/callcenter/queues/' + queue +'/members/count?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isNumber(result['info'], "Bar response info");
                        done();
                    });
            });

            it('DELETE queue [/api/v2/callcenter/queues/:name]', function (done) {
                request
                    .delete('/api/v2/callcenter/queues/' + postQueue.name +'?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result['info'], "Bar response info");
                        done();
                    });
            });

        });

        /**
         * Contact book
         */

        describe('Contact book', function () {
            var createdId;

            it('POST [/api/v2/contacts/]', function (done) {
                var def = {
                    "name": "test",
                    "phones": ["test"],
                    "tags": ["test"]
                };
                request
                    .post('/api/v2/contacts/?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(def)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.data, 'Bad response data.');
                        var data = result.data;
                        assert.equal(data['name'], "test");
                        assert.isArray(data['phones'], "Bad request phones");
                        assert.isArray(data['tags'], "Bad request tags");
                        assert.ok(data['_id'], "Bad request ID");
                        createdId = data['_id'];
                        done();
                    });
            });

            it('GET [/api/v2/contacts/]', function (done) {
                request
                    .get('/api/v2/contacts/?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isArray(result.data, 'Bad response data.');
                        done();
                    });
            });

            it('GET [/api/v2/contacts/:id]', function (done) {
                request
                    .get('/api/v2/contacts/' + createdId + '/?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.data, 'Bad response data.');
                        var data = result.data;
                        assert.equal(data['name'], "test");
                        assert.isArray(data['phones'], "Bad request phones");
                        assert.isArray(data['tags'], "Bad request tags");
                        assert.ok(data['_id'], "Bad request ID");
                        done();
                    });
            });

            it('POST [/api/v2/contacts/searches]', function (done) {
                var query = {
                    "limit": 1,
                    "filter": {
                        "name": "test"
                    }
                };
                request
                    .post('/api/v2/contacts/searches?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(query)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.data, 'Bad response data.');
                        var data = result.data[0];
                        assert.equal(data['name'], "test");
                        assert.isArray(data['phones'], "Bad request phones");
                        assert.isArray(data['tags'], "Bad request tags");
                        assert.ok(data['_id'], "Bad request ID");
                        done();
                    });
            });

            it('PUT [/api/v2/contacts/:id]', function (done) {
                var data = {
                    "name": "asd",
                    "phones": ["11"]
                };
                request
                    .put('/api/v2/contacts/' + createdId + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(data)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.data, 'Bad response data.');
                        var data = result.data;
                        assert.equal(data['name'], "test");
                        assert.isArray(data['phones'], "Bad request phones");
                        assert.isArray(data['tags'], "Bad request tags");
                        assert.ok(data['_id'], "Bad request ID");
                        done();
                    });
            });

            it('DELETE [/api/v2/contacts/:id]', function (done) {
                request
                    .delete('/api/v2/contacts/' + createdId + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.equal(result['info'], 1);
                        done();
                    });
            });
        });

        /**
         * Gateway api
         */
        describe('Gateway', function () {
            var def = {
                "name": "test",
                "username": "username",
                "var": ['test=1']
            };

            it('POST [/api/v2/gateway]', function (done) {

                request
                    .post('/api/v2/gateway?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(def)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        done();
                    });
            });

            it('GET [/api/v2/gateway]', function (done) {

                request
                    .get('/api/v2/gateway?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.isArray(result.info, 'Bad response data.');
                        var item;
                        for (var key in result.info) {
                            if (result.info[key].Gateway == def.name)
                                item = result.info[key];
                        };

                        assert.ok(item, "Bad response created gateway");
                        done();
                    });
            });

            it('GET [/api/v2/gateway/:name]', function (done) {

                request
                    .get('/api/v2/gateway/' + def.name + '?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        assert.equal(result.info['username'], def.username);
                        done();
                    });
            });

            it('PUT [/api/v2/gateway/:name/up]', function (done) {

                request
                    .put('/api/v2/gateway/' + def.name + '/up?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        done();
                    });
            });

            it('PUT [/api/v2/gateway/:name/down]', function (done) {

                request
                    .put('/api/v2/gateway/' + def.name + '/down?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        done();
                    });
            });

            it('PUT [/api/v2/gateway/:name/var]', function (done) {
                var updateParam = [{
                    "myVar": "testEngine"
                }];
                request
                    .put('/api/v2/gateway/' + def.name + '/var?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .send(updateParam)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        assert.equal(result.info['myVar'], updateParam.myVar);
                        done();
                    });
            });

            it('DELETE [/api/v2/gateway/:name]', function (done) {

                request
                    .delete('/api/v2/gateway/' + def.name +'?domain=' + domain)
                    .expect('Content-Type', /json/)
                    .set('x-key', rootCredentials.key)
                    .set('x-access-token', rootCredentials.token)
                    .expect(200)
                    .end(function (err, res) {
                        if (err) return done(err);
                        var result = res.body;
                        assert.equal(result['status'], "OK");
                        assert.ok(result.info, 'Bad response data.');
                        done();
                    });
            });
        });

    });

    describe("ERROR", function() {
        describe("Auth", function () {
            it("POST: Bad login.", function(done) {
                request
                    .post('/login')
                    .expect('Content-Type', /json/)
                    .send({
                        "username": "roott",
                        "password": "roott"
                    })
                    .expect(401, done);
            });

            it("Get: bad token.", function(done) {
                request
                    .get('/api/v2/status')
                    .set('x-key', 'asda')
                    .set('x-access-token', 'ddddddd')
                    .expect(500)
                    .end(function(err, res){
                        if (err) return done(err);
                        assert.equal(res.status, 500);
                        assert.equal(res.body.message, 'Oops something went wrong');
                        done()
                    });
            });

            it("Get: non token & key.", function(done) {
                request
                    .get('/api/v2/status')
                    .expect(401)
                    .end(function(err, res){
                        if (err) return done(err);
                        assert.equal(res.status, 401);
                        assert.equal(res.body.message, 'Invalid Token or Key');
                        done();
                    })
            });
        });
    });

    after(function(done) {
        request
            .post('/logout')
            .expect('Content-Type', /json/)
            .set('x-key', rootCredentials.key)
            .set('x-access-token', rootCredentials.token)
            .expect(200)
            .end(function (err, res) {
                if (err) return done(err);
                assert.equal(res.body.status, 'OK');
                assert.equal(res.body.info, 'Successful logout.');
                done();
            });
    });
});

describe('WSS', function () {
    var WebSocket = require('ws'),
        ws,
        apiQ = {}
        ;
    var wsServer = 'ws://10.10.10.25:10022';
    var ROOT_PASSWORD = 'ROOT_PASSWORD';
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

    it('Connect', function (done) {
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

    it ('Domain list', function (done) {
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

    it ('Create domain', function (done) {
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

    it ('Create user', function (done) {
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

    it ('Delete domain', function (done) {
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

    it('Check commends stack', function (done) {
        if (Object.keys(apiQ).length === 0) {
            done()
        } else {
            done(Object.keys(apiQ).length);
        }
    });
});