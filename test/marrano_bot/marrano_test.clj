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
