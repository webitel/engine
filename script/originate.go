package script

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

var (
	sdp = "v=0\r\no=Webitel 0 0 IN IP4 0.0.0.0\r\ns=Webitel\r\nc=IN IP4 0.0.0.0\r\nt=0 0\r\nm=audio 9 RTP/AVP 8 0 101\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:101 telephone-event/8000\r\na=sendrecv\r\n"
)

type JsonParams struct {
	Method  string  `json:"method"`
	Ruri    string  `json:"ruri"`
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

type JsonRequest struct {
	Jsonrpc string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	Params  JsonParams `json:"params"`
	Id      int        `json:"id"`
}

type Headers map[string]string

type JsonrpcResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Status  string `json:"Status"`
	RURI    string `json:"RURI"`
	NextHop string `json:"Next-hop"`
	Message string `json:"Message"`
}

func (jp *JsonParams) StringHeaders() string {
	buffer := bytes.Buffer{}

	for k, v := range jp.Headers {
		buffer.WriteString(k)
		buffer.WriteString(": ")
		buffer.WriteString(v)
		buffer.WriteString("\r\n")
	}

	return buffer.String()
}

func (jp *JsonParams) JSON() string {
	if len(jp.Body) > 0 {
		return fmt.Sprintf(`{"method":"%s","ruri":"%s", "headers": "%s", "body": "%s"}`, jp.Method, jp.Ruri, jp.StringHeaders(), jp.Body)
	}
	return fmt.Sprintf(`{"method":"%s","ruri":"%s", "headers": "%s"}`, jp.Method, jp.Ruri, jp.StringHeaders())
}

func (r *JsonRequest) JSON() string {
	return fmt.Sprintf(`{"jsonrpc":"2.0", "method":"%s", "id":%d, "params": %s}`, r.Method, r.Id, r.Params.JSON())
}

func TestOriginate(uri, from, to string) {

	id := 10
	r := JsonRequest{
		Jsonrpc: "2.0",
		Id:      id,
		Method:  "t_uac_dlg",
		Params: JsonParams{
			Method: "INVITE",
			Ruri:   uri,
			Headers: Headers{
				"Contact":      uri,
				"From":         "<sip:1@webitel.lo>;tag=9ef03cf7b43150b3770c2250a6a253d7-2125",
				"To":           to,
				"Content-Type": "application/sdp",
			},
			Body: sdp,
		},
	}

	fmt.Println(r.JSON())
	fmt.Println("--------------")

	req, err := http.NewRequest("POST", "http://192.168.177.192:8000/mi/", bytes.NewBuffer([]byte(r.JSON())))
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(100) * time.Second,
	}
	res, err := client.Do(req)

	if err != nil {
		panic(err.Error())
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	result := JsonrpcResponse{}
	if err = json.Unmarshal(b, &result); err != nil {
		panic(err.Error())
	}

	fmt.Println(result)
	fmt.Println(string(b))

	time.Sleep(time.Millisecond * 1000)

	Refer(id, &result.Result)

}

func Refer(id int, result *Result) {
	reader := textproto.NewReader(bufio.NewReader(strings.NewReader(result.Message)))
	headers, _ := reader.ReadMIMEHeader()

	h := Headers{}

	//h["Via"] = "SIP/2.0/UDP 10.10.10.114:5079;rport;branch=z9hG4bK2039075550"

	h["To"] = headers.Get("From")
	h["From"] = headers.Get("To")
	h["Refer-To"] = "<sip:1@webitel.lo>"
	//h["Refer-By"] = headers.Get("Contact")
	//h["Contact"] = headers.Get("Contact")
	h["Call-Id"] = headers.Get("Call-Id")
	h["CSeq"] = "11 REFER"

	id++

	r := JsonRequest{
		Jsonrpc: "2.0",
		Id:      id,
		Method:  "t_uac_dlg",
		Params: JsonParams{
			Method:  "REFER",
			Ruri:    result.RURI,
			Headers: h,
		},
	}

	fmt.Println("-------------------REFER ------------------")
	fmt.Println(r.JSON())
	fmt.Println("--------------------------------------------")

	req, err := http.NewRequest("POST", "http://192.168.177.192:8000/mi/", bytes.NewBuffer([]byte(r.JSON())))
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(10) * time.Second,
	}
	res, err := client.Do(req)

	if err != nil {
		panic(err.Error())
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	jresult := JsonrpcResponse{}
	if err = json.Unmarshal(b, &jresult); err != nil {
		panic(err.Error())
	}

	fmt.Println(jresult)
	fmt.Println(string(b))

	jresult.Result.RURI = result.RURI

	time.Sleep(time.Millisecond * 500)

	Bye(id, &jresult.Result)
}

func Bye(id int, result *Result) {

	reader := textproto.NewReader(bufio.NewReader(strings.NewReader(result.Message)))
	headers, _ := reader.ReadMIMEHeader()

	h := Headers{}

	h["To"] = headers.Get("From")
	h["CSeq"] = "12 BYE"
	//h["From"] = headers.Get("From")
	//h["Contact"] = "sip:400@webitel.lo"
	h["Call-Id"] = headers.Get("Call-Id")

	id++

	r := JsonRequest{
		Jsonrpc: "2.0",
		Id:      id,
		Method:  "t_uac_dlg",
		Params: JsonParams{
			Method:  "BYE",
			Ruri:    result.RURI,
			Headers: h,
		},
	}

	fmt.Println("-------------------BYE----------------------")
	fmt.Println(r.JSON())
	fmt.Println("--------------------------------------------")

	res, err := http.NewRequest("POST", "http://192.168.177.192:8000/mi/", bytes.NewBuffer([]byte(r.JSON())))
	if err != nil {
		panic(err.Error())
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	//jresult := JsonrpcResponse{}
	//if err = json.Unmarshal(b, &jresult); err != nil {
	//	panic(err.Error())
	//}
	//
	//fmt.Println(jresult)
	fmt.Println(string(b))
}
