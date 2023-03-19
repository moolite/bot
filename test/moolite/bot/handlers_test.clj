(ns moolite.bot.handlers-test
  (:require [moolite.bot.handlers :as h]
            [moolite.bot.parse :as p]
            [clojure.core.match :refer [match]]
            [clojure.test :refer [deftest is]]))

(def sample
  {:date 1595506720,
   :entities [{:offset 0, :type "bot_command", :length 5}],
   :chat {:first_name "Pinco",
          :username "pinco_pallino",
          :type "private",
          :id 123456,
          :last_name "Pallino"},
   :message_id 212466,
   :from {:first_name "Pinco",
          :language_code "en",
          :is_bot false,
          :username "pinco_pallino",
          :id 123456789,
          :last_name "Pallino"},
   :text "/link"})

(deftest test-api
  (let [res (#'h/telegram-handler {:body-params {:message (assoc sample :text "/link + https://example.com")}})]
    (is (= (:chat_id res)
           (get-in sample [:message :chat :id]))
        "should answer in the same chat")
    (is (= (get-in res [:body :method])
           "sendMessage")
        "should call the 'sendMessage' API"))

  (let [res (#'h/telegram-handler {:body-params {:message (assoc sample :text "")}})]
    (is (= (:chat_id res)
           (get-in sample [:message :chat :id]))
        "should answer in the same chat")
    (is (= (get-in res [:body :method])
           "sendMessage")
        "should call the 'sendMessage' API")))
