const sql =  [];
// QUEUE
sql.push(`
create table IF NOT EXISTS callback_queue
(
	id serial not null
		constraint callback_queue_id_pk
			primary key,
	name varchar(50) not null,
	description varchar(250),
	domain varchar(70) not null
);
`);

//WIDGET
sql.push(`
create table IF NOT EXISTS widget
(
	id serial not null
		constraint widget_id_pk
			primary key,
	name varchar(50) not null,
	description varchar(200),
	config json,
	domain varchar(255) not null,
	limit_by_number boolean,
	limit_by_ip smallint,
	"_file_path" varchar(250),
	blacklist cidr[],
	queue_id integer not null
		constraint widget_callback_queue_id_fk
			references callback_queue,
	language varchar(4),
	callflow_id varchar(24)
);

create index IF NOT EXISTS widget_id_index
	on widget (id)
;
`);

//MEMBERS
sql.push(`
create table IF NOT EXISTS callback_members
(
	id serial not null
		constraint callback_members_pkey
			primary key,
	number varchar(50) not null,
	href varchar(255),
	user_agent varchar(300),
	location jsonb,
	domain varchar(70) not null,
	queue_id integer not null
		constraint callback_members_callback_queue_id_fk
			references callback_queue
				on update cascade on delete cascade,
	done boolean,
	done_by varchar(100),
	callback_time bigint,
	widget_id integer
		constraint callback_members_widget_id_fk
			references widget,
	done_at bigint,
	created_on timestamp default (now())::timestamp without time zone,
	request_ip varchar(50),
	logs jsonb
)
;

create unique index IF NOT EXISTS callback_members_id_uindex
	on callback_members (id)
;
`);

sql.push(`
create table IF NOT EXISTS callback_members_comment
(
	id serial not null
		constraint callback_members_comment_pkey
			primary key,
	created_by varchar(100) not null,
	text text not null,
	created_on bigint not null,
	member_id integer not null
		constraint callback_members_comment_callback_members_id_fk
			references callback_members
				on update cascade on delete cascade
)
;

create unique index IF NOT EXISTS callback_members_comment_id_uindex
	on callback_members_comment (id)
;
`);

sql.push(`
CREATE OR REPLACE FUNCTION insert_member_public(
   in widget_id BIGINT,
   in memberNumber VARCHAR(50),
   in href VARCHAR(255),
   in user_agent VARCHAR(300),
   in location jsonb,
   in domain VARCHAR(70),
   in callback_time BIGINT,
   in _request_ip VARCHAR(50),
   in memberLogs jsonb,
   OUT destination_number VARCHAR(50),
   OUT queue_name VARCHAR(50),
   OUT call_timeout SMALLINT,
   OUT error SMALLINT,
   OUT member_json jsonb
 )
    AS $$
    DECLARE
      queue_id BIGINT;
      limit_by_number BOOLEAN;
      limit_by_ip SMALLINT;
    BEGIN
      SELECT
        widget.queue_id,
        widget.limit_by_number,
        widget.limit_by_ip,
        json_extract_path_text(widget.config,'destinationNumber')::VARCHAR(50),
        c.name,
        json_extract_path_text(widget.config,'hookCountDown')::SMALLINT
      INTO queue_id, limit_by_number, limit_by_ip, destination_number, queue_name, call_timeout
      FROM widget
        INNER JOIN callback_queue c on widget.queue_id = c.id
      where widget.id = widget_id AND (widget.blacklist is null or not _request_ip::cidr <<= ANY(widget.blacklist) ) LIMIT 1;

      if queue_id is NULL THEN
        error := -1;
        return ;
      END IF;

      if exists(
        SELECT count(id) as c
        FROM callback_members m
        WHERE (m.request_ip = _request_ip ) AND m.created_on BETWEEN ((now() at time zone 'utc')::TIMESTAMP - INTERVAL '30 min')::TIMESTAMP AND (now() at time zone 'utc')
        HAVING count(*) >= limit_by_ip
      ) THEN
        error := -2;
        return;
      END IF;

      if limit_by_number = true AND exists(
        SELECT count(*) as c
        FROM callback_members m
        WHERE (m.number = memberNumber ) AND m.created_on BETWEEN ((now() at time zone 'utc')::TIMESTAMP - INTERVAL '30 min')::TIMESTAMP AND (now() at time zone 'utc')
        HAVING count(*) >= 1
      ) THEN
        error := -3;
        return;
      END IF;


      with rows as (
        INSERT INTO callback_members (
          number,
          href,
          user_agent,
          location,
          domain,
          queue_id,
          callback_time,
          widget_id,
          request_ip,
          logs)
        VALUES (
          memberNumber,
          href,
          user_agent,
          location,
          domain,
          queue_id,
          callback_time,
          widget_id,
          _request_ip,
          memberLogs
        )  RETURNING *
      )
      SELECT row_to_json(rows) into member_json from rows;


      return;
    END;
    $$ LANGUAGE plpgsql;
`);

module.exports = sql;