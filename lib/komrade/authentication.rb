require "komrade/user"

module Komrade
  module Authentication
    include Press

    def authenticate_user
      unless authenticated?
        if proper_request?
          id, token = *auth.credentials
          if User.auth?(id, token)
            session[:user_id] = params[:user_id] = id
          else
            pdfm __FILE__, __method__, at: "unauthenticated", id: id, ip: ip
            unauthenticated!
          end
        else
          pdfm __FILE__, __method__, at: "bad-request", ip: ip
          bad_request!
        end
      end
    end

    def proper_request?
      auth.provided? && auth.basic?
    end

    def authenticated?
      session[:user_id]
    end

      def auth
      @auth ||= Rack::Auth::Basic::Request.new(request.env)
    end

    def bad_request!
      throw(:halt, [400, JSON.dump(msg: "Bad Request")])
    end

    def unauthenticated!
      response['WWW-Authenticate'] = "Basic Restricted Area"
      throw(:halt, [401, JSON.dump(msg: "Unauthorized")])
    end

    def ip
      request.env["REMOTE_ADDR"]
    end
  end
end
