begin;

create table queues (
	token uuid default uuid_generate_v4(),
	heroku_id text,
	plan text,
	callback_url text
);

create table jobs (
	queue uuid,
	id uuid,
	payload json,
	created_at timestamptz default now(),
	locked_at timestamptz
);

create index jobs_by_queue on jobs (queue);

commit;
