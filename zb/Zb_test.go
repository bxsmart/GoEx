package zb

import (
	"github.com/bxsmart/GoEx"
	"log"
	"net/http"
	"testing"
	"time"
)

var (
	api_key       = ""
	api_secretkey = ""
	zb            = New(http.DefaultClient, api_key, api_secretkey)
)

func TestZb_GetAccount(t *testing.T) {
	acc, err := zb.GetAccount()
	t.Log(err)
	t.Log(acc.SubAccounts[goex.BTC])
}

func TestZb_GetTicker(t *testing.T) {
	ticker, _ := zb.GetTicker(goex.BCH_USD)
	t.Log(ticker)
}

func TestZb_GetDepth(t *testing.T) {
	dep, _ := zb.GetDepth(2, goex.BCH_USDT)
	t.Log(dep)
}

func TestZb_LimitSell(t *testing.T) {
	ord, err := zb.LimitSell("0.001", "75000", goex.NewCurrencyPair2("BTC_QC"))
	t.Log(err)
	t.Log(ord)
}

func TestZb_LimitBuy(t *testing.T) {
	ord, err := zb.LimitBuy("2", "4", goex.NewCurrencyPair2("1ST_QC"))
	t.Log(err)
	t.Log(ord)
}

func TestZb_CancelOrder(t *testing.T) {
	r, err := zb.CancelOrder("201802014255365", goex.NewCurrencyPair2("BTC_QC"))
	t.Log(err)
	t.Log(r)
}

func TestZb_GetUnfinishOrders(t *testing.T) {
	ords, err := zb.GetUnfinishOrders(goex.NewCurrencyPair2("1ST_QC"))
	t.Log(err)
	t.Log(ords)
}

func TestZb_GetOneOrder(t *testing.T) {
	ord, err := zb.GetOneOrder("20180201341043", goex.NewCurrencyPair2("1ST_QC"))
	t.Log(err)
	t.Log(ord)
}

/*
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
*/
func TestZb_GetDepthWithWs(t *testing.T) {
	zb.GetDepthWithWs(goex.NewCurrencyPair2("ETH_QC"), func(depth *goex.Depth) {
		log.Printf("%+v, %+v", depth.AskList[len(depth.AskList) - 1:], depth.BidList[0:1])
	})

	time.Sleep(time.Minute)
}
