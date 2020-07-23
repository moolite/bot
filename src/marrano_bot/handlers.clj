(ns marrano-bot.handlers
  (:require [reitit.core :as r]
            [reitit.ring :as ring]
            [reitit.ring.middleware.muuntaja :as muuntaja]
            [reitit.ring.middleware.exception :as exception]
            [reitit.dev.pretty :as pretty]
            [config.core :refer [env]]
            [marrano-bot.marrano :refer [bot-api]]
            [clojure.java.io :as io]
            [taoensso.timbre :as timbre :refer [info debug warn error]]))

(def secret
  (or (:secret env)
      "test"))

(defn telegram-handler [r]
  (let [body (:body-params r)
        message (merge {:text ""} ; text can be nil!!!
                       (:messge body))]
    (debug "body" body)
    {:status 200
     :body (bot-api message)}))

(def stack
  (ring/ring-handler
    (ring/router
     [["/" {:get (fn [_] {:status 200 :body "v0.1.0 - marrano-bot"})}]
      ["/t" ["/"
             ["" {:get (fn [_] {:status 200 :body ""})}]
             [secret {:post telegram-handler
                      :get (fn [_] {:status 200 :body {:results "Ko"}})}]]]]
     {:data {:muuntaja muuntaja.core/instance
             :middleware [muuntaja/format-middleware
                          exception/exception-middleware]}
      :exception pretty/exception})
    (ring/redirect-trailing-slash-handler {:method :strip})))

(comment
 (stack {:request-method :get
         :headers {"Content-Type" "application/json"
                   "accept" "application/json"}
         :uri "/t/test"}

    (-> (stack {:request-method :post
                :headers {"Content-Type" "application/json"}
                "accept" "application/json"
                :uri "/t/test"
                :body {:message {:chat {:id 123}}
                                :text "/paris"}})
        (:body)
        (slurp))

    {:update_id 82255996,
     :message {:date 1595448322,
               :animation {:file_name "mp4.mp4", :mime_type "video/mp4",
                           :width 258, :duration 3, :file_size 93745,
                           :file_unique_id "AgADGwIAAlt-xFI",
                           :file_id "CgACAgQAAxkBAAEDPYNfGJwCnelTctoBQPONU6bO9UynswACGwIAAlt-xFIsrwcfAAEn19oaBA",
                           :height 148},
               :chat {:type "group", :title "Marrani Unlimited Ltd, l'angolo delle russacchiotte, polacchine, ucrainine, estonine e ungheresine", :id -284819895, :all_members_are_administrators true}, :document {:file_name "mp4.mp4", :mime_type "video/mp4", :file_size 93745, :file_unique_id "AgADGwIAAlt-xFI", :file_id "CgACAgQAAxkBAAEDPYNfGJwCnelTctoBQPONU6bO9UynswACGwIAAlt-xFIsrwcfAAEn19oaBA"}, :message_id 212355, :from {:first_name "Luca", :is_bot false, :username "LuKeLuky", :id 467996968, :last_name "Bertolani"}, :via_bot {:first_name "Tenor GIF Search", :is_bot true, :username "gif", :id 140267078}}}))
