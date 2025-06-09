package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/extend"
)

func main() {

	pt := extend.NewPullTrade("./data/trade")

	m, err := tdx.NewManage(nil)
	logs.PanicErr(err)

	err = pt.Pull(m, 2025, "sz000001")
	logs.Err(err)

}
