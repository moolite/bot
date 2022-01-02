(ns moolite.bot.handlers-test
  (:require [moolite.bot.marrano :as m]
            [moolite.bot.parse :as p]
            [clojure.test :refer :all]))

(deftest test-command?
  (is (= true (#'p/command? "!foo is so bar"))))

(deftest test-parse-text
  (let [[cmd pred] (#'p/parse "!foo is hard")]
    (is (= cmd "foo") "parses a simple message"))
  (let [[cmd pred] (#'p/parse "!FoO is not case sensitive")]
    (is (= cmd "foo") "is not case sensitive")))

(def sample
 {:update_id 82256110,
  :message {:date 1595506720,
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
            :text "/link"}})

(deftest test-api
  (let [res (#'m/bot-api (merge (:message sample) {:text "/link"}))]
    (is (= (:chat_id res)
           (get-in sample [:message :chat :id]))
        "should answer in the same chat")
    (is (= (:method res)
           "sendMessage")
        "should call the 'sendMessage' API")))
