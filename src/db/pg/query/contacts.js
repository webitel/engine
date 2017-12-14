/**
 * Created by I. Navrotskyj on 13.10.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

const sqlContactItem = `
    SELECT * FROM v_contacts_list 
    where v_contacts_list.id = $1 AND v_contacts_list.domain = $2
`;

const sqlContactCreate = `
WITH i as (
  INSERT INTO contacts (domain, name, company_name, job_name, description, photo, custom_data, tags) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING id
), comm as (
  select * from (
    SELECT d->>'number'::varchar(120) as number, (d->>'type_id')::bigint as type_id
    FROM (
      select json_array_elements($9)::JSON as d
    ) s
  ) t
  WHERE not type_id ISNULL and not number ISNULL
), ins_com as (
    INSERT INTO contacts_communication(contact_id, number, type_id)
    SELECT i.id, number, type_id
    FROM comm, i
)
SELECT i.id from i
`;


const sqlContactUpdate = `
WITH con AS (
    UPDATE contacts
      SET name = $1
      , company_name = $2
      , job_name = $3
      , description = $4
      , photo = $5
      , custom_data = $6
      , tags = $7
    WHERE id = $8 AND domain = $9
    RETURNING id        
), comm as (
  select con.id as contact_id, t.* from (
    SELECT (d->>'id')::BIGINT as id, d->>'number'::varchar(120) as number, (d->>'type_id')::bigint as type_id
    FROM (
      select json_array_elements($10)::JSONB as d
    ) s
  ) t, con
  WHERE not type_id ISNULL and not number ISNULL
), del AS (
  DELETE FROM contacts_communication
  WHERE contact_id = (select id from con LIMIT 1) AND NOT id in (SELECT id FROM comm where not id is null)
), upd AS (
  UPDATE contacts_communication
    SET number = comm.number,
      type_id = comm.type_id
  FROM comm
  where comm.contact_id = contacts_communication.contact_id AND comm.id = contacts_communication.id
), ins AS (
  INSERT INTO contacts_communication (number, type_id, contact_id)
  SELECT number, type_id, contact_id from comm
  WHERE comm.id is NULL
)
select row_to_json(t) as contact
from (
  select *,
    (
      select array_to_json(array_agg(row_to_json(d)))
      from (
        select contacts_communication.id, contacts_communication.number, contacts_communication.type_id, ct.name as type_name
        from contacts_communication
        INNER JOIN communication_type as ct on ct.id = contacts_communication.type_id
        where contacts_communication.contact_id = contacts.id
      ) d
    ) as communications
  from contacts
  where contacts.id = (SELECT id FROM con LIMIT 1)
) t;
`;

const sqlContactDelete = `
DELETE 
FROM contacts 
WHERE id = $1 AND domain = $2
RETURNING * 
`;

const sqlContactDeleteByDomain = `
DELETE 
FROM contacts 
WHERE domain = $1
`;

const sqlCommTypeCreate = `
INSERT INTO communication_type (name, domain)
VALUES ($1, $2)
RETURNING *    
`;

const sqlCommTypeRemove = `
DELETE 
FROM communication_type 
WHERE domain = $1 AND id = $2
RETURNING *
`;

const sqlCommTypeUpdate = `
UPDATE communication_type
SET  name = $1
WHERE domain = $2 AND id = $3
RETURNING *
`;

const sqlTestYealink = `
SELECT xmlelement(
    NAME "YealinkIPPhoneDirectory",
    (SELECT xmlagg(f.directory)
      FROM (
        SELECT xmlelement(
          NAME "DirectoryEntry"
          ,xmlagg(xmlelement(name "Name", c.name))
          ,xmlagg((select xmlagg(xmlforest(number as "Telephone")) from contacts_communication WHERE contact_id = c.id))
      ) as directory
      FROM contacts c WHERE c.domain = $1
      GROUP BY c.id, c.name
      ) as f)
)
`;

const sqlTestVCARD = `
SELECT string_agg(concat('BEGIN:VCARD\nVERSION:3.0\nN:', name), '\nEND:VCARD\n\n') as data
FROM contacts;
`;

const sqlDeleteCommunicationTypesByDomain = `
    DELETE 
    FROM communication_type 
    WHERE domain = $1 
`;

function add(pool) {

    return {
        list: (request, cb) => {
            buildQuery(pool, request, "v_contacts_list", cb);
        },

        findById: (id, domainName, cb) => {
            pool.query(
                sqlContactItem,
                [
                    +id,
                    domainName
                ], (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount && res.rows[0]) {
                        if (res.rows[0].photo) {
                            //TODO query encode
                            res.rows[0].photo = res.rows[0].photo.toString();
                        }

                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}@${domainName}`));
                    }
                }
            )
        },

        create: (contact = {}, domain, cb) => {
            try {

                pool.query(
                    //domain, name, company_name, job_name, description, photo, custom_data, tags, communications
                    sqlContactCreate,
                    [
                        domain,
                        contact.name,
                        contact.company_name,
                        contact.job_name,
                        contact.description,
                        contact.photo, //photo
                        contact.custom_data ? JSON.stringify(contact.custom_data) : null,
                        contact.tags,
                        contact.communications ? JSON.stringify(contact.communications) : null
                    ], (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount ) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${domain}`));
                        }
                    }
                )
            } catch (e) {
                return cb(e);
            }
        },

        update: (contact = {}, id, domain, cb) => {
            try {
                pool.query(
                    //name, company_name, job_name, description, photo, custom_data, tags, id, domain, communications
                    sqlContactUpdate,
                    [
                        contact.name,
                        contact.company_name,
                        contact.job_name,
                        contact.description,
                        contact.photo, //photo
                        contact.custom_data ? JSON.stringify(contact.custom_data) : null,
                        contact.tags,
                        +id,
                        domain,
                        contact.communications ? JSON.stringify(contact.communications) : null
                    ], (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount && res.rows[0]) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${domain}`));
                        }
                    }
                )
            } catch (e) {
                return cb(e);
            }
        },

        deleteById: (id, domain, cb) => {
            pool.query(
                sqlContactDelete,
                [
                    +id,
                    domain
                ], (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount && res.rows[0]) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}@${domain}`));
                    }
                }
            )
        },

        deleteByDomain: (domain, cb) => {
            pool.query(
                sqlContactDeleteByDomain,
                [
                    domain
                ], (err, res) => {
                    if (err) {
                        return cb(err);
                    }

                    pool.query(sqlDeleteCommunicationTypesByDomain, [domain], err => {
                        if (err) {
                            log.error(err)
                        }
                    });

                    return cb(null, res.rows)
                }
            )
        },

        types: {
            list: (request, cb) => {
                buildQuery(pool, request, "communication_type", cb);
            },

            create: (name, domain, cb) => {
                pool.query(
                    sqlCommTypeCreate,
                    [
                        name,
                        domain
                    ], (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount && res.rows[0]) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${name}@${domain}`));
                        }
                    }
                )
            },

            delete: (domain, id, cb) => {
                pool.query(
                    sqlCommTypeRemove,
                    [
                        domain,
                        +id
                    ], (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount && res.rows[0]) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${domain}`));
                        }
                    }
                )
            },

            update: (domain, id, name, cb) => {
                pool.query(
                    sqlCommTypeUpdate,
                    [
                        name,
                        domain,
                        +id
                    ], (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount && res.rows[0]) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${domain}`));
                        }
                    }
                )
            }
        },

        importData: {
            yeaLink: (domain, cb) => {
                pool.query(
                    sqlTestYealink,
                    [domain],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0].xmlelement)
                        } else {
                            return cb(new CodeError(404, `Not found ${domain}`));
                        }
                    }

                )
            },

            vCard: (domain, cb) => {
                pool.query(
                    sqlTestVCARD,
                    [],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0].data)
                        } else {
                            return cb(new CodeError(404, `Not found ${'dsa'}`));
                        }
                    }

                )
            }
        }
    };
}

module.exports = add;