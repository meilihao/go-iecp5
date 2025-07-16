package tests

import (
	"fmt"
	"log"

	"github.com/thinkgos/go-iecp5/asdu"
	"github.com/thinkgos/go-iecp5/client"
	"github.com/thinkgos/go-iecp5/cs104"
)

var (
	_ cs104.ClientHandlerInterface = new(myClientHandler)
)

type myClientHandler struct {
}

// OnInterrogation 总召唤回复
func (c *myClientHandler) InterrogationHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetInterrogationCmd()
	fmt.Printf("interrogation reply, addr: %d, value: %d\n", addr, value)
	return nil
}

// OnCounterInterrogation 总计数器回复
func (c *myClientHandler) CounterInterrogationHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetCounterInterrogationCmd()
	fmt.Printf("counter interrogation reply, addr: %d, request: 0x%02X, rreeze: 0x%02X\n",
		addr, value.Request, value.Freeze)
	return nil
}

// OnRead 读定值回复
func (c *myClientHandler) ReadHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	return c.ASDUHandler(conn, packet)
}

// OnTestCommand 测试下发回复
func (c *myClientHandler) TestCommandHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetTestCommand()
	fmt.Printf("test cmd reply, addr: %d, value: %t\n", addr, value)
	return nil
}

// OnClockSync 时钟同步回复
func (c *myClientHandler) ClockSyncHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetClockSynchronizationCmd()
	fmt.Printf("clock sync reply, addr: %d, value: %d\n", addr, value.UnixMilli())
	return nil
}

// OnResetProcess 进程重置回复
func (c *myClientHandler) ResetProcessHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetResetProcessCmd()
	fmt.Printf("reset process reply, addr: %d, value: 0x%02X\n", addr, value)
	return nil
}

// OnDelayAcquisition 延迟获取回复
func (c *myClientHandler) DelayAcquisitionHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	addr, value := packet.GetDelayAcquireCommand()
	fmt.Printf("delay acquisition reply, addr: %d, value: %d\n", addr, value)
	return nil
}

// OnASDU 数据正体
func (c *myClientHandler) ASDUHandler(conn asdu.Connect, packet *asdu.ASDU) error {
	log.Printf("---ASDU %+v", packet)

	// 读取设备数据
	switch client.GetDataType(packet.Type) {
	case client.SinglePoint:
		c.onSinglePoint(packet)
	case client.DoublePoint:
		c.onDoublePoint(packet)
	case client.MeasuredValueScaled:
		c.onMeasuredValueScaled(packet)
	case client.MeasuredValueNormal:
		c.onMeasuredValueNormal(packet)
	case client.StepPosition:
		c.onStepPosition(packet)
	case client.BitString32:
		c.onBitString32(packet)
	case client.MeasuredValueFloat:
		c.onMeasuredValueFloat(packet)
	case client.IntegratedTotals:
		c.onIntegratedTotals(packet)
	case client.EventOfProtectionEquipment:
		c.onEventOfProtectionEquipment(packet)
	case client.PackedStartEventsOfProtectionEquipment:
		c.onPackedStartEventsOfProtectionEquipment(packet)
	case client.PackedOutputCircuitInfo:
		c.onPackedOutputCircuitInfo(packet)
	case client.PackedSinglePointWithSCD:
		c.onPackedSinglePointWithSCD(packet)
	default:
		return nil
	}

	return nil
}

func (c *myClientHandler) onSinglePoint(packet *asdu.ASDU) {
	// [M_SP_NA_1], [M_SP_TA_1] or [M_SP_TB_1] 获取单点信息信息体集合
	for _, p := range packet.GetSinglePoint() {
		fmt.Printf("single point, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onDoublePoint(packet *asdu.ASDU) {
	// [M_DP_NA_1], [M_DP_TA_1] or [M_DP_TB_1] 获得双点信息体集合
	for _, p := range packet.GetDoublePoint() {
		fmt.Printf("double point, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onMeasuredValueScaled(packet *asdu.ASDU) {
	// [M_ME_NB_1], [M_ME_TB_1] or [M_ME_TE_1] 获得测量值，标度化值信息体集合
	for _, p := range packet.GetMeasuredValueScaled() {
		fmt.Printf("measured value scaled, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onMeasuredValueNormal(packet *asdu.ASDU) {
	// [M_ME_NA_1], [M_ME_TA_1],[ M_ME_TD_1] or [M_ME_ND_1] 获得测量值,规一化值信息体集合
	for _, p := range packet.GetMeasuredValueNormal() {
		fmt.Printf("measured value normal, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onStepPosition(packet *asdu.ASDU) {
	// [M_ST_NA_1], [M_ST_TA_1] or [M_ST_TB_1] 获得步位置信息体集合
	for _, p := range packet.GetStepPosition() {
		// state：false: 设备未在瞬变状态 true： 设备处于瞬变状态
		fmt.Printf("step position, ioa: %d, state: %t, value: %d\n", p.Ioa, p.Value.HasTransient, p.Value.Val)
	}
}

func (c *myClientHandler) onBitString32(packet *asdu.ASDU) {
	// [M_ME_NC_1], [M_ME_TC_1] or [M_ME_TF_1].获得测量值,短浮点数信息体集合
	for _, p := range packet.GetMeasuredValueFloat() {
		fmt.Printf("bigtstring32, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onMeasuredValueFloat(packet *asdu.ASDU) {
	// [M_ME_NC_1], [M_ME_TC_1] or [M_ME_TF_1].获得测量值,短浮点数信息体集合
	for _, p := range packet.GetMeasuredValueFloat() {
		fmt.Printf("measured value float, ioa: %d, value: %v\n", p.Ioa, p.Value)
	}
}

func (c *myClientHandler) onIntegratedTotals(packet *asdu.ASDU) {
	// [M_IT_NA_1], [M_IT_TA_1] or [M_IT_TB_1]. 获得累计量信息体集合
	for _, p := range packet.GetIntegratedTotals() {
		fmt.Printf("integrated totals, ioa: %d, count: %d, SQ: 0x%02X, CY: %t, CA: %t, IV: %t\n",
			p.Ioa, p.Value.CounterReading, p.Value.SeqNumber, p.Value.HasCarry, p.Value.IsAdjusted, p.Value.IsInvalid)
	}
}

func (c *myClientHandler) onEventOfProtectionEquipment(packet *asdu.ASDU) {
	// [M_EP_TA_1] [M_EP_TD_1] 获取继电器保护设备事件信息体
	for _, p := range packet.GetEventOfProtectionEquipment() {
		fmt.Printf("event of protection equipment, ioa: %d, event: %d, qdp: %d, mesc: %d, time: %d\n",
			p.Ioa, p.Event, p.Qdp, p.Msec, p.Time.UnixMilli())
	}
}

func (c *myClientHandler) onPackedStartEventsOfProtectionEquipment(packet *asdu.ASDU) {
	// [M_EP_TB_1] [M_EP_TE_1] 获取继电器保护设备事件信息体
	p := packet.GetPackedStartEventsOfProtectionEquipment()
	fmt.Printf("packed start events of protection equipment, ioa: %d, event: %d, qdp: %d, mesc: %d, time: %d\n",
		p.Ioa, p.Event, p.Qdp, p.Msec, p.Time.UnixMilli())
}

func (c *myClientHandler) onPackedOutputCircuitInfo(packet *asdu.ASDU) {
	// [M_EP_TC_1] [M_EP_TF_1] 获取继电器保护设备成组输出电路信息信息体
	p := packet.GetPackedOutputCircuitInfo()
	fmt.Printf("packed Output circuit, ioa: %d, qci: %d, qdp: %d, mesc: %d, time: %d\n",
		p.Ioa, p.Oci, p.Qdp, p.Msec, p.Time.UnixMilli())
}

func (c *myClientHandler) onPackedSinglePointWithSCD(packet *asdu.ASDU) {
	// [M_PS_NA_1]. 获得带变位检出的成组单点信息
	for _, p := range packet.GetPackedSinglePointWithSCD() {
		fmt.Printf("packed single point with SCD, ioa: %d, scd: %d, qds: %d\n", p.Ioa, p.Scd, p.Qds)
	}
}
