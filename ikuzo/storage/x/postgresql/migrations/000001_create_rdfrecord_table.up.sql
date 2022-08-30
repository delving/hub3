CREATE TABLE IF NOT EXISTS rdf_records (
    org_id text NOT NULL,
    dataset_id text NOT NULL,
    hub_id text PRIMARY KEY, 
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    deleted boolean default false,
    task_id text NOT NULL,
    content_hash text NOT NULL,
    mime_type text NOT NULL,
    graph bytea NOT NULL,
    version integer NOT NULL DEFAULT 1
);

