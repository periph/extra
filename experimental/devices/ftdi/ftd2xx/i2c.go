// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This functionality requires MPSSE.
//
// Implementation based on
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_255_USB%20to%20I2C%20Example%20using%20the%20FT232H%20and%20FT201X%20devices.pdf
//
// Page 18: MPSSE does not automatically support clock stretching for I²C.

package ftd2xx

/*
// Page 10-11.
func (d *device) setI2CLinesIdle() error {
	// Set all D0~D7 lines high.
	// D0: SCL
	// D1: SDA, open drain, pulled up externally
	// D2: DATA IN (?)
	// D3~D7 as inputs
	// C0~C7 to high
	// C6: LED
	// C0~C5, C6~C7 as input
	_, err := write([]byte{0x80, 0xFF, 0xFB, 0x82, 0xFF, 0x40})
	return err
}

// Page 11-12.
func (d *device) setI2CStart() error {
	return nil
}

// Page 12-13.
func (d *device) setI2CStop() error {
}

// Page 13-14.
func (d *device) readByteAndSendNAK() (byte, error) {
}

// Page 14-15.
func (d *device) readBytesAndSendNAK(b []byte) error {
}

// Page 15-16.
func (d *device) sendByteAndCheckACK(b byte) error {
}

// Page 16-17.
func (d *device) sendAddrAndCheckACK(b byte) error {
}

// TODO(maruel): Implement all the utility functions, then expose
// https://periph.io/x/periph/conn/i2c#Bus.
*/
