package lnbits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type Lnbits struct {
	LNBITS_BACKEND     string
	LNBITS_WALLET_ID   string
	LNBITS_INVOICE_KEY string
	LNBITS_ADMIN_KEY   string
}

func Connect(backend string, wallet_id string, invoice_key string, admin_key string) *Lnbits {
	lnbits := &Lnbits{
		LNBITS_BACKEND:     backend,
		LNBITS_WALLET_ID:   wallet_id,
		LNBITS_INVOICE_KEY: invoice_key,
		LNBITS_ADMIN_KEY:   admin_key,
	}
	return lnbits
}

func (ln Lnbits) Call(method string, path string, data map[string]interface{}) (gjson.Result, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return gjson.Result{}, err
	}

	config, err := http.NewRequest(strings.ToUpper(method), ln.LNBITS_BACKEND+path, bytes.NewBuffer(body))
	if err != nil {
		return gjson.Result{}, err
	}

	config.Header.Add("Content-Type", "application/json")
	if method == "POST" && path == "/v1/payments" {
		if strings.Contains(fmt.Sprintf("%v", data["bolt11"]), "lnb") == true {
			config.Header.Add("X-Api-Key", ln.LNBITS_ADMIN_KEY)
		} else {
			config.Header.Add("X-Api-Key", ln.LNBITS_INVOICE_KEY)
		}
	} else {
		config.Header.Add("X-Api-Key", ln.LNBITS_INVOICE_KEY)
	}

	request := &http.Client{}
	response, err := request.Do(config)
	if err != nil {
		return gjson.Result{}, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(body), nil
}

func (ln Lnbits) CreateInvoice(value int64, memo string, webhook string) (gjson.Result, error) {
	data := map[string]interface{}{
		"amount":  value,
		"memo":    memo,
		"unit":    "sat",
		"out":     false,
		"webhook": webhook,
	}
	return ln.Call("POST", "/v1/payments", data)
}

func (ln Lnbits) PayInvoice(bolt11 string) (gjson.Result, error) {
	data := map[string]interface{}{"bolt11": bolt11, "out": true}
	return ln.Call("POST", "/v1/payments", data)
}

func (ln Lnbits) StatusInvoice(payment_hash string) (gjson.Result, error) {
	return ln.Call("GET", "/v1/payments/"+payment_hash, nil)
}

func (ln Lnbits) DecodeInvoice(invoice string) (gjson.Result, error) {
	data := map[string]interface{}{"data": invoice}
	return ln.Call("GET", "/v1/payments/decode", data)
}
