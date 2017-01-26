/**
 * Created by igor on 26.01.17.
 */

"use strict";
    
const fs = require('fs'),
    conf = require(__appRoot + '/conf'),
    filePath = conf.get('application:auth:tokenSecretKey');

const TOKEN_DATA = fs.readFileSync(filePath);

module.exports = TOKEN_DATA;

