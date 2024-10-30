package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/example/common"
	"github.com/injoyai/tdx/protocol"
)

func main() {
	common.Test(func(c *tdx.Client) {
		resp, err := c.GetStockKlineDayAll(protocol.ExchangeSH, "000001")
		logs.PanicErr(err)

		for _, v := range resp.List {
			logs.Debug(v)
		}

		logs.Debug("总数：", resp.Count)
	})
}
