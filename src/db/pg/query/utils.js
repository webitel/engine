/**
 * Created by igor on 07.07.17.
 */

'use strict';

const log = require(__appRoot + '/lib/log')(module);
const escape = c => '"' + c.replace(/-|"|'|,\s|\\|\//g,'') + '"'; //todo
const escapeFilter = c => '\'' + c.replace(/-|"|'|,\s|\\|\//g,'') + '\''; //todo
module.exports = {
    buildQuery: (request, table) => {
        const page = parseInt(request['pageNumber'], 10) || 0,
            limit = parseInt(request['limit'], 10) || 40;

        let filters = [];

        if (!request.filter)
            request.filter = {};

        if (request.domain)
            request.filter.domain = request.domain;

        for (let key in request.filter) {
            if (typeof request.filter[key] === 'string') {
                filters.push(`${key}=${escapeFilter(request.filter[key])}`)
            } else {
                filters.push(`${key}=${request.filter[key]}`)
            }
        }

        const r = `SELECT ${ !request.columns ? '*' : Object.keys(request.columns).map(escape) } 
            FROM "${table}" 
            ${filters.length > 0 ? 'WHERE ' + filters.join(' AND ') : ''} 
            LIMIT ${limit} OFFSET ${page > 0 ? ((page - 1) * limit) : 0}`;
        log.trace(`db: ${r}`);
        return r
    }
};

function getFilters(filters) {
    filters.map(n => '')
}