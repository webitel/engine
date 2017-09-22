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

   inout member jsonb,

   OUT destination_number VARCHAR(50),
   OUT queue_name VARCHAR(50),
   OUT call_timeout SMALLINT,
   OUT error SMALLINT
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
      where widget.id = widget_id AND (widget.blacklist is null or not cast(member->>'request_ip' as CIDR) <<= ANY(widget.blacklist)) LIMIT 1;

      if queue_id is NULL THEN
        error := -1;
        return ;
      END IF;

      if exists(
        SELECT count(id) as c
        FROM callback_members m
        WHERE (m.request_ip = member->>'request_ip'::varchar(50) ) AND m.created_on BETWEEN ((now() at time zone 'utc')::TIMESTAMP - INTERVAL '30 min')::TIMESTAMP AND (now() at time zone 'utc')
        HAVING count(*) >= limit_by_ip
      ) THEN
        error := -2;
        return;
      END IF;

      if limit_by_number = true AND exists(
        SELECT count(*) as c
        FROM callback_members m
        WHERE (m.number = cast(member->>'number' as varchar(50)) ) AND m.created_on BETWEEN ((now() at time zone 'utc')::TIMESTAMP - INTERVAL '30 min')::TIMESTAMP AND (now() at time zone 'utc')
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
          done,
          done_at,
          done_by,
          logs)
        VALUES (
          member->>'number'::VARCHAR(50),
          cast(member->>'href' as VARCHAR(255)),
          cast(member->>'user_agent' as VARCHAR(300)),
          member->'location',
          cast(member->>'domain' as VARCHAR(70)),
          queue_id,
          cast(member->>'callback_time' as bigint),
          widget_id,
          cast(member->>'request_ip' as VARCHAR(50)),
          cast(member->>'done' as BOOLEAN),
          cast(member->>'done_at' as BIGINT),
          cast(member->>'done_by' as VARCHAR(100)),
          member->'logs'
        )  RETURNING *
      )
      SELECT row_to_json(rows) into member from rows;
      return;
    END;
    $$ LANGUAGE plpgsql;
`);


sql.push(`
create table  IF NOT EXISTS callflow_default
(
	id serial not null
		constraint callflow_default_pkey
			primary key,
	destination_number varchar(120) not null,
	name varchar(120) not null,
	"order" integer default currval('callflow_default_id_seq'::regclass) not null,
	disabled boolean default false,
	debug boolean default false,
	domain varchar(75) not null,
	fs_timezone varchar(35),
	callflow json,
	callflow_on_disconnect json,
	cf_diagram json,
	version smallint default 2,
	call_count bigint default 0,
	description varchar(500)
)
;

create unique index  IF NOT EXISTS  callflow_default_id_uindex
	on callflow_default (id)
;

create index  IF NOT EXISTS  callflow_default_destination_number_disabled_domain_index
	on callflow_default (destination_number, disabled, domain)
;

create index  IF NOT EXISTS  callflow_default_order_index
	on callflow_default ("order")
;
`);

sql.push(`
create table IF NOT EXISTS callflow_extension
(
	id serial not null
		constraint callflow_extension_pkey
			primary key,
	destination_number varchar(50) not null,
	domain varchar(75) not null,
	user_id varchar(120) not null,
	name varchar(75),
	version smallint default 2,
	callflow json,
	callflow_on_disconnect json,
	cf_diagram json,
	fs_timezone varchar(35)
)
;

create unique index IF NOT EXISTS callflow_extension_id_uindex
	on callflow_extension (id)
;

create index IF NOT EXISTS callflow_extension_destination_number_domain_index
	on callflow_extension (destination_number, domain)
;
`);

sql.push(`
create table IF NOT EXISTS callflow_public
(
	id serial not null
		constraint callflow_public_pkey
			primary key,
	destination_number varchar(120) [],
	name varchar(120) not null,
	disabled boolean default false,
	domain varchar(75) not null,
	fs_timezone varchar(35),
	callflow json,
	callflow_on_disconnect json,
	cf_diagram json,
	call_count bigint default 0,
	version smallint default 2,
	description varchar(500),
	debug boolean default false
)
;

create unique index IF NOT EXISTS callflow_public_id_uindex
	on callflow_public (id)
;

DROP INDEX IF EXISTS callflow_public_destination_number_index_unique;

create or REPLACE function callflow_public_check_duplicate_destination() returns trigger as $$
declare
  domain_b VARCHAR(75);
begin
    SELECT domain
    INTO domain_b
    FROM callflow_public WHERE destination_number @> NEW.destination_number AND disabled != true AND id != NEW.id
    LIMIT  1;

    if not domain_b is NULL THEN
      RAISE 'Duplicate destination number: % at domain: %', NEW.destination_number, domain_b  USING ERRCODE = '23505';
    END IF;

    return new;
end
$$ language plpgsql;

DROP TRIGGER  IF EXISTS callflow_public_check_destination_tg ON callflow_public;

CREATE TRIGGER callflow_public_check_destination_tg
BEFORE INSERT OR UPDATE ON callflow_public
    FOR EACH ROW EXECUTE PROCEDURE callflow_public_check_duplicate_destination();

create index IF NOT EXISTS callflow_public_destination_number_index
	on callflow_public USING gin (destination_number)
;

create index IF NOT EXISTS callflow_public_disabled_index
	on callflow_public (disabled)
;
`);


sql.push(`
create table IF NOT EXISTS  callflow_variables
(
	id serial not null
		constraint callflow_variables_pkey
			primary key,
	domain varchar(75) not null,
	variables jsonb
)
;

create unique index IF NOT EXISTS  callflow_variables_id_uindex
	on callflow_variables (id)
;

create unique index IF NOT EXISTS  callflow_variables_domain_uindex
	on callflow_variables (domain)
;
`);

module.exports = sql;