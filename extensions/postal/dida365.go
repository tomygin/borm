package postal

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type Dida struct {
	Account  string
	Password string

	client  *http.Client
	header  map[string]string
	inboxId string //默认收件箱
}

func (d *Dida) Init() bool {
	//基本初始化
	d.header = make(map[string]string)
	d.header[`User-Agent`] = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36 Edg/111.0.1661.43`
	d.header[`Referer`] = `https://www.dida365.com/`
	d.header[`Accept`] = `application/json, text/plain, */*`
	d.header[`Accept-Encodin`] = `gzip, deflate, br`
	d.header[`Content-Type`] = `application/json;charset=UTF-8`
	d.header[`Accept-Language`] = `zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6`
	d.header[`sec-ch-ua`] = `"Microsoft Edge";v="111", "Not(A:Brand";v="8", "Chromium";v="111"`
	d.header[`sec-ch-ua-platform`] = `"Windows"`
	d.header[`Access-Control-Request-Headers`] = `content-type,x-device,x-requested-with`
	cookiejar, _ := cookiejar.New(nil)
	d.client = &http.Client{Jar: cookiejar}

	loginUrl := `https://api.dida365.com/api/v2/user/signon?wc=true&remember=true`
	loginData, _ := json.Marshal(map[string]string{
		"password": d.Password,
		"username": d.Account,
	})

	body, err := d.req("POST", loginData, loginUrl)
	if err != nil {
		return false
	}

	tmp := make(map[string]interface{})
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return false
	}

	d.inboxId = tmp[`inboxId`].(string)

	return true
}

func (d *Dida) Send(title, msg string) bool {
	url := `https://api.dida365.com/api/v2/batch/task`
	data := didaJsonTmp(title, msg, d.inboxId)

	d.req("POST", data, url)
	return true
}

var _ Msger = (*Dida)(nil)

//---tool---

func dida_use_one_id() string {
	tmp := strings.Repeat("abcdef0123456789", 3)
	b := []byte(tmp)
	id := make([]byte, 24)
	for i := range id {
		id[i] = b[rand.Intn(len(id))]
	}
	return string(id)

}

func dida_use_one_utc_time() string {
	template := "2006-01-02T%15:04:05.000+0000"
	return time.Now().Format(template)
}

func didaJsonTmp(title, content, inboxId string) []byte {
	currTime := dida_use_one_utc_time()
	tmp := make(map[string]interface{})
	addtask := map[string]interface{}{
		`items`:        []interface{}{},
		`reminders`:    []interface{}{},
		`exDate`:       []interface{}{},
		`dueDate`:      nil,
		`priority`:     0,
		`progress`:     0,
		`assignee`:     nil,
		`startDate`:    nil,
		`isFloating`:   false,
		`status`:       0,
		`projectId`:    inboxId,
		`kind`:         nil,
		`createdTime`:  currTime,
		`modifiedTime`: currTime,
		`title`:        title,
		`tags`:         []interface{}{},
		`timeZone`:     `Asia/Shanghai`,
		`id`:           dida_use_one_id(),
		`content`:      content,
	}
	add := make([]map[string]interface{}, 1)
	add[0] = addtask
	tmp[`add`] = add
	tmp[`update`] = []interface{}{}
	tmp[`delete`] = []interface{}{}
	tmp[`addAttachments`] = []interface{}{}
	tmp[`updateAttachments`] = []interface{}{}
	tmp[`deleteAttachments`] = []interface{}{}

	jsonTmp, _ := json.Marshal(tmp)

	return jsonTmp
}

func (d *Dida) req(method string, json []byte, url string) (body []byte, err error) {
	r, _ := http.NewRequest(method, url, bytes.NewBuffer(json))

	for k, v := range d.header {
		r.Header.Add(k, v)
	}

	res, err := d.client.Do(r)

	if err != nil {
		return
	}

	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	return
}
