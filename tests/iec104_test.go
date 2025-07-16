package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/thinkgos/go-iecp5/asdu"
	"github.com/thinkgos/go-iecp5/client"
	"github.com/thinkgos/go-iecp5/server"
)

func TestClient(t *testing.T) {
	// srv := startServer()
	// defer srv.Stop()

	settings := client.NewSettings()
	settings.LogCfg = &client.LogCfg{Enable: true}
	c := client.New(settings, &myClientHandler{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	c.SetOnConnectHandler(func(c *client.Client) {
		// 连接成功以后做的操作
		fmt.Printf("connected %s iec104 server\n", settings.Host)
	})

	// Connect后会发送server active
	if err := c.Connect(); err != nil {
		t.Errorf("client connect error %v\n", err)
		t.FailNow()
	}

	fmt.Println("client connected")

	fn := func(c *client.Client) {
		fmt.Println("send cmds")

		//// 发送总召唤
		if err := c.SendInterrogationCmd(commonAddr); err != nil {
			t.Errorf("send interrogation cmd error %v\n", err)
			t.FailNow()
		}

		// 累积量召唤
		if err := c.SendCounterInterrogationCmd(commonAddr); err != nil {
			t.Errorf("send counter interrogation cmd error %v\n", err)
			t.FailNow()
		}

		// read cmd
		if err := c.SendReadCmd(commonAddr, 100); err != nil {
			t.Errorf("send counter interrogation cmd error %v\n", err)
			t.FailNow()
		}

		// 时钟同步
		if err := c.SendClockSynchronizationCmd(commonAddr, time.Now()); err != nil {
			t.Errorf("send clock sync cmd error %v\n", err)
			t.FailNow()
		}

		// test cmd
		if err := c.SendTestCmd(commonAddr); err != nil {
			t.Errorf("send test cmd error %v\n", err)
			t.FailNow()
		}

		// 单点控制
		if err := c.SendCmd(commonAddr, asdu.C_SC_NA_1, asdu.InfoObjAddr(1000), false); err != nil {
			t.Errorf("send single cmd error %v\n", err)
			t.FailNow()
		}

		// 测试等待回复，不能结束太快
		time.Sleep(time.Second * 10)
		wg.Done()
	}

	go fn(c)

	wg.Wait()

	if err := c.Close(); err != nil {
		t.Errorf("close error %v\n", err)
		t.FailNow()
	}
}

func startServer() *server.Server {
	s := server.New(server.NewSettings(), &myServerHandler{})
	s.Start()
	return s
}
