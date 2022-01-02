(ns build
  (:require [clojure.tools.build.api :as b]
            [org.corfield.build :as bb]))

(def lib 'moolite.bot)
(def main 'moolite.bot.core)
(def uber-file "bot.jar")

;; if you want a version of MAJOR.MINOR.COMMITS:
(def version (format "1.0.%s" (b/git-count-revs nil)))

(defn run-tests "Run tests" [opts]
  (-> opts (bb/run-tests)))

(defn ci "Run the CI pipeline of tests (and build the uberjar)." [opts]
  (-> opts
      (assoc :lib lib :main main)
      ; (bb/run-tests)
      (bb/clean)
      (bb/uber)))

(defn uber "Build an uberjar" [opts])
