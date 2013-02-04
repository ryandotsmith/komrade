drop function if exists lock_jobs(uuid, int);
-- token, resolution (in minutes), min time, max time
create function lock_jobs(uuid, int)
returns table(
	id uuid,
	created_at timestamptz,
	locked_at timestamptz,
	payload json
)
as $$
update jobs
set locked_at = now()
where
	jobs.id in (
		select id
		from jobs
		where queue = $1 and locked_at is null
		order by created_at
		limit $2
		for update nowait
	)
RETURNING id, created_at, locked_at, payload
$$ language sql;

drop function if exists summarize_failed(uuid, text);
-- token, resolution (in minutes), min time, max time
create function summarize_failed(uuid, text)
returns table(
	count bigint,
	payload text
)
as $$
select
  count(*) as count,
  payload::text
from
  failed_jobs
where
  queue = $1
  and date_trunc($2, created_at) = date_trunc($2, now())
group by payload::text
order by count
$$ language sql;
