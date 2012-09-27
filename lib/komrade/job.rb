require "zlib"
require "json"
require "sequel"

require "komrade/conf"

module Komrade
  module Job
    extend self

    DB = Sequel.connect(Conf.database_url)

    def put(user_id, name, payload)
      d = {entity: key(user_id, name), payload: JSON.dump(payload))}
      r = DB[:jobs].returning(:resource_id).insert(d)
      {"job-id" => r[0][:resource_id]}
    end

    def get(user_id, name)
      DB[:jobs].where(entity: key(user_id, name)).map do |r|
        {"job-id" => r[:resource_id], "payload" => r[:payload]}
      end
    end

    private

    def key(*things)
      Zlib.crc32(things.join)
    end

  end
end
