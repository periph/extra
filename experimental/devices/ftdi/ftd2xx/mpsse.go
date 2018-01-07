// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

/*
// Requires a f232h, ft2232h, ft4232h.
func (d *device) setupMPSSE() error {
	// Pre-state:
	// - Write EEPROM i.IsFifo = true so the device DBus is started in tristate.

	// http://www.ftdichip.com/Support/Documents/AppNotes/AN_255_USB%20to%20I2C%20Example%20using%20the%20FT232H%20and%20FT201X%20devices.pdf
	// Page 6
	FT_SetUSBParameters(ftHandle, 65536, 65535); // Set USB request transfer sizes
	FT_SetChars(ftHandle, false, 0, false, 0); // Disable event/error characters
	FT_SetTimeouts(ftHandle, 5000, 5000); // Set rd/wr timeouts to 5 sec
	FT_SetLatencyTimer(ftHandle, 16); // Latency timer at default 16ms
	FT_SetBitMode(ftHandle, 0x0, 0x00); // Reset mode to setting in EEPROM
	FT_SetBitMode(ftHandle, 0x0, 0x02); // Switch to MPSEE

	// Write a bad command and ensure it returned correctly.

	// FT_Write(ftHandle, OutputBuffer, dwNumBytesToSend, &dwNumBytesSent)
	if _, err := write([]byte{0xAA}); err != nil {
		return err
	}
	var b [2]byte
	if _, err := read(b[:]); err != nil {
		return err
	}
	// 0xFA means invalid command, 0xAA is the command echoed back.
	if b[0] != 0xFA || b[1] != 0xAA {
		return err
	}
	// Then repeat with 0xAB. No idea why.

	// 0x8A: Disable clock divide-by-5; resulting in 60MHz master clock.
	// 0x97: Disable adaptive clocking.
	// 0x8C: Enable 3 phase data clocking: data is valid on both clock edges.
	// Other I²C stuff skipped.
	// 0x85: Disable internal loppback.
	write([]byte{0x8A, 0x97, 0x8C, 0x85})
	return nil
}
*/
