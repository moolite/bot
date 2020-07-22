(defproject marrano-bot "0.1.0-SNAPSHOT"
  :description "a marrano telegram bot"
  :url "https://bot.frenz.click"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.10.1"]
                 [org.clojure/core.async "0.4.474"]
                 [metosin/reitit "0.5.5"]
                 [metosin/reitit-middleware "0.5.2"]
                 [metosin/muuntaja "0.6.7"]
                 [ring-logger "1.0.1"]
                 [yogthos/config "1.1.1"]
                 [http-kit "2.3.0"]
                 [morse "0.4.0"]]

  :main ^:skip-aot marrano-bot.core
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}
             :prod {:resource-paths ["config/prod"]}
             :dev  {:resource-paths ["config/dev"]}})
