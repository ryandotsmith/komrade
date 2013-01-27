begin;

create table queues (
	token uuid default uuid_generate_v4(),
	heroku_id text,
	plan text,
	callback_url text,
	in_minute bigint default 0,
	in_hour bigint default 0,
	in_day bigint default 0,
	out_minute bigint default 0,
	out_hour bigint default 0,
	out_day bigint default 0,
	error_minute bigint default 0,
	error_hour bigint default 0,
	error_day bigint default 0
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

create table queue_stats (
	queue uuid,
	action text,
	bucket text,
	ts timestamptz
);

commit;
