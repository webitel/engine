
class Response {
    constructor(id, queue, router) {
        this.id = id;
        this.queue = queue;
        this._router = router; // ?

        this._status = 200;
    }

    status(status) {
        this._status = status;
        return this;
    }

    json(body = {}) {
        this._router.send(this, {
            "exec-args": {
                "callId": this.id,
                "status": this._status,
                "data": body
            }
        });
    }

    text(data) {
        this._router.send(this, {
            "exec-args": {
                "callId": this.id,
                "status": this._status,
                data
            }
        });
    }
}

module.exports = Response;