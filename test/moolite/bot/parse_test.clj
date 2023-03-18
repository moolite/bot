(ns moolite.bot.parse-test
  (:require [moolite.bot.parse :as p]
            [clojure.core.match :refer [match]]
            [clojure.test :refer :all]))

(deftest test-text-lang
  (match (#'p/text-lang "/link rm https://ddg.gg")
    [[[command] [:abraxas abraxas] [:del] [:url url]]] (and (is (= command :command) "Parse /.. as :command")
                                                            (is (= abraxas "link") "Parse abraxoides")
                                                            (is (= url "https://ddg.gg" "Parse URLs"))))
  (match (#'p/text-lang "ricorda questo e quest'altro")
    [[[:abraxas abraxas] & _]] (is (= abraxas "ricorda") "can parse commands without /"))

  (match (#'p/text-lang "/link - https://example.com")
    [[[:command] [:abraxas "link"] [operation]]] (is (= operation :del) "Operations can be symbol -"))

  (match (#'p/text-lang "/link + https://example.com")
    [[[:command] [:abraxas "link"] [operation]]] (is (= operation :add) "Operations can be symbol +"))

  (match (#'p/text-lang "/d20 4d6k2")
    [[[:command] [:abraxas "d20"] [:text text]]] (is text "4d6k2")))

(deftest test-parse-message
  (is (match (p/parse-message {:text "/d20"})
        [[[:command] [:abraxas "d20"]]] true
        false)
      "returns a parsed message")
  (is (match (p/parse-message {:text "la donzelletta va per la campagna"})
        [[[:command _] [:text _]]] true
        false)
      "matches non actionable messages"))
