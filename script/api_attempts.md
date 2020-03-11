Поиск попыток звонков.

Время с (created_at.from) по (created_at.to) обязательное.
#curl -X GET "http://10.10.10.25:1907/call_center/queues/attempts/history?size=5&created_at.from=1577829600000&created_at.to=1583336459089" -H "accept: application/json" -H "X-Webitel-Access: ss6azqxoupnzdxk93rhdz7b38h"
ответ: 
{"items":[{"variables":{"Hello":"Hello"},"id":"10486511","member":{"id":"5520159","name":"Hello"},"created_at":"1583318072556","destination":{"id":"0","destination":"c4aa482d164fd4b9ed85437bbd7545bd","type":{"id":"1","name":""},"priority":1,"description":"","resource":{"id":"1","name":""},"display":"","state":0,"last_activity_at":"591628929158","attempts":0,"last_cause":""},"weight":0,"originate_at":"0","answered_at":"0","bridged_at":"0","hangup_at":"0","resource":{"id":"2","name":"10"},"leg_a_id":"","leg_b_id":"","result":"dfdsadsa","agent":{"id":"100","name":"AAA"},"bucket":{"id":"1","name":"Hello"},"active":false,"queue":{"id":"2","name":"5033e723915391711d2b"}},{"variables":{"Hello":"Hello"},"id":"10486510","member":{"id":"6139370","name":"Hello"},"created_at":"1583318072556","destination":{"id":"0","destination":"6c7667767e9e5d5a04d64a6a7c68b050","type":{"id":"2","name":""},"priority":1,"description":"","resource":{"id":"2","name":""},"display":"","state":0,"last_activity_at":"4385974145555","attempts":0,"last_cause":""},"weight":0,"originate_at":"0","answered_at":"0","bridged_at":"0","hangup_at":"0","resource":{"id":"18","name":"423423"},"leg_a_id":"dsadsa","leg_b_id":"","result":"","agent":{"id":"100","name":"AAA"},"bucket":{"id":"1","name":"Hello"},"active":false,"queue":{"id":"2","name":"5033e723915391711d2b"}}],"next":false}

Поиск по member_id
#curl -X GET "http://10.10.10.25:1907/call_center/queues/attempts/history?size=5&created_at.from=1577829600000&created_at.to=1583336459089&member_id=5520159" -H "accept: application/json" -H "X-Webitel-Access: ss6azqxoupnzdxk93rhdz7b38h"
ответ:
{"items":[{"variables":{"Hello":"Hello"},"id":"10486511","member":{"id":"5520159","name":"Hello"},"created_at":"1583318072556","destination":{"id":"0","destination":"c4aa482d164fd4b9ed85437bbd7545bd","type":{"id":"1","name":""},"priority":1,"description":"","resource":{"id":"1","name":""},"display":"","state":0,"last_activity_at":"591628929158","attempts":0,"last_cause":""},"weight":0,"originate_at":"0","answered_at":"0","bridged_at":"0","hangup_at":"0","resource":{"id":"2","name":"10"},"leg_a_id":"","leg_b_id":"","result":"dfdsadsa","agent":{"id":"100","name":"AAA"},"bucket":{"id":"1","name":"Hello"},"active":false,"queue":{"id":"2","name":"5033e723915391711d2b"}}],"next":false}

Фильр работает как И
 возможно фильровать:
 - member_id - Абонент
 - id - Ид попытки
 - queue_id - очередь
 - agent_id - агент
 - result - результат попытки
 - bucket_id - Ид бакета попытки
 