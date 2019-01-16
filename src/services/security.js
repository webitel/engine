const conf = require(__appRoot + '/conf'),
    TOTPAuthenticator = require(__appRoot + '/lib/twoFactorAuthenticator'),
    CodeError = require(__appRoot + '/lib/error'),
    ENABLED = `${conf.get('application:auth:useTOTP')}` === 'true',
    NAME = `${conf.get('server:baseUrl')}`;

const Service = module.exports = {
    authenticator: null,
    init(cb) {
        application.PG.getQuery('metadata').item('root', 'security_settings', (err, settings) => {
            if (err && err.status !== 404) {
                cb(err);
                return;
            }

            if (!settings || !settings.data) {
                this.authenticator = new TOTPAuthenticator(NAME, {enabled: false});
            } else {
                this.authenticator = new TOTPAuthenticator(NAME, settings.data);
            }
            cb(null);
        });
    },
    getSettings(caller, options, cb) {
        if (caller.id !== 'root' || caller.domain) {
            return cb(new CodeError(403, `Forbidden`))
        }

        if (this.authenticator.isEnabled() && !options.code) {
            return cb(new CodeError(301, `Code is required`));
        }

        if (this.authenticator.isEnabled() && !this.verifying(options.code)) {
            return cb(new CodeError(403, `Forbidden`))
        }

        cb(null, this.authenticator.toJson())
    },
    setSettings(caller, options, cb) {
        if (caller.id !== 'root' || caller.domain) {
            return cb(new CodeError(403, `Forbidden`))
        }

        if (this.authenticator.isEnabled() && !options.code) {
            return cb(new CodeError(301, `Code is required`));
        }

        if (this.authenticator.isEnabled() && !this.verifying(options.code)) {
            return cb(new CodeError(403, `Forbidden`))
        }

        if (options.generateNewSecretKey) {
            return this.generateNewSecret(cb);
        }
        if (options.hasOwnProperty("setEnable")) {
            return this.setEnabled(!!options.setEnable, cb);
        }
        return cb(new CodeError(400, 'No method'))
    },

    setEnabled(enabled, cb) {
        this.authenticator.setEnabled(enabled);
        this.saveSettings(err => {
            if (err)
                return cb(err);

            cb(null, this.authenticator.toJson());
        });
    },

    generateNewSecret(cb) {
        this.authenticator.setSecret(TOTPAuthenticator.generateNewSecret());
        this.saveSettings(err => {
            if (err)
                return cb(err);

            cb(null, this.authenticator.toJson());
        });
    },

    saveSettings(cb) {
        application.PG.getQuery('metadata').createOrReplace('root', 'security_settings', this.authenticator.getMetadata(), cb);
    },

    isEnabled() {
        return ENABLED && this.authenticator.isEnabled()
    },

    verifying(token) {
        return this.authenticator.verifying(token);
    }
};