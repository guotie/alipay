package alipay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/*
_input_charset=utf-8&body=testjsdzbody&notify_url=http://www.test.com/create_direct_pay_by_user-JAVA-UTF-8/notify_url.jsp&
out_trade_no=9890879868657&partner=2088000000000000&payment_type=1&return_url=http://www.baidu.com&
seller_id=2088000000000000&service=create_direct_pay_by_user&subject=testjsdz&
total_fee=0.01svzitn**********pslfal77xlxm0qhc
*/
type AlipayParameters struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Body         string  `json:"body"`           //订单描述
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerEmail  string  `json:"seller_email"`   //卖家支付宝邮箱
	Service      string  `json:"service"`        //接口名称
	ShowUrl      string  `json:"show_url"`       // 商品展示url
	Subject      string  `json:"subject"`        //商品名称
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
	TotalFee     float32 `json:"total_fee"`      //总价
}

// 按照支付宝规则生成sign
func sign(param interface{}, key string) string {
	//解析为字节数组
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return ""
	}

	//重组字符串
	var sign string
	oldString := string(paramBytes)

	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)

	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)
	paramArray := strings.Split(oldString, ",")

	for _, v := range paramArray {
		detail := strings.SplitN(v, ":", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			//total_fee转化为2位小数
			if detail[0] == "total_fee" {
				number, _ := strconv.ParseFloat(detail[1], 32)
				detail[1] = strconv.FormatFloat(number, 'f', 2, 64)
			}
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}
	}

	//追加密钥
	sign += key

	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}

type Kvpair struct {
	K, V string
}

type Kvpairs []Kvpair

func (t Kvpairs) Less(i, j int) bool {
	return t[i].K < t[j].K
}

func (t Kvpairs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t Kvpairs) Len() int {
	return len(t)
}

func (t Kvpairs) Sort() {
	sort.Sort(t)
}

func (t Kvpairs) RemoveEmpty() (t2 Kvpairs) {
	for _, kv := range t {
		if kv.V != "" {
			t2 = append(t2, kv)
		}
	}
	return
}

func (t Kvpairs) Join() string {
	var strs []string
	for _, kv := range t {
		strs = append(strs, kv.K+"="+kv.V)
	}
	return strings.Join(strs, "&")
}

func md5Sign(str, key string) string {
	h := md5.New()
	io.WriteString(h, str)
	io.WriteString(h, key)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func verifySign(key string, u url.Values) (err error) {
	p := Kvpairs{}
	sign := ""
	for k := range u {
		v := u.Get(k)
		switch k {
		case "sign":
			sign = v
			continue
		case "sign_type":
			continue
		}
		p = append(p, Kvpair{k, v})
	}
	if sign == "" {
		err = fmt.Errorf("sign not found")
		return
	}
	p = p.RemoveEmpty()
	p.Sort()
	fmt.Println(u)
	if md5Sign(p.Join(), key) != sign {
		err = fmt.Errorf("sign invalid")
		return
	}
	return
}

// 重写上面的sign
func signParams(param *AlipayParameters, key string) Kvpairs {
	p := Kvpairs{
		Kvpair{`_input_charset`, `utf-8`},
		Kvpair{`out_trade_no`, param.OutTradeNo},
		Kvpair{`partner`, param.Partner},
		Kvpair{`payment_type`, fmt.Sprint(param.PaymentType)},
		Kvpair{`notify_url`, param.NotifyUrl},
		Kvpair{`return_url`, param.ReturnUrl},
		Kvpair{`subject`, param.Subject},
		Kvpair{`total_fee`, fmt.Sprintf("%.2f", param.TotalFee)},
		Kvpair{`body`, param.Body},
		Kvpair{`service`, param.Service},
		Kvpair{`show_url`, param.ShowUrl},
		Kvpair{`seller_id`, param.Partner},
	}
	p = p.RemoveEmpty()
	p.Sort()

	//fmt.Println(p)
	//fmt.Println(p.Join())
	//fmt.Println(key)

	//sign := "31f171cc5d5d5f50d74ffe4f44d32745"
	sign := md5Sign(p.Join(), key)
	p = append(p, Kvpair{`sign`, sign})
	p = append(p, Kvpair{`sign_type`, `MD5`})
	return p
}
