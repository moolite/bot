(defproject marrano-bot "0.1.0-SNAPSHOT"
  :description "a marrano telegram bot"
  :url "https://bot.frenz.click"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.9.0"]
                 [org.clojure/core.async "0.4.474"]
                 [ring/ring-devel "1.7.0-RC1"]
                 [metosin/muuntaja "0.5.0"]
                 [metosin/jsonista "0.2.1"]
                 [codax "1.2.0"]
                 [yogthos/config "1.1.1"]
                 [http-kit "2.3.0"]
                 [compojure "1.6.1"]]

  :plugins [[lein-ring "0.12.1"]]
  :ring {:handler marrano-bot.handlers/stack}

  :main ^:skip-aot marrano-bot.core
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}
             :prod {:resource-paths ["config/prod"]}
             :dev  {:resource-paths ["config/dev"]}})
