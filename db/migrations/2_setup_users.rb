Sequel.migration do
  up do
    create_table :users do
      primary_key :id
    end

    add_column :users, :resource_id,  "uuid DEFAULT uuid_generate_v4()"
    add_column :users, :token,        "text DEFAULT gen_random_bytes(32)"
    add_column :users, :created_at,   "timestamptz DEFAULT now()"
    add_column :users, :deleted_at,   "timestamptz"
  end

  down do
    drop_table :users
  end
end
