(ns marrano-bot.marrano-test
  (:require [marrano-bot.marrano :as m]
            [clojure.test :refer :all]))

(deftest test-command?
  (is (= true (#'m/command? "!foo is so bar"))))

(deftest test-parse-text
  (let [[cmd pred] (#'m/parse-text "!foo is hard")]
    (is (= cmd "foo") "parses a simple message"))
  (let [[cmd pred] (#'m/parse-text "!FoO is not case sensitive")]
    (is (= cmd "foo") "is not case sensitive")))

(def sample-data
  {:update_id 10000
   :message {:message_id 1365
             :date 1441645532
             :chat {:id 1111111
                    :last_name "Test Lastname"
                    :first_name "Test"
                    :username "Test"}
             :from {:id 1111111
                    :last_name "Test Lastname"
                    :first_name "Test"
                    :username "Test"}
             :text "/start"}})

(deftest test-api)
