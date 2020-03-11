#1. Поиск абонента по ID = 37
curl -X GET "http://dev.webitel.com/api/call_center/members?page=0&size=10&id=37" -H "accept: application/json" -H "X-Webitel-Access: ss6azqxoupnzdxk93rhdz7b38h"
#Ответ {"items":[{"id":"37","queue":{"id":"5","name":"INBOUND QUEUE"},"priority":10,"expire_at":"20","created_at":"1583324907957","variables":{"Hello":"Hello"},"name":"Hello","timezone":{"id":"20","name":"SystemV/YST9"},"communications":[{"destination":"Hello","type":{},"priority":10,"description":"\"Hello\"","display":"Hello"}],"skills":["20"],"min_offering_at":"20"}]}

#2. Поиск Абонентов по ID очереди
curl -X GET "https://dev.webitel.com/api/call_center/members?page=0&size=4&queue_id=5" -H "accept: application/json" -H "X-Webitel-Access: ss6azqxoupnzdxk93rhdz7b38h"
#Ответ {"next":true,"items":[{"id":"37","queue":{"id":"5","name":"INBOUND QUEUE"},"priority":10,"expire_at":"20","created_at":"1583324907957","variables":{"Hello":"Hello"},"name":"Hello","timezone":{"id":"20","name":"SystemV/YST9"},"communications":[{"destination":"Hello","type":{},"priority":10,"description":"\"Hello\"","display":"Hello"}],"skills":["20"],"min_offering_at":"20"},{"id":"38","queue":{"id":"5","name":"INBOUND QUEUE"},"priority":10,"expire_at":"20","created_at":"1583324909414","variables":{"Hello":"Hello"},"name":"Hello","timezone":{"id":"20","name":"SystemV/YST9"},"communications":[{"destination":"Hello","type":{},"priority":10,"description":"\"Hello\"","display":"Hello"}],"skills":["20"],"min_offering_at":"20"},{"id":"39","queue":{"id":"5","name":"INBOUND QUEUE"},"priority":10,"expire_at":"20","created_at":"1583324910230","variables":{"Hello":"Hello"},"name":"Hello","timezone":{"id":"20","name":"SystemV/YST9"},"communications":[{"destination":"Hello","type":{},"priority":10,"description":"\"Hello\"","display":"Hello"}],"skills":["20"],"min_offering_at":"20"},{"id":"40","queue":{"id":"5","name":"INBOUND QUEUE"},"priority":10,"expire_at":"20","created_at":"1583324910921","variables":{"Hello":"Hello"},"name":"Hello","timezone":{"id":"20","name":"SystemV/YST9"},"communications":[{"destination":"Hello","type":{},"priority":10,"description":"\"Hello\"","display":"Hello"}],"skills":["20"],"min_offering_at":"20"}]}


#3. Поиск Абонентов по Очереди 5 и с номером Hello
curl -X GET "https://dev.webitel.com/api/call_center/members?page=0&size=4&queue_id=5&destination=Hello" -H "accept: application/json" -H "X-Webitel-Access: ss6azqxoupnzdxk93rhdz7b38h"









