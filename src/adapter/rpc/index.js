const Router = require('./router');

module.exports = app => {

    const router = new Router(app.Broker);

    require('./account/account.resource').addRoutes(router);
};