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

// spiPort is an SPI port over an FTDI device in MPSSE mode using the data
// command on the AD bus.
type spiPort struct {
	f     *FT232H
	maxHz int64
}

func (s *spiPort) Close() error {
	s.f.mu.Lock()
	s.f.usingSPI = false
	s.f.mu.Unlock()
	return nil
}

func (s *spiPort) String() string {
	return s.f.String()
}

// Connect implements spi.Port.
func (s *spiPort) Connect(maxHz int64, m spi.Mode, bits int) (spi.Conn, error) {
	if maxHz > 30000000 {
		return nil, errors.New("d2xx: maximum supported clock is 30MHz")
	}
	if maxHz < 100 {
		return nil, errors.New("d2xx: minimum supported clock is 100Hz")
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	ew := gpio.FallingEdge
	er := gpio.RisingEdge
	clk := gpio.Low
	switch m {
	case spi.Mode0:
	case spi.Mode1:
		ew = gpio.RisingEdge
		er = gpio.FallingEdge
	case spi.Mode2:
		clk = gpio.High
	case spi.Mode3:
		ew = gpio.RisingEdge
		er = gpio.FallingEdge
		clk = gpio.High
	default:
		return nil, errors.New("d2xx: unknown mode")
	}

	s.f.mu.Lock()
	defer s.f.mu.Unlock()
	if s.maxHz == 0 || maxHz < s.maxHz {
		if _, err := s.f.h.mpsseClock(int(s.maxHz)); err != nil {
			return nil, err
		}
		s.maxHz = maxHz
	}
	// Note: D4~D8 are unusable.
	// D1 and D3 are output.
	mask := byte(1)<<1 | byte(1)<<3
	b := byte(0)
	if clk {
		b = 1
	}
	if err := s.f.h.mpsseDBus(mask, b); err != nil {
		return nil, err
	}
	s.f.usingSPI = true
	return &spiConn{f: s.f, ew: ew, er: er}, nil
}

// LimitSpeed implements spi.Port.
func (s *spiPort) LimitSpeed(maxHz int64) error {
	if maxHz > 30000000 {
		return errors.New("d2xx: maximum supported clock is 30MHz")
	}
	if maxHz < 100 {
		return errors.New("d2xx: minimum supported clock is 100Hz")
	}
	s.f.mu.Lock()
	defer s.f.mu.Unlock()
	if s.maxHz != 0 && s.maxHz <= maxHz {
		return nil
	}
	s.maxHz = maxHz
	_, err := s.f.h.mpsseClock(int(s.maxHz))
	return err
}

// CLK returns the SCK (clock) pin.
func (s *spiPort) CLK() gpio.PinOut {
	return s.f.D0
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiPort) MOSI() gpio.PinOut {
	return s.f.D1
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiPort) MISO() gpio.PinIn {
	return s.f.D2
}

// CS returns the CSN (chip select) pin.
func (s *spiPort) CS() gpio.PinOut {
	return s.f.D3
}

type spiConn struct {
	f  *FT232H
	ew gpio.Edge
	er gpio.Edge
}

func (s *spiConn) String() string {
	return s.f.String()
}

func (s *spiConn) Tx(w, r []byte) error {
	var p = [1]spi.Packet{{W: w, R: r}}
	return s.TxPackets(p[:])
}

func (s *spiConn) Duplex() conn.Duplex {
	// TODO(maruel): Support half if there's a need.
	return conn.Full
}

func (s *spiConn) TxPackets(pkts []spi.Packet) error {
	// Do not keep the lock during this function. This permits calling on the CBus
	// too.
	// TODO(maruel): One lock for CBus and one for DBus?
	for _, p := range pkts {
		if p.BitsPerWord&7 != 0 {
			return errors.New("d2xx: bits must be a multiple of 8")
		}
		if err := verifyBuffers(p.W, p.R); err != nil {
			return err
		}
	}
	for _, p := range pkts {
		if len(p.W) == 0 && len(p.R) == 0 {
			continue
		}
		// TODO(maruel): Assert CS.
		// TODO(maruel): Bits handling.
		if err := s.f.h.mpsseTx(p.W, p.R, s.ew, s.er, false); err != nil {
			return err
		}
		// TODO(maruel): Deassert CS.
	}
	return nil
}

// CLK returns the SCK (clock) pin.
func (s *spiConn) CLK() gpio.PinOut {
	return s.f.D0
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiConn) MOSI() gpio.PinOut {
	return s.f.D1
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiConn) MISO() gpio.PinIn {
	return s.f.D2
}

// CS returns the CSN (chip select) pin.
func (s *spiConn) CS() gpio.PinOut {
	return s.f.D3
}

func verifyBuffers(w, r []byte) error {
	if len(w) != 0 {
		if len(r) != 0 {
			if len(w) != len(r) {
				return errors.New("d2xx: both buffers must have the same size")
			}
		}
		// TODO(maruel): When the buffer is >64Kb, cut it in parts and do not
		// request a flush. Still try to read though.
		if len(w) > 65536 {
			return errors.New("d2xx: maximum buffer size is 64Kb")
		}
	} else if len(r) != 0 {
		if len(r) > 65536 {
			return errors.New("d2xx: maximum buffer size is 64Kb")
		}
	}
	return nil
}

var _ spi.PortCloser = &spiPort{}
var _ spi.Conn = &spiConn{}
