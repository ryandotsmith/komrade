require "json"
require "press"
require "sinatra/base"
require "rack/handler/mongrel"
require "komrade/conf"
require "komrade/authentication"
require "komrade/job"

module Komrade
  class Web < Sinatra::Base
    extend Press
    include Authentication

    helpers do
      def data
        @data ||= JSON.parse(request.body.read) rescue {}
      end
    end

    before do
      authenticate_user
      content_type :json
    end

    post "/queue" do
      self.class.pdfm __FILE__, __method__
      job_id = Job.put(session[:user_id], "default", data['payload'])
      [200, JSON.dump("job-id" => job_id)]
    end

    post "/queue/:name" do |name|
      self.class.pdfm __FILE__, __method__, name: name
      job_id = Job.put(session[:user_id], name, data['payload'])
      [200, JSON.dump("job-id" => job_id)]
    end

    get "/queue" do
      self.class.pdfm __FILE__, __method__
      job_id, payload = Job.get(session[:user_id], "default")
      [200, JSON.dump("job-id" => job_id, "payload" => payload)]
    end

    get "/queue/:name" do |name|
      self.class.pdfm __FILE__, __method__, name: name
      job_id, payload = Job.get(session[:user_id], name)
      [200, JSON.dump("job-id" => job_id, "payload" => payload)]
    end

    def self.run
      ctx app: Conf.app_name, task: "web"
      pdfm __FILE__, __method__, port: Conf.port.to_i
      Rack::Handler::Mongrel.run(Komrade::Web, Port: Conf.port.to_i)
    rescue => e
      pdfme __FILE__, __method__, e
      exit 1
    end
  end
end
