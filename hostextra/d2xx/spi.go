// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This functionality requires MPSSE.
//
// Interfacing SPI:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_114_FTDI_Hi_Speed_USB_To_SPI_Example.pdf
//
// Implementation based on
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_180_FT232H%20MPSSE%20Example%20-%20USB%20Current%20Meter%20using%20the%20SPI%20interface.pdf

package d2xx

import (
	"errors"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

type spiPort struct {
	f *FT232H
}

func (s *spiPort) Close() error {
	return nil
}

func (s *spiPort) String() string {
	return "SPI(" + s.f.String() + ")"
}

func (s *spiPort) Connect(maxHz int64, m spi.Mode, bits int) (spi.Conn, error) {
	if maxHz > 30000000 {
		return nil, errors.New("d2xx: maximum supported clock is 30MHz")
	}
	if _, err := s.f.h.mpsseClock(int(maxHz)); err != nil {
		return nil, err
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	ew := gpio.FallingEdge
	er := gpio.RisingEdge
	clk := gpio.Low
	switch m {
	case spi.Mode1:
		ew = gpio.RisingEdge
		er = gpio.FallingEdge
	case spi.Mode2:
		clk = gpio.High
	case spi.Mode3:
		ew = gpio.RisingEdge
		er = gpio.FallingEdge
		clk = gpio.High
	}
	// Would be faster to reuse the cached value. This call corrupts pins D3~D7.
	//s.f.dbus.mpsseDBusRead()
	//s.f.dbus.mpsseDBus(0x6, byte(clk))
	if clk == gpio.High {
	}
	return &spiConn{f: s.f, hz: int(maxHz), ew: ew, er: er}, nil
}

func (s *spiPort) LimitSpeed(maxHz int64) error {
	if maxHz > 30000000 {
		return errors.New("d2xx: maximum supported clock is 30MHz")
	}
	_, err := s.f.h.mpsseClock(int(maxHz))
	return err
}

type spiConn struct {
	f  *FT232H
	hz int
	ew gpio.Edge
	er gpio.Edge
}

func (s *spiConn) String() string {
	return "SPI(" + s.f.String() + ")"
}

func (s *spiConn) Tx(w, r []byte) error {
	// When the buffer is >64Kb, cut it in parts and do not request a
	// flush. Still try to read though.
	// Assert CS
	// Should deassert before calling read.
	// Deassert CS
	return s.f.h.mpsseTx(w, r, s.ew, s.er, false)
}

func (s *spiConn) Duplex() conn.Duplex {
	return conn.Full
}

func (s *spiConn) TxPackets(p []spi.Packet) error {
	// The idea is to push the commands and read a bit but not ask to flush.
	return errors.New("d2xx: TxPackets not implemented")
}

var _ spi.PortCloser = &spiPort{}
var _ spi.Conn = &spiConn{}
