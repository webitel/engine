const accountService = require(__appRoot + '/services/account');
const rootCaller = require(__appRoot + '/middleware/checkPermissions').ROOT;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(router) {
    router.addRoute(`account.user_data`, userData);
}

function userData(request, response) {
    const options = {
        domain: request.body.domain,
        columns: request.body.columns,
        filter: request.body.filter,
        convertToArray: true
    };
    accountService.accountList(rootCaller, options, (err, res) => {
        if (err) {
            return response.status(500).json(err)
        }
        response.status(200).json(res);
    })
}