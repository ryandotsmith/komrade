require "zlib"
require "sequel"

module Komrade
  class Job

    DB = Sequel.connect(Conf.database_url)

    def self.entity(user_id, name)
      Zlib.crc32([user_id, name].join)
    end

    def self.put(user_id, name, payload)
      d = {entity: entity(user_id, name), payload: JSON.dump(payload))}
      r = DB[:jobs].returning(:resource_id).insert(d)
      {"job-id" => r[0][:resource_id]}
    end

    def self.get(user_id, name)
      DB[:jobs].where(entity: entity(user_id, name)).map do |r|
        {"job-id" => r[:resource_id], "payload" => r[:payload]}
      end
    end

  end
end
