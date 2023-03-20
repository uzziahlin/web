package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

type StringValue struct {
	data string
	err  error
}

func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}

	return strconv.ParseInt(s.data, 10, 64)
}

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string

	QueryData url.Values

	RespStatus int
	RespData   []byte

	MatchedRoute string

	T TemplateEngine

	UserValues map[string]any
}

func (c *Context) Render(tplName string, data any) error {

	ctx := c.Req.Context()

	var err error

	c.RespData, err = c.T.Render(ctx, tplName, data)

	if err != nil {
		c.RespStatus = http.StatusInternalServerError
		return err
	}

	c.RespStatus = http.StatusOK

	return nil

}

// BindJSON 将body的json数据绑定到传入参数类型上
func (c *Context) BindJSON(data any) error {
	typ := reflect.TypeOf(data)
	if typ.Kind() != reflect.Pointer {
		return errors.New("传入类型必须为指针类型")
	}

	decoder := json.NewDecoder(c.Req.Body)

	return decoder.Decode(data)
}

func (c *Context) GetForm(key string) StringValue {
	err := c.Req.ParseForm()
	if err != nil {
		return StringValue{"", err}
	}

	data, ok := c.Req.Form[key]

	if !ok {
		return StringValue{"", errors.New("key不存在")}
	}

	return StringValue{
		data: data[0],
		err:  nil,
	}
}

func (c *Context) GetQuery(key string) StringValue {

	if c.QueryData == nil {
		c.QueryData = c.Req.URL.Query()
	}

	data, ok := c.QueryData[key]

	if !ok {
		return StringValue{
			data: "",
			err:  errors.New("key不存在"),
		}
	}

	return StringValue{
		data: data[0],
		err:  nil,
	}
}

func (c *Context) WriteJSONOK(resp any) error {
	return c.WriteJSON(http.StatusOK, resp)
}

func (c *Context) WriteJSON(status int, resp any) error {

	data, err := json.Marshal(resp)

	if err != nil {
		return err
	}

	c.RespStatus = status
	c.RespData = data

	return nil
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}
