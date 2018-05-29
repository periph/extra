// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import "unsafe"

// EEPROM is the unprocessed EEPROM content.
//
// The EEPROM is in 3 parts: the 56 bytes header, the 4 strings and the rest
// which is used as an 'user area'. The size of the user area depends on the
// length of the strings. The user area content is not included in this struct.
type EEPROM struct {
	// Raw is the raw EEPROM content. It is normally around 56 bytes and excludes
	// the strings.
	Raw []byte

	// The following condition must be true: len(Manufacturer) + len(Desc) <= 40.
	Manufacturer   string
	ManufacturerID string
	Desc           string
	Serial         string
}

func (e *EEPROM) Interpret(t Type, p *ProcessedEEPROM) {
	if len(e.Raw) == 0 {
		return
	}
	// Use the custom structs instead of the ones provided by the library. The
	// reason is that it had to be written for Windows anyway, and this enables
	// using a single code path everywhere.
	hdr := (*eepromHeader)(unsafe.Pointer(&e.Raw[0]))
	p.MaxPower = uint16(hdr.MaxPower)
	p.SelfPowered = hdr.SelfPowered != 0
	p.RemoteWakeup = hdr.RemoteWakeup != 0
	p.PullDownEnable = hdr.PullDownEnable != 0
	switch t {
	case "FT232H":
		h := (*eepromFt232h)(unsafe.Pointer(&e.Raw[0]))
		p.CSlowSlew = h.ACSlowSlew != 0
		p.CSchmittInput = h.ACSchmittInput != 0
		p.CDriveCurrent = uint8(h.ACDriveCurrent)
		p.DSlowSlew = h.ADSlowSlew != 0
		p.DSchmittInput = h.ADSchmittInput != 0
		p.DDriveCurrent = uint8(h.ADDriveCurrent)
		p.Cbus0 = uint8(h.Cbus0)
		p.Cbus1 = uint8(h.Cbus1)
		p.Cbus2 = uint8(h.Cbus2)
		p.Cbus3 = uint8(h.Cbus3)
		p.Cbus4 = uint8(h.Cbus4)
		p.Cbus5 = uint8(h.Cbus5)
		p.Cbus6 = uint8(h.Cbus6)
		p.Cbus7 = uint8(h.Cbus7)
		p.Cbus8 = uint8(h.Cbus8)
		p.Cbus9 = uint8(h.Cbus9)
		p.FT1248Cpol = h.FT1248Cpol != 0
		p.FT1248Lsb = h.FT1248Lsb != 0
		p.FT1248FlowControl = h.FT1248FlowControl != 0
		p.IsFifo = h.IsFifo != 0
		p.IsFifoTar = h.IsFifoTar != 0
		p.IsFastSer = h.IsFastSer != 0
		p.IsFT1248 = h.IsFT1248 != 0
		p.PowerSaveEnable = h.PowerSaveEnable != 0
		p.DriverType = uint8(h.DriverType)
	case "FT232R":
		h := (*eepromFt232r)(unsafe.Pointer(&e.Raw[0]))
		p.IsHighCurrent = h.IsHighCurrent != 0
		p.UseExtOsc = h.UseExtOsc != 0
		p.InvertTXD = h.InvertTXD != 0
		p.InvertRXD = h.InvertRXD != 0
		p.InvertRTS = h.InvertRTS != 0
		p.InvertCTS = h.InvertCTS != 0
		p.InvertDTR = h.InvertDTR != 0
		p.InvertDSR = h.InvertDSR != 0
		p.InvertDCD = h.InvertDCD != 0
		p.InvertRI = h.InvertRI != 0
		p.Cbus0 = uint8(h.Cbus0)
		p.Cbus1 = uint8(h.Cbus1)
		p.Cbus2 = uint8(h.Cbus2)
		p.Cbus3 = uint8(h.Cbus3)
		p.Cbus4 = uint8(h.Cbus4)
		p.DriverType = uint8(h.DriverType)
	default:
		// TODO(maruel): Implement me!
	}
}

// ProcessedEEPROM is the interpreted EEPROM content.
//
// Interpretation depends on the device and this struct us prone to change as
// new FTDI devices are supported.
type ProcessedEEPROM struct {
	MaxPower       uint16 // 0 < MaxPower <= 500
	SelfPowered    bool   // false if powered by the USB bus
	RemoteWakeup   bool   //
	PullDownEnable bool   // true if pull down in suspend enabled

	// FT232H specific data.
	CSlowSlew         bool  // AC bus pins have slow slew
	CSchmittInput     bool  // AC bus pins are Schmitt input
	CDriveCurrent     uint8 // valid values are 4mA, 8mA, 12mA, 16mA
	DSlowSlew         bool  // non-zero if AD bus pins have slow slew
	DSchmittInput     bool  // non-zero if AD bus pins are Schmitt input
	DDriveCurrent     uint8 // valid values are 4mA, 8mA, 12mA, 16mA
	Cbus0             uint8 // Cbus Mux control
	Cbus1             uint8 // Cbus Mux control
	Cbus2             uint8 // Cbus Mux control
	Cbus3             uint8 // Cbus Mux control
	Cbus4             uint8 // Cbus Mux control
	Cbus5             uint8 // Cbus Mux control
	Cbus6             uint8 // Cbus Mux control
	Cbus7             uint8 // Cbus Mux control
	Cbus8             uint8 // Cbus Mux control
	Cbus9             uint8 // Cbus Mux control
	FT1248Cpol        bool  // FT1248 clock polarity - clock idle high (true) or clock idle low (false)
	FT1248Lsb         bool  // FT1248 data is LSB (true), or MSB (false)
	FT1248FlowControl bool  // FT1248 flow control enable
	IsFifo            bool  // Interface is 245 FIFO
	IsFifoTar         bool  // Interface is 245 FIFO CPU target
	IsFastSer         bool  // Interface is Fast serial
	IsFT1248          bool  // Interface is FT1248
	PowerSaveEnable   bool  //
	DriverType        uint8 //

	// FT232R specific data.
	IsHighCurrent bool // If interface is high current
	UseExtOsc     bool // Use External Oscillator
	InvertTXD     bool // Invert TXD
	InvertRXD     bool // Invert RXD
	InvertRTS     bool // Invert RTS
	InvertCTS     bool // Invert CTS
	InvertDTR     bool // Invert DTR
	InvertDSR     bool // Invert DSR
	InvertDCD     bool // Invert DCD
	InvertRI      bool // Invert RI
	//Cbus0         uint8 // Cbus Mux control
	//Cbus1         uint8 // Cbus Mux control
	//Cbus2         uint8 // Cbus Mux control
	//Cbus3         uint8 // Cbus Mux control
	//Cbus4         uint8 // Cbus Mux control
	//DriverType    uint8 //
}

//

// ft232rCBusMuxCtl is stored in the FT232R EEPROM to control each CBus pin.
type ft232rCBusMuxCtl uint8

const (
	// TXDEN; Tx Data Enable. Used with RS485 level converters to enable the line
	// driver during data transmit. It is active one bit time before the start
	// bit up to until the end of the stop bit (C0~C4).
	ft232rCBusTxdEnable ft232rCBusMuxCtl = 0x00
	// PWREN#; Output is low after the device has been configured by USB, then
	// high during USB suspend mode (C0~C4).
	//
	// Must be used with an external 10kΩ pull up.
	ft232rCBusPwrEnable ft232rCBusMuxCtl = 0x01
	// RXLED#; Pulses low when receiving data (C0~C4).
	ft232rCBusRxLED ft232rCBusMuxCtl = 0x02
	// TXLED#; Pulses low when transmitting data (C0~C4).
	ft232rCBusTxLED ft232rCBusMuxCtl = 0x03
	// TX&RXLED#; Pulses low when either receiving or transmitting data (C0~C4).
	ft232rCBusTxRxLED ft232rCBusMuxCtl = 0x04
	// SLEEP# Goes low during USB suspend mode (C0~C4).
	ft232rCBusSleep ft232rCBusMuxCtl = 0x05
	// CLK48 48Mhz +/-0.7% clock output (C0~C4).
	ft232rCBusClk48 ft232rCBusMuxCtl = 0x06
	// CLK24 24Mhz clock output (C0~C4).
	ft232rCBusClk24 ft232rCBusMuxCtl = 0x07
	// CLK12 12Mhz clock output (C0~C4).
	ft232rCBusClk12 ft232rCBusMuxCtl = 0x08
	// CLK6 6Mhz +/-0.7% clock output (C0~C4).
	ft232rCBusClk6 ft232rCBusMuxCtl = 0x09
	// CBitBangI/O; CBus bit-bang mode option (C0~C3).
	ft232rCBusIOMode ft232rCBusMuxCtl = 0x0A
	// BitBangWRn; CBus WR# strobe output (C0~C3).
	ft232rCBusBitBangWR ft232rCBusMuxCtl = 0x0B
	// BitBangRDn; CBus RD# strobe output (C0~C3).
	ft232rCBusBitBangRD ft232rCBusMuxCtl = 0x0C
)

// ft232hCBusMuxCtl is stored in the FT232H EEPROM to control each CBus pin.
type ft232hCBusMuxCtl uint8

const (
	// TriSt-PU; Sets in Tristate (pull up) (C0~C6, C8, C9) on 75kΩ.
	ft232hCBusTristatePU ft232hCBusMuxCtl = 0x00
	// TXLED#; Pulses low when transmitting data (C0~C6, C8, C9).
	ft232hCBusTxLED ft232hCBusMuxCtl = 0x01
	// RXLED#; Pulses low when receiving data (C0~C6, C8, C9).
	ft232hCBusRxLED ft232hCBusMuxCtl = 0x02
	// TX&RXLED#; Pulses low when either receiving or transmitting data (C0~C6,
	// C8, C9).
	ft232hCBusTxRxLED ft232hCBusMuxCtl = 0x03
	// PWREN#; Output is low after the device has been configured by USB, then
	// high during USB suspend mode (C0~C6, C8, C9).
	//
	// Must be used with an external 10kΩ pull up.
	ft232hCBusPwrEnable ft232hCBusMuxCtl = 0x04
	// SLEEP#; Goes low during USB suspend mode (C0~C6, C8, C9).
	ft232hCBusSleep ft232hCBusMuxCtl = 0x05
	// DRIVE1; Drives pin to logic 0 (C0~C6, C8, C9).
	ft232hCBusDrive0 ft232hCBusMuxCtl = 0x06
	// DRIVE1; Drives pin to logic 1 (C0, C5, C6, C8, C9).
	ft232hCBusDrive1 ft232hCBusMuxCtl = 0x07
	// I/O Mode; CBus bit-bang mode option (C5, C6, C8, C9).
	ft232hCBusIOMode ft232hCBusMuxCtl = 0x08
	// TXDEN; Tx Data Enable. Used with RS485 level converters to enable the line
	// driver during data transmit. It is active one bit time before the start
	// bit up to until the end of the stop bit (C0~C6, C8, C9).
	ft232hCBusTxdEnable ft232hCBusMuxCtl = 0x09
	// CLK30 30MHz clock output (C0, C5, C6, C8, C9).
	ft232hCBusClk30 ft232hCBusMuxCtl = 0x0A
	// CLK15 15MHz clock output (C0, C5, C6, C8, C9).
	ft232hCBusClk15 ft232hCBusMuxCtl = 0x0B
	// CLK7.5 7.5MHz clock output (C0, C5, C6, C8, C9).
	ft232hCBusClk7_5 ft232hCBusMuxCtl = 0x0C
)

// eepromHeader is FT_EEPROM_HEADER.
type eepromHeader struct {
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // Defaults to 0x0403; can be changed
	ProductID      uint16  // Defaults to 0x6001 for FT232H/FT232R, relevant value
	SerNumEnable   uint8   // bool Non-zero if serial number to be used
	Unused0        uint8   // For alignment.
	MaxPower       uint16  // 0mA < MaxPower <= 500mA
	SelfPowered    uint8   // bool 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // bool 0 = not capable, 1 = capable; RI# low will wake host in 20ms.
	PullDownEnable uint8   // bool Non zero if pull down in suspend enabled
	Unused1        uint8   // For alignment.
}

// eepromFt232h is FT_EEPROM_232H.
type eepromFt232h struct {
	// eepromHeader
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // Defaults to 0x0403; can be changed
	ProductID      uint16  // Defaults to 0x6001 for FT232H/FT232R, relevant value
	SerNumEnable   uint8   // bool Non-zero if serial number to be used
	Unused0        uint8   // For alignment.
	MaxPower       uint16  // 0mA < MaxPower <= 500mA
	SelfPowered    uint8   // bool 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // bool 0 = not capable, 1 = capable; RI# low will wake host in 20ms.
	PullDownEnable uint8   // bool Non zero if pull down in suspend enabled
	Unused1        uint8   // For alignment.

	// FT232H specific.
	ACSlowSlew        uint8            // bool Non-zero if AC bus pins have slow slew
	ACSchmittInput    uint8            // bool Non-zero if AC bus pins are Schmitt input
	ACDriveCurrent    uint8            // Valid values are 4mA, 8mA, 12mA, 16mA
	ADSlowSlew        uint8            // bool Non-zero if AD bus pins have slow slew
	ADSchmittInput    uint8            // bool Non-zero if AD bus pins are Schmitt input
	ADDriveCurrent    uint8            // Valid values are 4mA, 8mA, 12mA, 16mA
	Cbus0             ft232hCBusMuxCtl //
	Cbus1             ft232hCBusMuxCtl //
	Cbus2             ft232hCBusMuxCtl //
	Cbus3             ft232hCBusMuxCtl //
	Cbus4             ft232hCBusMuxCtl //
	Cbus5             ft232hCBusMuxCtl //
	Cbus6             ft232hCBusMuxCtl //
	Cbus7             ft232hCBusMuxCtl // C7 is limited a sit can only do 'suspend on C7 low'. Defaults pull down.
	Cbus8             ft232hCBusMuxCtl //
	Cbus9             ft232hCBusMuxCtl //
	FT1248Cpol        uint8            // bool FT1248 clock polarity - clock idle high (true) or clock idle low (false)
	FT1248Lsb         uint8            // bool FT1248 data is LSB (true), or MSB (false)
	FT1248FlowControl uint8            // bool FT1248 flow control enable
	IsFifo            uint8            // bool Non-zero if Interface is 245 FIFO
	IsFifoTar         uint8            // bool Non-zero if Interface is 245 FIFO CPU target
	IsFastSer         uint8            // bool Non-zero if Interface is Fast serial
	IsFT1248          uint8            // bool Non-zero if Interface is FT1248
	PowerSaveEnable   uint8            // bool Suspect on ACBus7 low.
	DriverType        uint8            // bool 0 is D2XX, 1 is VCP
}

// eepromFt232r is FT_EEPROM_232R.
type eepromFt232r struct {
	// eepromHeader
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // Defaults to 0x0403; can be changed
	ProductID      uint16  // Defaults to 0x6001 for FT232H/FT232R, relevant value
	SerNumEnable   uint8   // bool Non-zero if serial number to be used
	Unused0        uint8   // For alignment.
	MaxPower       uint16  // 0mA < MaxPower <= 500mA
	SelfPowered    uint8   // bool 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // bool 0 = not capable, 1 = capable; RI# low will wake host in 20ms.
	PullDownEnable uint8   // bool Non zero if pull down in suspend enabled
	Unused1        uint8   // For alignment.

	// FT232R specific.
	IsHighCurrent uint8            // bool High Drive I/Os; 3mA instead of 1mA (@3.3V)
	UseExtOsc     uint8            // bool Use external oscillator
	InvertTXD     uint8            // bool
	InvertRXD     uint8            // bool
	InvertRTS     uint8            // bool
	InvertCTS     uint8            // bool
	InvertDTR     uint8            // bool
	InvertDSR     uint8            // bool
	InvertDCD     uint8            // bool
	InvertRI      uint8            // bool
	Cbus0         ft232rCBusMuxCtl // Default ft232rCBusTxLED
	Cbus1         ft232rCBusMuxCtl // Default ft232rCBusRxLED
	Cbus2         ft232rCBusMuxCtl // Default ft232rCBusTxdEnable
	Cbus3         ft232rCBusMuxCtl // Default ft232rCBusPwrEnable
	Cbus4         ft232rCBusMuxCtl // Default ft232rCBusSleep
	DriverType    uint8            // bool 0 is D2XX, 1 is VCP
}
