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

    def self.route(verb, action, *)
      condition {@instrument_action = action}
      super
    end

    helpers do
      def data
        @data ||= JSON.parse(request.body.read) rescue {}
      end
    end

    before do
      authenticate_user
      content_type :json
      @start_req = Time.now
    end

    after do
      pdfm __FILE__, @instrument_action, elapsed: Time.now - @start_req
    end

    post "/queue" do
      [201, JSON.dump(Job.put(session[:user_id], "default", data['payload']))]
    end

    post "/queue/:name" do |name|
      [201, JSON.dump(Job.put(session[:user_id], name, data['payload']))]
    end

    get "/queue" do
      [200, JSON.dump(Job.get(session[:user_id], "default"))]
    end

    get "/queue/:name" do |name|
      [200, JSON.dump(Job.get(session[:user_id], name))]
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
