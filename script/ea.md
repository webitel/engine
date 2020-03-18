#1 Подключиться к разговору
curl -X POST "http://dev.webitel.com/calls/active/{CALL_ID}/eavesdrop"  -H "X-Webitel-Access: ТОКЕН" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"control\": true,\"listen_a\": true, \"listen_b\": true, \"whisper_a\": true, \"whisper_b\": true}"

Где:
 CALL_ID - ид активного звонка, можно получить из: curl -X GET "http://dev.webitel.com/calls/active" -H "accept: application/json"
 control - включить управление DTMF (2 чтобы поговорить с CALL_ID, 1 говорить с другой стороной, 3 - говорить с CALL_ID и другой стороной)
 listen_a - автоматически слушать CALL_ID
 listen_b - автоматически слушать другой канал
 whisper_a - возможность говорить с CALL_ID
 whisper_b - возможность говорить с другой стороной

Также возможно с БПМ
 this.webitel.eavesdrop({
    id: CALL_ID
 })
 
Звонок будет направлен пользователю для которого выдан токен (X-Webitel-Access)