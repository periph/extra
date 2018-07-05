// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"testing"

	"periph.io/x/extra/hostextra/d2xx/ftdi"
)

func TestDriver(t *testing.T) {
	defer reset(t)
	drv.numDevices = func() (int, error) {
		return 1, nil
	}
	drv.d2xxOpen = func(i int) (d2xxHandle, int) {
		if i != 0 {
			t.Fatalf("unexpected index %d", i)
		}
		return &d2xxFakeHandle{d: ftdi.FT232R, vid: 0x0403, pid: 0x6014}, 0
	}
	if b, err := drv.Init(); !b || err != nil {
		t.Fatalf("Init() = %t, %v", b, err)
	}
}

//

type d2xxFakeHandle struct {
	d   ftdi.DevType
	vid uint16
	pid uint16
	ua  []byte
	e   ftdi.EEPROM
}

func (d *d2xxFakeHandle) d2xxClose() int {
	return 0
}
func (d *d2xxFakeHandle) d2xxResetDevice() int {
	return 0
}
func (d *d2xxFakeHandle) d2xxGetDeviceInfo() (ftdi.DevType, uint16, uint16, int) {
	return d.d, d.vid, d.pid, 0
}
func (d *d2xxFakeHandle) d2xxEEPROMRead(dev ftdi.DevType, e *ftdi.EEPROM) int {
	*e = d.e
	return 0
}
func (d *d2xxFakeHandle) d2xxEEPROMProgram(e *ftdi.EEPROM) int {
	d.e = *e
	return 0
}
func (d *d2xxFakeHandle) d2xxEraseEE() int {
	return 0
}
func (d *d2xxFakeHandle) d2xxWriteEE(offset uint8, value uint16) int {
	return 1
}
func (d *d2xxFakeHandle) d2xxEEUASize() (int, int) {
	return len(d.ua), 0
}
func (d *d2xxFakeHandle) d2xxEEUARead(ua []byte) int {
	copy(ua, d.ua)
	return 0
}
func (d *d2xxFakeHandle) d2xxEEUAWrite(ua []byte) int {
	d.ua = make([]byte, len(ua))
	copy(d.ua, ua)
	return 0
}
func (d *d2xxFakeHandle) d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int {
	return 0
}
func (d *d2xxFakeHandle) d2xxSetUSBParameters(in, out int) int {
	return 0
}
func (d *d2xxFakeHandle) d2xxSetFlowControl() int {
	return 0
}
func (d *d2xxFakeHandle) d2xxSetTimeouts(readMS, writeMS int) int {
	return 0
}
func (d *d2xxFakeHandle) d2xxSetLatencyTimer(delayMS uint8) int {
	return 0
}
func (d *d2xxFakeHandle) d2xxSetBaudRate(hz uint32) int {
	return 0
}
func (d *d2xxFakeHandle) d2xxGetQueueStatus() (uint32, int) {
	return 0, 0
}
func (d *d2xxFakeHandle) d2xxRead(b []byte) (int, int) {
	return 0, 0
}
func (d *d2xxFakeHandle) d2xxWrite(b []byte) (int, int) {
	return 0, 0
}
func (d *d2xxFakeHandle) d2xxGetBitMode() (byte, int) {
	return 0, 0
}
func (d *d2xxFakeHandle) d2xxSetBitMode(mask, mode byte) int {
	return 0
}

func reset(t *testing.T) {
	drv.reset()
}

func init() {
	reset(nil)
}
