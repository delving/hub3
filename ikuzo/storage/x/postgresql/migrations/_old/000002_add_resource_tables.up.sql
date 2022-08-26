create table metadata_schema (
    schema_id serial not null primary key,
    uri text unique not null,
    json jsonb not null
);

create table predicate (
    predicate_id text not null primary key,  -- hashed uri
    uri text not null unique,
    search_label text not null,
    schema_id int references metadata_schema(schema_id)
);

create table datatype (
    datatype_id text not null primary key, -- hashed uri
    uri text not null unique,
    search_label text not null,
    comment text not null default 'no comment'
);

create table triple_object (
    object_id text not null primary key, -- hashed predicate uri and object (ntriples)
    object text not null,
    isResource bool default false,
    lang varchar(12) not null default '',
    datatype_id text references datatype(datatype_id),
    predicate_id text references predicate(predicate_id)
);

create table organization (
    org_id text not null primary key,
    domains text [] not null,
    rdf_base_url text not null,
    alt_base_url text []
);

create table dataset (
    dataset_id text not null primary key,
    description text,
    org_id text references organization(org_id)
);

create table resource (
    resource_id text not null primary key, -- hashed uri
    uri text not null unique,
    version_id text not null, -- hash of predicate array
    predicates text [] not null, -- array of triple_object object_ids
    modified date not null default now(),
    dataset_id text references dataset(dataset_id)
);



create type operation_t as enum ('insert', 'update', 'delete');

create table resource_audit (
    audit_ts timestamptz not null default now(),
    operation operation_t not null,
    dataset_id text references dataset(dataset_id),
    username text not null default "current_user"(),
    before jsonb,
    after jsonb
);

create or replace function resource_audit_trig()
 returns trigger
 language plpgsql
as $$
begin
    if tg_op = 'insert'
    then
        insert into resource_audit (operation, after, dataset_id)
        values (tg_op, to_jsonb(new), new.dataset_id);
        return new;

    elsif tg_op = 'update'
    then
        if new != old then
            insert into resource_audit (operation, before, after, dataset_id)
            values (tg_op, to_jsonb(old), to_jsonb(new), new.dataset_id);
        end if;
        return new;

    elsif tg_op = 'delete'
    then
        insert into audit.users_audit (operation, before, dataset_id)
        values (tg_op, to_jsonb(old), old.dataset_id);
        return old;
    end if;
end;
$$;

create trigger resource_audit_trig
    before insert or update or delete
        on resource
            for each row
 execute procedure resource_audit_trig();
