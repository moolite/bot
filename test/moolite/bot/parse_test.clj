(ns moolite.bot.parse-test
  (:require [moolite.bot.parse :as p]
            [clojure.core.match :refer [match]]
            [clojure.test :refer :all]))

(deftest test-text-lang
  (match (#'p/text-lang "/link rm https://ddg.gg")
    [_ [command] [:abraxas abraxas] [:del] [:url url]]
    (and (is (= command :command) "Parse /.. as :command")
         (is (= abraxas "link") "Parse abraxoides")
         (is (= url "https://ddg.gg") "Parse URLs"))
    :else (is false))

  (match (#'p/text-lang "ricorda questo e quest'altro")
    [_ [:abraxas abraxas] & _]
    (is (= abraxas "ricorda") "can parse commands without /")
    :else (is false))

  (match (#'p/text-lang "/link - https://example.com")
    [_ [:command] [:abraxas "link"] [operation] & _]
    (is (= operation :del) "Operations can be symbol -")
    :else (is false))

  (match (#'p/text-lang "/link + https://example.com")
    [_ [:command] [:abraxas "link"] [operation] & _]
    (is (= operation :add) "Operations can be symbol +")
    :else (is false))

  (match (#'p/text-lang "/d20 4d6k2")
    [_ [:command] [:abraxas "d20"] [:text text]]
    (is text "4d6k2")
    :else (is false))

  (match (#'p/text-lang "/d20@marrano-bot 4d6k2")
    [_ [:command] [:abraxas "d20"] [:text text]]
    (is text "4d6k2")
    :else (is false)))

(deftest test-parse-message
  (is (match (p/parse-message {:text "/d20"})
        [_ [:command] [:abraxas abraxas] & _]
        (is abraxas "d20")
        :else false)
      "returns a parsed message")
  (is (match (p/parse-message {:text "la donzelletta va per la campagna"})
        [_ [:abraxas "la"] [:text _]]
        true
        :else false)
      "matches non actionable messages"))

(match (p/parse-message {:text "/d20"})
  [:message [:command] & _]
  'foo
  :else 'boo)
(p/parse-message {:text "/link + https://foo"})
