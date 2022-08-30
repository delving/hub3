create extension if not exists temporal_tables;

create extension if not exists "uuid-ossp";

create table namespace (
    uuid uuid not null default uuid_generate_v1(),
    prefix text not null,
    uri text not null,
    temporary boolean default false,
    rank integer not null default 1,
    sys_period tstzrange not null,
    primary key (prefix, uri)
);

create table namespace_history (like namespace);

create trigger namespace_hist_trigger 
    before insert or update or delete 
        on namespace 
            for each row
    execute procedure versioning('sys_period', 'namespace_history', true);
