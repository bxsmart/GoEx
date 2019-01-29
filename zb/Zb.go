package zb

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/bxsmart/GoEx"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MARKET_URL = "http://api.zb.com/data/v1/"
	TICKER_API = "ticker?market=%s"
	DEPTH_API  = "depth?market=%s&size=%d"

	TRADE_URL                 = "https://trade.zb.com/api/"
	GET_ACCOUNT_API           = "getAccountInfo"
	GET_ORDER_API             = "getOrder"
	GET_UNFINISHED_ORDERS_API = "getUnfinishedOrdersIgnoreTradeType"
	CANCEL_ORDER_API          = "cancelOrder"
	PLACE_ORDER_API           = "order"
	WITHDRAW_API              = "withdraw"
	CANCELWITHDRAW_API        = "cancelWithdraw"
)

var onceWsConn sync.Once

type Zb struct {
	httpClient           *http.Client
	accessKey, secretKey string
	ws                   *WsConn
	wsDepthHandleMap     map[string]func(*Depth)
}

func New(httpClient *http.Client, accessKey, secretKey string) *Zb {
	zb := Zb{httpClient: httpClient, accessKey: accessKey, secretKey: secretKey}
	zb.wsDepthHandleMap = make(map[string]func(*Depth))
	return &zb
}

func (zb *Zb) GetExchangeName() string {
	return ZB
}

func (zb *Zb) GetTicker(currency CurrencyPair) (*Ticker, error) {
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	resp, err := HttpGet(zb.httpClient, MARKET_URL+fmt.Sprintf(TICKER_API, symbol))
	if err != nil {
		return nil, err
	}
	//log.Println(resp)
	tickermap := resp["ticker"].(map[string]interface{})

	ticker := new(Ticker)
	ticker.Pair = currency
	ticker.Date, _ = strconv.ParseUint(resp["date"].(string), 10, 64)
	ticker.Buy, _ = strconv.ParseFloat(tickermap["buy"].(string), 64)
	ticker.Sell, _ = strconv.ParseFloat(tickermap["sell"].(string), 64)
	ticker.Last, _ = strconv.ParseFloat(tickermap["last"].(string), 64)
	ticker.High, _ = strconv.ParseFloat(tickermap["high"].(string), 64)
	ticker.Low, _ = strconv.ParseFloat(tickermap["low"].(string), 64)
	ticker.Vol, _ = strconv.ParseFloat(tickermap["vol"].(string), 64)

	return ticker, nil
}

func (zb *Zb) GetDepth(size int, currency CurrencyPair) (*Depth, error) {
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	resp, err := HttpGet(zb.httpClient, MARKET_URL+fmt.Sprintf(DEPTH_API, symbol, size))
	if err != nil {
		return nil, err
	}

	//log.Println(resp)

	asks, isok1 := resp["asks"].([]interface{})
	bids, isok2 := resp["bids"].([]interface{})

	if isok2 != true || isok1 != true {
		return nil, errors.New("no depth data!")
	}
	//log.Println(asks)
	//log.Println(bids)

	depth := new(Depth)
	depth.Pair = currency

	for _, e := range bids {
		var r DepthRecord
		ee := e.([]interface{})
		r.Amount = ee[1].(float64)
		r.Price = ee[0].(float64)

		depth.BidList = append(depth.BidList, r)
	}

	for _, e := range asks {
		var r DepthRecord
		ee := e.([]interface{})
		r.Amount = ee[1].(float64)
		r.Price = ee[0].(float64)

		depth.AskList = append(depth.AskList, r)
	}

	return depth, nil
}

func (zb *Zb) buildPostForm(postForm *url.Values) error {
	postForm.Set("accesskey", zb.accessKey)

	payload := postForm.Encode()
	secretkeySha, _ := GetSHA(zb.secretKey)

	sign, err := GetParamHmacMD5Sign(secretkeySha, payload)
	if err != nil {
		return err
	}

	postForm.Set("sign", sign)
	//postForm.Del("secret_key")
	postForm.Set("reqTime", fmt.Sprintf("%d", time.Now().UnixNano()/1000000))
	return nil
}

func (zb *Zb) GetAccount() (*Account, error) {
	params := url.Values{}
	params.Set("method", "getAccountInfo")
	zb.buildPostForm(&params)
	//log.Println(params.Encode())
	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+GET_ACCOUNT_API, params)
	if err != nil {
		return nil, err
	}

	var respmap map[string]interface{}
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		log.Println("json unmarshal error")
		return nil, err
	}

	if respmap["code"] != nil && respmap["code"].(float64) != 1000 {
		return nil, errors.New(string(resp))
	}

	acc := new(Account)
	acc.Exchange = zb.GetExchangeName()
	acc.SubAccounts = make(map[Currency]SubAccount)

	resultmap := respmap["result"].(map[string]interface{})
	coins := resultmap["coins"].([]interface{})

	acc.NetAsset = ToFloat64(resultmap["netAssets"])
	acc.Asset = ToFloat64(resultmap["totalAssets"])

	for _, v := range coins {
		vv := v.(map[string]interface{})
		subAcc := SubAccount{}
		subAcc.Amount = ToFloat64(vv["available"])
		subAcc.ForzenAmount = ToFloat64(vv["freez"])
		subAcc.Currency = NewCurrency(vv["key"].(string), "").AdaptBchToBcc()
		acc.SubAccounts[subAcc.Currency] = subAcc
	}

	//log.Println(string(resp))
	//log.Println(acc)

	return acc, nil
}

func (zb *Zb) placeOrder(amount, price string, currency CurrencyPair, tradeType int) (*Order, error) {
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	params := url.Values{}
	params.Set("method", "order")
	params.Set("price", price)
	params.Set("amount", amount)
	params.Set("currency", symbol)
	params.Set("tradeType", fmt.Sprintf("%d", tradeType))
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+PLACE_ORDER_API, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//log.Println(string(resp));

	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	code := respmap["code"].(float64)
	if code != 1000 {
		log.Println(string(resp))
		return nil, errors.New(fmt.Sprintf("%.0f", code))
	}

	orid := respmap["id"].(string)

	order := new(Order)
	order.Amount, _ = strconv.ParseFloat(amount, 64)
	order.Price, _ = strconv.ParseFloat(price, 64)
	order.Status = ORDER_UNFINISH
	order.Currency = currency
	order.OrderTime = int(time.Now().UnixNano() / 1000000)
	order.OrderID, _ = strconv.Atoi(orid)

	switch tradeType {
	case 0:
		order.Side = SELL
	case 1:
		order.Side = BUY
	}

	return order, nil
}

func (zb *Zb) LimitBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	return zb.placeOrder(amount, price, currency, 1)
}

func (zb *Zb) LimitSell(amount, price string, currency CurrencyPair) (*Order, error) {
	return zb.placeOrder(amount, price, currency, 0)
}

func (zb *Zb) CancelOrder(orderId string, currency CurrencyPair) (bool, error) {
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	params := url.Values{}
	params.Set("method", "cancelOrder")
	params.Set("id", orderId)
	params.Set("currency", symbol)
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+CANCEL_ORDER_API, params)
	if err != nil {
		log.Println(err)
		return false, err
	}

	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		log.Println(err)
		return false, err
	}

	code := respmap["code"].(float64)

	if code == 1000 {
		return true, nil
	}

	//log.Println(respmap)
	return false, errors.New(fmt.Sprintf("%.0f", code))
}

func parseOrder(order *Order, ordermap map[string]interface{}) {
	//log.Println(ordermap)
	//order.Currency = currency;
	order.OrderID, _ = strconv.Atoi(ordermap["id"].(string))
	order.OrderID2 = ordermap["id"].(string)
	order.Amount = ordermap["total_amount"].(float64)
	order.DealAmount = ordermap["trade_amount"].(float64)
	order.Price = ordermap["price"].(float64)
	//	order.Fee = ordermap["fees"].(float64)
	if order.DealAmount > 0 {
		order.AvgPrice = ToFloat64(ordermap["trade_money"]) / order.DealAmount
	} else {
		order.AvgPrice = 0
	}

	order.OrderTime = int(ordermap["trade_date"].(float64))

	orType := ordermap["type"].(float64)
	switch orType {
	case 0:
		order.Side = SELL
	case 1:
		order.Side = BUY
	default:
		log.Printf("unknown order type %f", orType)
	}

	_status := TradeStatus(ordermap["status"].(float64))
	switch _status {
	case 0:
		order.Status = ORDER_UNFINISH
	case 1:
		order.Status = ORDER_CANCEL
	case 2:
		order.Status = ORDER_FINISH
	case 3:
		order.Status = ORDER_UNFINISH
	}

}

func (zb *Zb) GetOneOrder(orderId string, currency CurrencyPair) (*Order, error) {
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	params := url.Values{}
	params.Set("method", "getOrder")
	params.Set("id", orderId)
	params.Set("currency", symbol)
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+GET_ORDER_API, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//println(string(resp))
	ordermap := make(map[string]interface{})
	err = json.Unmarshal(resp, &ordermap)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	order := new(Order)
	order.Currency = currency

	parseOrder(order, ordermap)

	return order, nil
}

func (zb *Zb) GetUnfinishOrders(currency CurrencyPair) ([]Order, error) {
	params := url.Values{}
	symbol := currency.AdaptBchToBcc().AdaptUsdToUsdt().ToSymbol("_")
	params.Set("method", "getUnfinishedOrdersIgnoreTradeType")
	params.Set("currency", symbol)
	params.Set("pageIndex", "1")
	params.Set("pageSize", "100")
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+GET_UNFINISHED_ORDERS_API, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	respstr := string(resp)
	//println(respstr)

	if strings.Contains(respstr, "\"code\":3001") {
		log.Println(respstr)
		return nil, nil
	}

	var resps []interface{}
	err = json.Unmarshal(resp, &resps)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var orders []Order
	for _, v := range resps {
		ordermap := v.(map[string]interface{})
		order := Order{}
		order.Currency = currency
		parseOrder(&order, ordermap)
		orders = append(orders, order)
	}

	return orders, nil
}

func (zb *Zb) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	return nil, nil
}

func (zb *Zb) GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, error) {
	return nil, nil
}

func (zb *Zb) Withdraw(amount string, currency Currency, fees, receiveAddr, safePwd string) (string, error) {
	params := url.Values{}
	params.Set("method", "withdraw")
	params.Set("currency", strings.ToLower(currency.AdaptBchToBcc().String()))
	params.Set("amount", amount)
	params.Set("fees", fees)
	params.Set("receiveAddr", receiveAddr)
	params.Set("safePwd", safePwd)
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+WITHDRAW_API, params)
	if err != nil {
		log.Println("withdraw fail.", err)
		return "", err
	}

	respMap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		log.Println(err, string(resp))
		return "", err
	}

	if respMap["code"].(float64) == 1000 {
		return respMap["id"].(string), nil
	}

	return "", errors.New(string(resp))
}

func (zb *Zb) CancelWithdraw(id string, currency Currency, safePwd string) (bool, error) {
	params := url.Values{}
	params.Set("method", "cancelWithdraw")
	params.Set("currency", strings.ToLower(currency.AdaptBchToBcc().String()))
	params.Set("downloadId", id)
	params.Set("safePwd", safePwd)
	zb.buildPostForm(&params)

	resp, err := HttpPostForm(zb.httpClient, TRADE_URL+CANCELWITHDRAW_API, params)
	if err != nil {
		log.Println("cancel withdraw fail.", err)
		return false, err
	}

	respMap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		log.Println(err, string(resp))
		return false, err
	}

	if respMap["code"].(float64) == 1000 {
		return true, nil
	}

	return false, errors.New(string(resp))
}

func (zb *Zb) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("unimplements")
}

func (zb *Zb) MarketBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	panic("unsupport the market order")
}

func (zb *Zb) MarketSell(amount, price string, currency CurrencyPair) (*Order, error) {
	panic("unsupport the market order")
}

/**
[{
		"dataType": "dishLength",
		"rate": "24790.55",
		"currentIsBuy": false,
		"dayNumber": 143312714.22,
		"lastTime": 1547889691047,
		"transction": [],
		"currentPrice": 0.00003279,
		"high": 0.00003346,
		"totalBtc": 143312714.22,
		"low": 0.00003247,
		"channel": "dish_length_5_zbbtcdefault",
		"listDown": [[0.00003274, 971.95], [0.00003266, 6798.66], [0.00003258, 3000.00], [0.00003257, 2482.75], [0.00003255, 5052.57]],
		"listUp": [[0.00003281, 622.12], [0.00003282, 661.56], [0.00003284, 50.00], [0.0000329, 6798.66], [0.000033, 5000.00]]
	}
]

@description
Decoder step
1. decode by base64
2. gzip decode
3. json parser
*/
func (zb *Zb) GetDepthWithWs(pair CurrencyPair, handle func(depth *Depth)) error {
	zb.createWsConn()
	sub := fmt.Sprintf("dish_length_5_%sdefault", strings.ToLower(pair.ToSymbol("")))

	zb.wsDepthHandleMap[sub] = handle
	return zb.ws.Subscribe(map[string]interface{}{
		"binary": "true",
		"channel": sub,
		"event": "addChannel",
		"isZip": "true",
	})
}

func (zb *Zb) getPairFromChannel(ch string) CurrencyPair {
	s := strings.Split(ch[:len(ch) - len("default")], "_")
	var currA, currB string
	if strings.HasSuffix(s[3], "usdt") {
		currB = "usdt"
	} else if strings.HasSuffix(s[3], "pax") {
		currB = "pax"
	} else if strings.HasSuffix(s[3], "btc") {
		currB = "btc"
	} else if strings.HasSuffix(s[3], "eth") {
		currB = "eth"
	} else if strings.HasSuffix(s[3], "qc") {
		currB = "qc"
	}

	currA = strings.TrimSuffix(s[3], currB)

	a := NewCurrency(currA, "")
	b := NewCurrency(currB, "")
	pair := NewCurrencyPair(a, b)
	return pair
}


func (zb *Zb) parseDepthData(tick map[string]interface{}) *Depth {
	// bid 买, asks卖
	asks, _ := tick["listUp"].([]interface{})
	bids, _ := tick["listDown"].([]interface{})

	depth := new(Depth)
	for _, r := range asks {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.AskList = append(depth.AskList, dr)
	}

	for _, r := range bids {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.BidList = append(depth.BidList, dr)
	}

	sort.Sort(sort.Reverse(depth.BidList))

	return depth
}


func (zb *Zb) createWsConn() {
	onceWsConn.Do(func() {
		zb.ws = NewWsConn("wss://kline.zb.cn/websocket")
		zb.ws.Heartbeat(func() interface{} {
			return map[string]interface{}{"ping": time.Now().Unix()}
		}, 5*time.Second)

		zb.ws.ReConnect()
		zb.ws.ReceiveMessage(func(msg []byte) {
			resp := string(msg)
			var dataMap map[string]interface{}
			err := json.Unmarshal(msg, &dataMap)
			if err == nil && dataMap["code"].(float64) == 1008 {
				log.Println(resp)
				return
			}

			decodeBytes, err := base64.StdEncoding.DecodeString(resp)
			if err != nil {
				log.Println(err)
			}

			gzipReader, err := gzip.NewReader(bytes.NewReader(decodeBytes))
			data, _ := ioutil.ReadAll(gzipReader)
			var dataArr []map[string]interface{}
			data = data[1 : len(data)-1]
			err = json.Unmarshal(data, &dataArr)
			if err != nil || len(dataArr)<1 {
				log.Println("json unmarshal error for ", string(data))
				return
			}

			datamap := dataArr[0]
			if datamap["ping"] != nil {
				zb.ws.UpdateActivedTime()
				zb.ws.SendWriteJSON(map[string]interface{}{"pong": datamap["ping"]}) // 回应心跳
				return
			}

			if datamap["pong"] != nil { //
				zb.ws.UpdateActivedTime()
				return
			}

			if datamap["id"] != nil { //忽略订阅成功的回执消息
				log.Println(string(data))
				return
			}

			ch := datamap["channel"].(string)
			pair := zb.getPairFromChannel(ch)
			if zb.wsDepthHandleMap[ch] != nil {
				depth := zb.parseDepthData(datamap)
				depth.Pair = pair
				(zb.wsDepthHandleMap[ch])(depth)
				return
			}
		})
	})

}
