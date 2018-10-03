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
  on callback_members (id);

create index IF NOT EXISTS callback_members_created_on_index
  on callback_members (created_on);

create index IF NOT EXISTS callback_members_callback_time_domain_done_number_index
  on callback_members (callback_time, domain, done, number);

create index IF NOT EXISTS callback_members_callback_time_index
  on callback_members (callback_time);

create index IF NOT EXISTS callback_members_callback_time_domain_done_index
  on callback_members (callback_time, domain, done);
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

CREATE OR REPLACE FUNCTION callflow_default_check_destination_number(reg_txt varchar(120))
  RETURNS BOOLEAN AS
$BODY$
BEGIN
  PERFORM regexp_matches('', reg_txt);
  RETURN reg_txt != '';
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;
  
DO $$
BEGIN

  BEGIN
    ALTER TABLE callflow_default ADD CONSTRAINT callflow_default_check_destination_number_cs CHECK (callflow_default_check_destination_number(destination_number) = TRUE);
  EXCEPTION
    WHEN duplicate_object THEN NULL;
  END;

END $$;

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

create unique index IF NOT EXISTS callflow_extension_destination_number_domain_index
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
    FROM callflow_public WHERE destination_number && NEW.destination_number AND disabled IS NOT TRUE AND id != NEW.id
    LIMIT  1;

    if not domain_b is NULL AND NEW.disabled IS NOT TRUE THEN
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

sql.push(`
create table IF NOT EXISTS tcp_dump
(
	id serial not null
		constraint tcp_dump_pkey
			primary key,
	created_on bigint default (date_part('epoch'::text, now()))::bigint not null,
	filter varchar(250),
	duration smallint default 0 not null,
	description varchar(250),
	meta_file json
);

create unique index IF NOT EXISTS tcp_dump_id_uindex
	on tcp_dump (id)
;`);

//Dialer agents
sql.push(`
create table IF NOT EXISTS agents
(
	name varchar(255) not null
		constraint agents_name_pk
			primary key,
	system varchar(255),
	uuid varchar(255),
	type varchar(255),
	contact varchar(1024),
	status varchar(255),
	state varchar(255),
	max_no_answer integer default 0 not null,
	wrap_up_time integer default 0 not null,
	reject_delay_time integer default 0 not null,
	busy_delay_time integer default 0 not null,
	no_answer_delay_time integer default 0 not null,
	last_bridge_start integer default 0 not null,
	last_bridge_end integer default 0 not null,
	last_offered_call integer default 0 not null,
	last_status_change integer default 0 not null,
	no_answer_count integer default 0 not null,
	calls_answered integer default 0 not null,
	talk_time integer default 0 not null,
	ready_time integer default 0 not null,
	external_calls_count integer default 0 not null
)
;


ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_set_stats timestamp default now();
ALTER TABLE agents ADD COLUMN IF NOT EXISTS logged_in integer default 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS logged_out integer default 0;

create unique index IF NOT EXISTS agents_name_uindex
	on agents (name)
;

create or REPLACE function setagent() returns trigger
	language plpgsql
as $$
BEGIN
  if new.status != old.status OR new.state != old.state THEN

    NEW.last_set_stats = NOW() at time zone 'utc';

     IF OLD.state = 'Waiting' AND (OLD.status = 'Available' OR OLD.status = 'Available (On Demand)') AND NEW.ready_time < EXTRACT(EPOCH FROM NEW.last_set_stats) THEN
       UPDATE agent_in_dialer
         SET idle_sec = idle_sec +
                        (EXTRACT(EPOCH FROM (NEW.last_set_stats -
                         GREATEST(active, to_timestamp(GREATEST(NEW.last_bridge_end + OLD.wrap_up_time, NEW.ready_time, OLD.last_status_change)) at time zone 'utc'))))
       WHERE agent_name = new.name AND NOT active is NULL;
     ELSEIF OLD.status = 'On Break' AND NEW.status != 'On Break' THEN
       UPDATE agent_in_dialer
         SET on_break_sec = on_break_sec + (EXTRACT(EPOCH FROM (NEW.last_set_stats - GREATEST(active, to_timestamp(OLD.last_status_change) at time zone 'utc' ))))
       WHERE agent_name = new.name AND NOT active is NULL ;
     END IF;

    IF OLD.status = 'Logged Out' AND NEW.status != 'Logged Out' THEN
      SELECT EXTRACT(EPOCH FROM NOW())::INT INTO NEW.logged_in;
      NEW.logged_out = 0;
    ELSEIF NEW.status = 'Logged Out' AND OLD.status != 'Logged Out' THEN
      SELECT EXTRACT(EPOCH FROM NOW())::INT INTO NEW.logged_out;
      NEW.logged_in = 0;
    END IF;

  END IF;
  RETURN NEW;
END;
$$
;
DROP TRIGGER  IF EXISTS agents_setagent ON agents;

create trigger agents_setagent
	before update
	on agents
	for each row
	execute procedure setagent()
;

`);
//Dialer agents in dialer
sql.push(`
create table IF NOT EXISTS agent_in_dialer
(
    id serial not null
		constraint agent_in_dialer_id_pk
			primary key,
	agent_name varchar(100) not null,
	dialer_id varchar(24) not null,
	call_count integer default 0,
	missed_call integer default 0,
	process varchar(255),
	call_time_sec integer default 0,
	connected_time_sec integer default 0,
	idle_sec integer default 0,
	on_break_sec integer default 0,
	wrap_time_sec integer default 0,
	active timestamp,
	last_status varchar(120),
	bridged_count integer default 0,
	last_offered_call integer
)
;

create unique index IF NOT EXISTS agent_in_dialer_agent_name_dialer_id_uindex
	on agent_in_dialer (agent_name, dialer_id)
;

create unique index IF NOT EXISTS agent_in_dialer_agent_name_active_uindex
	on agent_in_dialer (agent_name, active)
;

create unique index IF NOT EXISTS agent_in_dialer_id_uindex
	on agent_in_dialer (id)
;


`);

sql.push(`
create table IF NOT EXISTS metadata
(
	id serial not null
		constraint metadata_pkey
			primary key,
	domain varchar(120) not null,
	object_name varchar(50) not null,
	data jsonb
)
;

create unique index IF NOT EXISTS metadata_id_uindex
	on metadata (id)
;

create unique index IF NOT EXISTS metadata_domain_object_name_uindex
	on metadata (domain, object_name)
;
`);

sql.push(`
create table IF NOT EXISTS contacts
(
	id bigserial not null
		constraint contacts_pkey
			primary key,
	domain varchar(100) not null,
	name varchar(120) not null,
	company_name varchar(120),
	job_name varchar(120),
	description varchar(500),
	photo bytea,
	custom_data jsonb,
	tags varchar(50) []
)
;

create unique index IF NOT EXISTS contacts_id_uindex
	on contacts (id)
;

create index IF NOT EXISTS contacts_domain_index
	on contacts (domain)
;

create index IF NOT EXISTS contacts_name_index
	on contacts (name)
;

create index IF NOT EXISTS contacts_company_name_index
	on contacts (company_name)
;

`);

sql.push(`
create table IF NOT EXISTS communication_type
(
	id serial not null
		constraint communication_type_pkey
			primary key,
	name varchar(50) not null,
	domain varchar(120) not null
)
;

create unique index IF NOT EXISTS communication_type_id_uindex
	on communication_type (id)
;

create unique index IF NOT EXISTS communication_type_domain_name_uindex
	on communication_type (domain, name)
;

create index IF NOT EXISTS communication_type_name_index
	on communication_type (name)
;

create index IF NOT EXISTS communication_type_domain_index
	on communication_type (domain)
;
`);

sql.push(`
create table IF NOT EXISTS contacts_communication
(
	id bigserial not null
		constraint contacts_communication_pkey
			primary key,
	contact_id bigint not null
		constraint contacts_communication_contacts_id_fk
			references contacts
				on update cascade on delete cascade,
	number varchar(50) not null,
	digits varchar(50),
	type_id bigint not null
		constraint contacts_communication_communication_type_id_fk
			references communication_type
)
;

create unique index IF NOT EXISTS contacts_communication_id_uindex
	on contacts_communication (id)
;

create index IF NOT EXISTS contacts_communication_contact_id_index
	on contacts_communication (contact_id)
;

create index IF NOT EXISTS contacts_communication_number_index
	on contacts_communication (number)
;

create index IF NOT EXISTS contacts_communication_type_id_index
	on contacts_communication (type_id)
;

CREATE OR REPLACE VIEW v_contacts_list AS
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
      from contacts;
`);

sql.push(`
create table IF NOT EXISTS dialer_templates
(
	id serial not null
		constraint dialer_templates_pkey
			primary key,
	dialer_id varchar(100) not null,
	name varchar(100) not null,
	type varchar(20) not null,
	template jsonb,
	description varchar(500),
	action varchar(20) not null,
	last_response_text text,
	before_delete boolean default false,
	process_state varchar(20),
	process_start integer,
	process_id varchar(30),
	cron varchar(50),
	next_process_id integer,
	success_data bytea,
	execute_time integer
)
;

create unique index IF NOT EXISTS dialer_templates_id_uindex
	on dialer_templates (id)
;
`);

sql.push(`

create table IF NOT EXISTS callflow_private
(
	uuid varchar(50) not null
		constraint callflow_temp_uuid_pk
			primary key,
	domain varchar(75) not null,
  created_on integer default (date_part('epoch'::text, timezone('utc'::text, (now())::timestamp without time zone)))::integer,
	deadline integer default 60,
	fs_timezone varchar(35),
	callflow json
)
;

create unique index IF NOT EXISTS callflow_private_uuid_domain_uindex
	on callflow_private (uuid, domain)
;

create unique index IF NOT EXISTS callflow_private_uuid_uindex
	on callflow_private (uuid)
;

create index IF NOT EXISTS callflow_private_created_on_deadline_index
	on callflow_private (created_on, deadline)
;
`);

sql.push(`
create table IF NOT EXISTS user_stats
(
	id varchar(70) not null
		constraint user_stats_pkey
			primary key,
	updated_at bigint,
	status varchar(50),
	state varchar(50),
	description varchar(50),
	cc boolean,
	ws boolean
)
;

create unique index IF NOT EXISTS user_stats_id_uindex
	on user_stats (id)
;


`);

module.exports = sql;