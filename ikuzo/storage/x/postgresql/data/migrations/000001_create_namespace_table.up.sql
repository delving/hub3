create table if not exists namespaces (
  id bigserial primary key,
  prefix text not null,
  uri text not null,
  weight integer not null default 1,
  version integer not null default 1
);
