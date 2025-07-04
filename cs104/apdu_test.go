package cs104

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/thinkgos/go-iecp5/asdu"
)

var (
	globalP = asdu.ParamsWide
)

func TestSFrame(t *testing.T) {
	// start
	req := newSFrame(0x01) // server->client
	assert.Equal(t, []byte{0x68, 0x04, 0x01, 0x0, 0x02, 0x0}, req)
}

func TestUFrame(t *testing.T) {
	// start
	startReq := newUFrame(uStartDtActive) // client->server
	assert.Equal(t, []byte{0x68, 0x04, 0x07, 0x0, 0x0, 0x0}, startReq)
	startResp := newUFrame(uStartDtConfirm) // server->client
	assert.Equal(t, []byte{0x68, 0x04, 0x0B, 0x0, 0x0, 0x0}, startResp)
	// stop
	stopReq := newUFrame(uStopDtActive) // client->server
	assert.Equal(t, []byte{0x68, 0x04, 0x13, 0x0, 0x0, 0x0}, stopReq)
	stopResp := newUFrame(uStopDtConfirm) // server->client
	assert.Equal(t, []byte{0x68, 0x04, 0x23, 0x0, 0x0, 0x0}, stopResp)
	// test
	testReq := newUFrame(uTestFrActive) // client->server
	assert.Equal(t, []byte{0x68, 0x04, 0x43, 0x0, 0x0, 0x0}, testReq)
	testResp := newUFrame(uTestFrConfirm) // server->client
	assert.Equal(t, []byte{0x68, 0x04, 0x83, 0x0, 0x0, 0x0}, testResp)
}

// https://blog.redisant.cn/docs/iec104-tutorial/chapter9/
func TestSendInterrogation(t *testing.T) {
	/*
		总召启动、总召确认、总召结束报文格式完全相同，只是传输原因不同（0、FF扇区无数据总召后即结束召唤）。在装置确认总召命令后，若某扇区有数据则先上送数据，再上送总召结束报文

		过程:
		1. c->s : 68 0E DE 00 22 06 | 64 01 06 00 07 11 | 00 00 00 | 14 （总召）

			启动：68

			长度：0E

			----------------------------------

			I格式控制域：DE 00 22 06

			----------------------------------

			ASDU：64

			可变结构：01

			传输原因：06 00

			comaddr：07 11

			信息体地址：00 00 00

			QOI：14
		2. s->c : 68 0E 10 00 0A 00 | 64 01 07 00 07 11 | 00 00 00 | 14 （确认）
	*/
	rStart := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.C_IC_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 1},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.Activation},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x1107),
	})

	rStart.AppendInfoObjAddr(asdu.InfoObjAddr(0))
	rStart.AppendBytes(byte(asdu.QOIStation))

	rStartBuf, _ := rStart.MarshalBinary()
	rStartData, _ := newIFrame(0xDE>>1, 0x06<<7+0x22>>1, rStartBuf)
	assert.Equal(t, []byte{0x68, 0x0E, 0xDE, 0x00, 0x22, 0x06, 0x64, 0x01, 0x06, 0x00, 0x07, 0x11, 0x00, 0x00, 0x00, 0x14}, rStartData)

	rDone := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.C_IC_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 1},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.ActivationCon},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x1107),
	})

	rDone.AppendInfoObjAddr(asdu.InfoObjAddr(0))
	rDone.AppendBytes(byte(asdu.QOIStation))

	rDoneBuf, _ := rDone.MarshalBinary()
	rDoneData, _ := newIFrame(0x10>>1, 0x0A>>1, rDoneBuf)
	assert.Equal(t, []byte{0x68, 0x0E, 0x10, 0x00, 0x0A, 0x00, 0x64, 0x01, 0x07, 0x00, 0x07, 0x11, 0x00, 0x00, 0x00, 0x14}, rDoneData)
}

// https://blog.redisant.cn/docs/iec104-tutorial/chapter9/
func TestCallCount(t *testing.T) {
	/*
		过程:
		1. Client -> Server: 计数量召唤命令（传输原因=激活）
			68-0E | 00-00-00-00 | 65-01-06-00-01-00 | 00-00-00-05

		2. Server -> Client: 计数量召唤命令（传输原因=激活确认）
			68-0E-00-00-02-00-65-01-07-00-01-00-00-00-00-05

		3. Server -> Client: 累计量（传输原因=响应计数量召唤）M_IT_NA_1 -> BCR
			68-1A-02-00-02-00-0F-02-25-00-01-00-01-00-00-00-00-00-00-00-02-00-00-00-00-00-00-00

		4. Server -> Client: 计数量召唤命令（传输原因=激活终止）
			68-0E-04-00-02-00-65-01-0A-00-01-00-00-00-00-05
	*/
	rStart := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.C_CI_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 1},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.Activation},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x0001),
	})

	rStart.AppendInfoObjAddr(asdu.InfoObjAddr(0))
	rStart.AppendBytes(byte(asdu.QCCTotal))

	rStartBuf, _ := rStart.MarshalBinary()
	rStartData, _ := newIFrame(0x0>>1, 0x0>>1, rStartBuf)
	assert.Equal(t, []byte{0x68, 0x0E, 0x00, 0x00, 0x00, 0x00, 0x65, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x05}, rStartData)

	rDone := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.C_CI_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 1},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.ActivationCon},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x0001),
	})

	rDone.AppendInfoObjAddr(asdu.InfoObjAddr(0))
	rDone.AppendBytes(byte(asdu.QCCTotal))

	rDoneBuf, _ := rDone.MarshalBinary()
	rDoneData, _ := newIFrame(0x0>>1, 0x02>>1, rDoneBuf)
	assert.Equal(t, []byte{0x68, 0x0E, 0x00, 0x00, 0x02, 0x00, 0x65, 0x01, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x05}, rDoneData)

	// 68-1A
	// 02-00-02-00
	// 0F-02-25-00-01-00
	// 01-00-00 00-00-00-00-00
	// 02-00-00 00-00-00-00-00
	rResp := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.M_IT_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 2},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.RequestByGeneralCounter},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x0001),
	})

	targetData := []byte{0x68, 0x1A, 0x02, 0x00, 0x02, 0x00, 0x0F, 0x02, 0x25, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData[6:])
	assert.Nil(t, err, nil)

	brcs := asduPack.GetIntegratedTotals()
	assert.Equal(t, 2, len(brcs))
	spew.Dump(brcs)

	rResp.AppendInfoObjAddr(asdu.InfoObjAddr(0x1))
	rResp.AppendBinaryCounterReading(brcs[0].Value)

	rResp.AppendInfoObjAddr(asdu.InfoObjAddr(0x2))
	rResp.AppendBinaryCounterReading(brcs[1].Value)

	rRespBuf, _ := rResp.MarshalBinary()
	rRespData, _ := newIFrame(0x02>>1, 0x02>>1, rRespBuf)
	assert.Equal(t, targetData, rRespData)

	rEnd := asdu.NewASDU(globalP, asdu.Identifier{
		Type:       asdu.C_CI_NA_1,
		Variable:   asdu.VariableStruct{IsSequence: false, Number: 1},
		Coa:        asdu.CauseOfTransmission{Cause: asdu.ActivationTerm},
		OrigAddr:   0,
		CommonAddr: asdu.CommonAddr(0x0001),
	})

	rEnd.AppendInfoObjAddr(asdu.InfoObjAddr(0))
	rEnd.AppendBytes(byte(asdu.QCCTotal))

	rEndBuf, _ := rEnd.MarshalBinary()
	rEndData, _ := newIFrame(0x4>>1, 0x02>>1, rEndBuf)
	assert.Equal(t, []byte{0x68, 0x0E, 0x04, 0x00, 0x02, 0x00, 0x65, 0x01, 0x0A, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x05}, rEndData)
}

// [单点遥信](https://www.claves.cn/archives/8761)
func TestParseM_SP_NA_1(t *testing.T) {
	/*
		01 -> SIQ

		01（类型标识，单点遥信）02（可变结构限定词，有2个遥信上送）14 00 （传输原因，响应总召）01 00（公共地址）
		03 00 00（信息体地址，第3号遥信）00（遥信分）
		04 00 00（信息体地址，第4号遥信）01（遥信合）
	*/
	targetData := []byte{0x01, 0x02, 0x14, 0x00, 0x01, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x01}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetSinglePoint()
	assert.Equal(t, 2, len(datas))
	spew.Dump(datas)
}

// [双点遥信](https://www.claves.cn/archives/8761)
func TestParseM_DP_NA_1(t *testing.T) {
	/*
		03 -> DIQ

		03（类型标识，双点遥信）05（可变结构限定词，有5个遥信上送）14 00 （传输原因，响应总召）01 00（公共地址）
		01 00 00（信息体地址，第1号遥信）02（遥信合）
		06 00 00（信息体地址，第6号遥信）02（遥信合）
		0A 00 00（信息体地址，第10号遥信）01（遥信分）
		0B 00 00（信息体地址，第11号遥信）02（遥信合）
		0C 00 00（信息体地址，第12号遥信）01（遥信分）
	*/
	targetData := []byte{0x03, 0x05, 0x14, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x06, 0x00, 0x00, 0x02, 0x0A, 0x00, 0x00, 0x01, 0x0B, 0x00, 0x00, 0x02, 0x0C, 0x00, 0x00, 0x01}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetDoublePoint()
	assert.Equal(t, 5, len(datas))
	spew.Dump(datas)
}

// [](https://blog.csdn.net/weixin_45089823/article/details/130258022)
func TestParseM_ME_NA_1(t *testing.T) {
	/*
		09 -> NVA + QDS

		09（类型标识，带品质描述的遥测）82（可变结构限定词，有2个连续遥测上送）14 00（传输原因，响应总召唤）01 00（公共地址）
		01 07 00（信息体地址，从0x0701开始第0号遥测）A1 10（遥测值0x10A1）00 （品质描述）89 15（遥测值0x1589）00 （品质描述）
	*/
	targetData := []byte{0x09, 0x82, 0x14, 0x00, 0x01, 0x00, 0x01, 0x07, 0x00, 0xA1, 0x10, 0x00, 0x89, 0x15, 0x00}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetMeasuredValueNormal()
	assert.Equal(t, 2, len(datas))
	spew.Dump(datas)
}

// [短浮点遥测](https://www.claves.cn/archives/8761)
// **example中float32字节序错误与redisant.cn iec104server模拟器相反, 这里已改为与模拟器一致**
func TestParseM_ME_NC_1(t *testing.T) {
	/*
		0D -> IEEE STD 754 + QDS

		0D（类型标识， 带品质描述的遥测）82（可变结构限定词，有2个遥信上送）14 00 （传输原因，响应总召）01 00（公共地址）01 07 00（信息体地址， 从0x0701开始第0号遥测）
		33 33 A3 C0（遥测值, -5.1）00（品质描述）
		9A 19 5C 43（遥测值, 220.1）00（品质描述）
	*/
	targetData := []byte{0x0D, 0x82, 0x14, 0x00, 0x01, 0x00, 0x01, 0x07, 0x00, 0x33, 0x33, 0xA3, 0xC0, 0x00, 0x9A, 0x19, 0x5C, 0x43, 0x00}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetMeasuredValueFloat()
	assert.Equal(t, 2, len(datas))
	spew.Dump(datas)
}

// [SOE主动上送](https://www.claves.cn/archives/8761)
func TestParseM_SP_TB_1(t *testing.T) {
	/*
		1E : SIQ+ CP56Time2a

		1E（类型标识，单点遥信）01（可变结构限定词，有1个SOE）03 00（传输原因， 突发事件）01 00（公共地址）
		08 00 00（信息体地址，第8号遥测）00（遥信分）
		AD 39 （毫秒, 14765）1C（分钟, 28）10 （时, 16） DA (日与星期, 周六, 26日) 0B (月, 11) 05 (年, 05, 2005年)
	*/
	targetData := []byte{0x1E, 0x01, 0x03, 0x00, 0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0xAD, 0x39, 0x1C, 0x10, 0xDA, 0x0B, 0x05}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetSinglePoint()
	assert.Equal(t, 1, len(datas))
	spew.Dump(datas)
}

// [](https://blog.csdn.net/weixin_45089823/article/details/130258022)
func TestParseM_DP_TB_1(t *testing.T) {
	/*
		1f -> DIQ+ CP56Time2a

		1f（类型标识，双点遥信）01（可变结构限定词，有1个SOE）03 00（传输原因，表突发事件）01 00（公共地址）
		0a 00 00（信息体地址，第10号遥信）01（遥信分）2f（毫秒低位）40（毫秒高位）1c（分钟）10（时）7a（日与星期）0b（月）05（年）
	*/
	targetData := []byte{0x1F, 0x01, 0x03, 0x00, 0x01, 0x00, 0x0A, 0x00, 0x00, 0x01, 0x2f, 0x40, 0x1C, 0x10, 0x7a, 0x0B, 0x05}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	datas := asduPack.GetDoublePoint()
	assert.Equal(t, 1, len(datas))
	spew.Dump(datas)
}

// [](https://blog.csdn.net/weixin_45089823/article/details/130258022)
func TestParseC_DC_NA_1(t *testing.T) {
	/*
		2e -> DCO

		2e（类型标识）01（可变结构限定词）06 00（传输原因）01 00（公共地址）05 0b 00（信息体地址，遥控号0xb05-0xb01=4）82（控合）

		DCO执行流程(可能没有撤销):
		1. 预置 : 82
		2. 预置确认 : 82
		3. 执行 : 02
		4. 执行确认 : 02
		5. 撤销: 02
		6. 撤销确认: 02
	*/
	targetData := []byte{0x2e, 0x01, 0x06, 0x00, 0x01, 0x00, 0x05, 0x0b, 0x00, 0x82}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	data := asduPack.GetDoubleCmd()
	spew.Dump(data)
}

// [对时命令](https://www.claves.cn/archives/8761)
func TestParseC_CS_NA_1(t *testing.T) {
	/*
		67 -> CP56Time2a

		c->s:
			67（类型标识，时钟同步）01（可变结构限定词）0600（传输原因，激活）0100（公共地址）000000（信息体地址）
			0102（毫秒，513）03（分钟，3分）04（时，4时）81（日与星期，周四，1日）09（月，9月）05（年，05，2005年）
		s->c:
			67（类型标识，时钟同步）01（可变结构限定词）0700（传输原因，激活确认）0100（公共地址）000000（信息体地址）
			0102（毫秒，513）03（分钟，3分）04（时，4时）81（日与星期，周四，1日）09（月，9月）05（年，05，2005年）
	*/
	// 因内容仅cot有差异, 这里仅解析0600
	targetData := []byte{0x67, 0x01, 0x06, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x81, 0x09, 0x05}
	asduPack := asdu.NewEmptyASDU(globalP)
	err := asduPack.UnmarshalBinary(targetData)
	assert.Nil(t, err, nil)

	addr := asduPack.DecodeInfoObjAddr()
	spew.Dump(addr)
	dt := asduPack.DecodeCP56Time2a()
	spew.Dump(dt)
}
