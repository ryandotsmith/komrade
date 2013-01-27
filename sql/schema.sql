begin;

create table queues (
	token uuid default uuid_generate_v4(),
	heroku_id text,
	plan text,
	callback_url text,
	failed_count int default 0
);

create table jobs (
	queue uuid,
	id uuid,
	payload json,
	created_at timestamptz default now(),
	locked_at timestamptz,
	failed_count int default 0
);

create index jobs_by_queue on jobs (queue);

create table failed_jobs (
	job_id uuid,
	queue uuid,
	id uuid,
	payload json,
	created_at timestamptz default now()
);

commit;
