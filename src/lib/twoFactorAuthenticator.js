const speakeasy = require('speakeasy');

module.exports = class TOTPAuthenticator {
    constructor(name = '', {secret = null, qrData = null, enabled = false}) {
        this.secret_ = secret;
        this.enabled_ = enabled;
        this.name_ = encodeURIComponent(name || "Webitel Auth");
    }

    static generateNewSecret() {
        return speakeasy.generateSecret().base32;
    }

    setSecret(secret) {
        this.secret_ = secret;
    }
    setEnabled(enable) {
        this.enabled_ = enable;
        if (this.enabled_ && !this.secret_) {
            this.secret_ = TOTPAuthenticator.generateNewSecret();
        }
    }

    toJson() {
        return {
            qr_data: `otpauth://totp/${this.name_}?secret=${this.secret_}`,
            name: this.name_,
            ...this.getMetadata()
        }
    }

    getMetadata() {
        return {
            secret: this.secret_,
            enabled: this.enabled_
        }
    }

    isEnabled() {
        return this.enabled_;
    }

    verifying(token = "") {
        if (!this.enabled_) {
            return true;
        }
        return speakeasy.time.verify({secret: this.secret_, encoding: 'base32', token})
    }
};