/**
 * Created by igor on 07.07.17.
 */

'use strict';

const log = require(__appRoot + '/lib/log')(module);
const escape = c => '"' + c.replace(/-|"|'|,\s|\\|\//g,'') + '"'; //todo

module.exports = {
    buildQuery: (pool, request, table, cb) => {
        const page = parseInt(request['pageNumber'], 10) || 0,
            limit = parseInt(request['limit'], 10) || 40;

        const parameters = [];
        const filters = [];
        let sort = null;

        if (!request.filter)
            request.filter = {};

        for (let key in request.filter) {
            if (typeof request.filter[key] === 'string') {
                filters.push(`${key} like $${parameters.push(request.filter[key] + '%')}`)
            } else {
                filters.push(`${key} = $${parameters.push(request.filter[key])}`)
            }
        }
        if (request.domain) {
            filters.push(`domain = $${parameters.push(request.domain)}`)
        }

        for (let key in request.sort) {
            sort = `${escape(key)} ${request.sort[key] === -1 ? 'ASC' : 'DESC'}`;
            break;
        }

        pool.query(
            `SELECT ${ request.columns && request.columns.length > 0 ? request.columns.map(escape) : '*'} 
             FROM "${table}"
             ${filters.length > 0 ? 'WHERE ' + filters.join(' AND ') : ''}
             ORDER BY ${sort ? sort : 'id'}
             LIMIT $${parameters.push(limit)} OFFSET $${parameters.push(page > 0 ? ((page - 1) * limit) : 0)}
            `,
            parameters,
            (err, res) => {
                if (err) {
                    return cb(err);
                }

                return cb(null, res.rows)
            }
        );
    }
};