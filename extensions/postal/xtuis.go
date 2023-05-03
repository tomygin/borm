package postal

import (
	"fmt"
	"net/http"
)

type Xtuis struct {
	Token string
	url   string
}

func (x *Xtuis) Init() bool {
	x.url = "http://wx.xtuis.cn/" + x.Token + ".send?text=%s" + "&desp=%s"
	return true
}

func (x *Xtuis) Send(title, msg string) bool {
	url := fmt.Sprintf(x.url, title, msg)
	resp, ok := http.Get(url)
	return resp.StatusCode == 200 && ok == nil
}

var _ Msger = (*Xtuis)(nil)
