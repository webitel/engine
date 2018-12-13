
class Request {
    constructor(msg = {properties: {}, fields: {}}) {
        this.properties = msg.properties;
        this.encodings = msg.properties.contentEncoding || 'utf8';

        this.api = null;
        this.body = {};

        this.exchange = (this.properties.headers && this.properties.headers['x-api-resp-exchange']) || msg.fields.exchange;
        this.routingKey = this.properties.headers && this.properties.headers['x-api-resp-key'];

        const data = getJson(msg.content.toString(this.encodings));
        if (data) {
            this.api = data['exec-api'];
            this.body = data['exec-args'] || {};
        }
    }

    getId() {
        return this.properties.correlationId || this.routingKey;
    }

    getQueue() {
        return this.properties.replyTo;
    }
}

module.exports = Request;

function getJson(data = "") {
    try {
        return JSON.parse(data)
    } catch (e) {
        return {};
    }
}