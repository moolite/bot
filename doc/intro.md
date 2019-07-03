# Introduction to marrano-bot

TODO: write [great documentation](http://jacobian.org/writing/what-to-write/)

'
```clojure
{:update_id 82144142,
 :message {:caption "Пацаны как вам фигурка? 😏",
           :date 1562137806,
           :forward_from_chat {:id -1001291855291,
                               :title "Жопы Твоих Одноклассниц",
                               :type "channel"},
           :chat {:id -284819895,
                  :title "Marrani Unlimited Ltd, l'angolo delle russacchiotte,
                  polacchine, ucrainine, estonine e ungheresine",
                  :type "group",
                  :all_members_are_administrators true},
           :message_id 103590,
           :photo [
            {:file_id "AgADAgADcaoxG7b4UUieC2oKhVfpJ5D48Q4ABCPRDVbRljx5UzEAAgI",
            :file_size 1543, :width 68, :height 90}
            {:file_id "AgADAgADcaoxG7b4UUieC2oKhVfpJ5D48Q4ABH0KOMudMqEHVDEAAgI",
            :file_size 17463, :width 241, :height 320}
            {:file_id "AgADAgADcaoxG7b4UUieC2oKhVfpJ5D48Q4ABC-j2tKAYEK0UjEAAgI",
            :file_size 65658, :width 733, :height 974}
            {:file_id "AgADAgADcaoxG7b4UUieC2oKhVfpJ5D48Q4ABGl7E-5eBhFqVTEAAgI",
            :file_size 68819, :width 602, :height 800}],
           :from {:id 318062977,
                  :is_bot false,
                  :first_name "bot",
                  :last_name "filorusso",
                  :username "liemmo",
                  :language_code "en"},
           :forward_from_message_id 1405,
           :forward_date 1560949248}}
```

```clojure
{:update_id 82144161,
 :message {:message_id 103609,
           :from {:id 318062977, :is_bot false, :first_name "bot",
                  :last_name "filorusso", :username "liemmo", :language_code "en"},
           :chat {:id -284819895, :title "Marrani Unlimited Ltd, l'angolo delle
                  russacchiotte, polacchine, ucrainine, estonine e ungheresine",
                  :type "group", :all_members_are_administrators true},
           :date 1562147526,
           :photo [{:file_id "AgADBAADB7IxGztn4FAUXsGc74tQcIzysBoABOtcBdIQCXLQiMECAAEC",
                    :file_size 689, :width 90, :height 33}
                   {:file_id "AgADBAADB7IxGztn4FAUXsGc74tQcIzysBoABBlcQDutwFOCicECAAEC",
                    :file_size 5622, :width 320, :height 116}
                   {:file_id "AgADBAADB7IxGztn4FAUXsGc74tQcIzysBoABCcJH48Vx0FwisECAAEC",
                    :file_size 21162, :width 732, :height 266}]}}
```

```clojure
{:update_id 82144163, :message {:video {:duration 10, :width 244, :height 432,
:mime_type "video/mp4", :thumb {:file_id "AAQCABPYErgPAASbqijGdh0c1zUWAAIC",
:file_size 11347, :width 181, :height 320}, :file_id
"BAADAgAD1QMAAr__sEhvAnVqy435nwI", :file_size 551497}, :caption "🖤
https://t.me/besplatnie-prostitutki", :date 1562148556, :forward_from_chat {:id
-1001205882637, :title "Dolls 🖤", :type "channel"}, :caption_entities [{:offset
0, :length 38, :type "text_link", :url
"http://new-sexy-dating2.com/?u=g91kte4&o=58zp7gn"}], :chat {:id -284819895,
:title "Marrani Unlimited Ltd, l'angolo delle russacchiotte, polacchine,
ucrainine, estonine e ungheresine", :type "group",
:all_members_are_administrators true}, :message_id 103611, :from {:id 318062977,
:is_bot false, :first_name "bot", :last_name "filorusso", :username "liemmo",
:language_code "en"}, :forward_from_message_id 5873, :forward_date 1561716458}}

{:update_id 82144164, :message {:caption "🖤 https://vk.cc/9wWNjE", :date 1562148565,
:forward_from_chat {:id -1001205882637, :title "Dolls 🖤", :type "channel"},
:caption_entities [{:offset 3, :length 20, :type "url"}], :chat {:id -284819895,
:title "Marrani Unlimited Ltd, l'angolo delle russacchiotte, polacchine,
ucrainine, estonine e ungheresine", :type "group",
:all_members_are_administrators true}, :message_id 103612, :photo [{:file_id
"AgADAgADX6wxG_QtuEhM_XnhY9n5HIL7tw8ABPCseNyjku7YLmoAAgI", :file_size 1441,
:width 60, :height 90} {:file_id
"AgADAgADX6wxG_QtuEhM_XnhY9n5HIL7tw8ABG9938FS4EIzL2oAAgI", :file_size 23034,
:width 215, :height 320} {:file_id
"AgADAgADX6wxG_QtuEhM_XnhY9n5HIL7tw8ABKHUX-wFKOxLMGoAAgI", :file_size 65889,
:width 405, :height 604}], :from {:id 318062977, :is_bot false, :first_name
"bot", :last_name "filorusso", :username "liemmo", :language_code "en"},
:media_group_id "12497188520147100", :forward_from_message_id 5885,
:forward_date 1561797716}}

{:update_id 82144165,
 :message {:date 1562148565,
           :forward_from_chat {:id -1001205882637,
                               :title "Dolls 🖤", :type "channel"},
           :chat {:id -284819895,
                  :title "Marrani Unlimited Ltd, l'angolo delle russacchiotte, polacchine,
                   ucrainine, estonine e ungheresine", :type "group",
                  :all_members_are_administrators true},
           :message_id 103613,
           :photo [{:file_id
                "AgADAgADYKwxG_QtuEiO9OJpN8RFnTTctw8ABNxR8-nhNCXd_WwAAgI", :file_size 1297,
                :width 54, :height 90} {:file_id
                "AgADAgADYKwxG_QtuEiO9OJpN8RFnTTctw8ABEu5ZY5BwBWF_mwAAgI", :file_size 19696,
                :width 192, :height 320} {:file_id
                "AgADAgADYKwxG_QtuEiO9OJpN8RFnTTctw8ABMCam078V_BP_2wAAgI", :file_size 56573,
                :width 362, :height 604}],
           :from {:id 318062977, :is_bot false, :first_name
                       "bot", :last_name "filorusso", :username "liemmo", :language_code "en"},
           :media_group_id "12497188520147100", :forward_from_message_id 5886,
           :forward_date 1561797716}
}

{:update_id 82144166,
 :message {:date 1562148565, :forward_from_chat {:id -1001205882637, :title "Dolls 🖤", :type "channel"},
 :chat {:id -284819895,
        :title "Marrani Unlimited Ltd, l'angolo delle russacchiotte,
        polacchine,ucrainine, estonine e ungheresine",
        :type "group", :all_members_are_administrators true},
 :message_id 103614, :photo [{:file_id
    "AgADAgADYawxG_QtuEhN86qDrpGYmxnXtw8ABNgzrYl484smJ24AAgI", :file_size 1447,
    :width 60, :height 90} {:file_id
    "AgADAgADYawxG_QtuEhN86qDrpGYmxnXtw8ABFGkn9KdujFTKG4AAgI", :file_size 23440,
    :width 215, :height 320} {:file_id
    "AgADAgADYawxG_QtuEhN86qDrpGYmxnXtw8ABH1xbjMoCItEKW4AAgI", :file_size 66682,
    :width 406, :height 604}],
 :from {:id 318062977, :is_bot false, :first_name "bot", :last_name "filorusso",
        :username "liemmo", :language_code "en"},
 :media_group_id "12497188520147100", :forward_from_message_id 5887,
 :forward_date 1561797716}}
```


# Video
```clojure
{:update_id 82144176,
 :message {:message_id 103624,
           :from {:id 318062977, :is_bot false,
                  :first_name "bot",
                  :last_name "filorusso",
                  :username "liemmo",
                  :language_code "en"},
           :chat {:id -284819895,
                  :title "Marrani Unlimited Ltd, l'angolo delle russacchiotte, polacchine,
                  ucrainine, estonine e ungheresine",
                  :type "group",
                  :all_members_are_administrators true},
           :date 1562158143,
           :video {:duration 35,
                   :width 352,
                   :height 640,
                   :mime_type "video/mp4",
                   :thumb {:file_id "AAQEABOA8LIaAAQyRQT8p50uyw53AAIC",
                           :file_size 16440,
                           :width 176,
                           :height 320},
                   :file_id "BAADBAADzQUAAjtn6FAETZn_i9QHtgI",
                   :file_size 9452898},
           :caption "Video from Lor"}}
```
