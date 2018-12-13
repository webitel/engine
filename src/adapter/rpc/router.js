const Request = require('./request');
const Response = require('./response');
const log = require(__appRoot + '/lib/log')(module);

class Router {
    constructor(broker) {
        this.routes = new Map();
        this.broker = broker;

        this.notFound = (request, response) => {
            response.status(404).text(`Not found api ${request.api}`)
        }; //TODO

        this.badRequest = (request, response) => {
            response.status(400).text(`Bad request api ${request.api}`)
        }; //TODO

        this.broker.on('rpc_command', this.onData.bind(this));
    }

    onData(data) {
        const request = new Request(data);
        const middleware = this.getMiddleware(request.api);
        try {
            log.debug(`execute [${request.getId()}] ${request.api} response to queue ${request.getQueue()}`);
            middleware(request, new Response(request.getId(), request.getQueue(), this))
        } catch (e) {
            log.error(e);
            //TODO internal error response;
        }
    }

    send(response, body) {
        this.broker.publishToQueue(response.queue, body, {correlationId: response.id})
    }

    addRoute(id, middleware) {
        if (this.routes.has(id)) {
            throw `Route ${id} already exists.`
        }
        this.routes.set(id, middleware);
    }

    setNotFound(middleware) {
        if (typeof middleware !== 'function') {
            throw `Not found middleware must to be function`
        }
        this.notFound = middleware
    }

    getMiddleware(id) {
        if (this.routes.has(id))
            return this.routes.get(id);

        return this.notFound;
    }
}

module.exports = Router;