Sequel.migration do
  up do
    run <<-SQL
      CREATE EXTENSION pgcrypto;
      CREATE EXTENSION "uuid-ossp";
    SQL

    create_table :jobs do
      primary_key :id
    end

    add_column :jobs, :resource_id,  "uuid DEFAULT uuid_generate_v4()"
    add_column :jobs, :entity,       "integer"
    add_column :jobs, :payload,      "text"
    add_column :jobs, :locked_at,    "timestamptz"
    add_column :jobs, :created_at,   "timestamptz DEFAULT now()"

    run <<-SQL
-- We are declaring the return type to be jobs.
-- This is ok since I am assuming that all of the users added queues will
-- have identical columns to jobs.
-- When QC supports queues with columns other than the default, we will have to change this.

CREATE OR REPLACE FUNCTION lock_head(entity int, top_boundary integer)
RETURNS SETOF jobs AS $$
DECLARE
  unlocked integer;
  relative_top integer;
  job_count integer;
BEGIN
  -- The purpose is to release contention for the first spot in the table.
  -- The select count(*) is going to slow down dequeue performance but allow
  -- for more workers. Would love to see some optimization here...

  EXECUTE 'SELECT count(*) FROM '
    || '(SELECT * FROM jobs WHERE entity = '
    || quote_literal(entity)
    || ' LIMIT '
    || quote_literal(top_boundary)
    || ') limited'
  INTO job_count;

  SELECT TRUNC(random() * (top_boundary - 1))
  INTO relative_top;

  IF job_count < top_boundary THEN
    relative_top = 0;
  END IF;

  LOOP
    BEGIN
      EXECUTE 'SELECT id FROM jobs '
        || ' WHERE locked_at IS NULL'
        || ' AND entity = '
        || quote_literal(entity)
        || ' ORDER BY id ASC'
        || ' LIMIT 1'
        || ' OFFSET ' || quote_literal(relative_top)
        || ' FOR UPDATE NOWAIT'
      INTO unlocked;
      EXIT;
    EXCEPTION
      WHEN lock_not_available THEN
        -- do nothing. loop again and hope we get a lock
    END;
  END LOOP;

  RETURN QUERY EXECUTE 'UPDATE jobs '
    || ' SET locked_at = (CURRENT_TIMESTAMP)'
    || ' WHERE id = $1'
    || ' AND locked_at is NULL'
    || ' RETURNING *'
  USING unlocked;

  RETURN;
END;
$$ LANGUAGE plpgsql;
    SQL
  end

  down do
    drop_table :jobs
  end
end
