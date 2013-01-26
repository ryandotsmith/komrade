create table queues (
	token uuid default uuid_generate_v4(),
	heroku_id text,
	plain text,
	callback_url text
);

create table jobs (
	queue uuid,
	id uuid,
	payload json
);

create index jobs_by_queue on jobs (queue);
