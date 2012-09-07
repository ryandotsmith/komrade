require "zlib"
require "sequel"

module Komrade
  class Job

    DB = Sequel.connect(Conf.database_url)

    def self.entity(user_id, name)
      Zlib.crc32([user_id, name].join)
    end

    def self.put(user_id, name, payload)
      entity = entity(user_id, name)
      DB[:jobs].returning(:resource_id).insert(entity: entity, payload: JSON.dump(payload)).pop[:resource_id]
    end

    def self.get(user_id, name)
      entity = entity(user_id, name)
      DB[:jobs].where(entity: entity).first.values_at(:resource_id, :payload)
#     DB[:jobs].returning(:resource_id, :payload).where(entity: entity).limit(1).delete
    end
  end
end
