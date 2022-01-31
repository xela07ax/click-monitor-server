package reqestWrkHub

import (
	"encoding/json"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/lib/inputRpc"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"

	"strings"
	"time"
)

func (c *Client) CallThirdParty(ipAddress, userAgent string) (result model.IPQSRow, err error) {
	if ipAddress == "" {
		err = fmt.Errorf("iP address IS EMPTY")
		return
	}
	result.Timestamp = time.Now()
	result.Uag = userAgent
	heandler := model.Handle{
		Time:        result.Timestamp,
		Send:        !c.hub.cfg.Mock,
		RedirectUrl: fmt.Sprintf("%s%s/%s", c.hub.cfg.UrlPostback, c.sender.IpqsKey, ipAddress),
		Params:      fmt.Sprintf("allow_public_access_points=true&fast=false&lighter_penalties=true&mobile=false&strictness=1&user_agent=%s", strings.ReplaceAll(userAgent, " ", "%20")),
		Method:      "GET",
	}
	dat, err := json.Marshal(heandler)
	respRpc, err := inputRpc.RpcRequest(c.hub.cfg.Service, string(dat), c.sender.HostRepiter, c.hub.Loger)
	if respRpc != nil {
		result.RespStatus = respRpc.RespStatus
		result.RespCode = respRpc.RespCode
		result.RespBody = respRpc.RespBody
	}

	return
}