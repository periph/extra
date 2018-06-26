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
	"fmt"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// spiMPSEEPort is an SPI port over a FTDI device in MPSSE mode using the data
// command on the AD bus.
type spiMPSEEPort struct {
	c spiMPSEEConn

	// Mutable.
	maxFreq physic.Frequency
}

func (s *spiMPSEEPort) Close() error {
	s.c.f.mu.Lock()
	s.c.f.usingSPI = false
	s.c.f.mu.Unlock()
	return nil
}

func (s *spiMPSEEPort) String() string {
	return s.c.f.String()
}

// Connect implements spi.Port.
func (s *spiMPSEEPort) Connect(f physic.Frequency, m spi.Mode, bits int) (spi.Conn, error) {
	if f > 30*physic.MegaHertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 30MHz", f)
	}
	if f < 100*physic.Hertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	s.c.ew = gpio.FallingEdge
	s.c.er = gpio.RisingEdge
	clk := gpio.Low
	switch m {
	case spi.Mode0:
	case spi.Mode1:
		s.c.ew = gpio.RisingEdge
		s.c.er = gpio.FallingEdge
	case spi.Mode2:
		clk = gpio.High
	case spi.Mode3:
		s.c.ew = gpio.RisingEdge
		s.c.er = gpio.FallingEdge
		clk = gpio.High
	default:
		return nil, errors.New("d2xx: unknown mode")
	}

	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	if s.maxFreq == 0 || f < s.maxFreq {
		if _, err := s.c.f.h.mpsseClock(s.maxFreq); err != nil {
			return nil, err
		}
		s.maxFreq = f
	}
	// Note: D4~D7 are unusable.
	// TODO(maruel): Keep them as-is when transmitting.
	// D1 and D3 are output.
	mask := byte(1)<<1 | byte(1)<<3
	b := byte(0)
	if clk {
		b = 1
	}
	if err := s.c.f.h.mpsseDBus(mask, b); err != nil {
		return nil, err
	}
	s.c.f.usingSPI = true
	return &s.c, nil
}

// LimitSpeed implements spi.Port.
func (s *spiMPSEEPort) LimitSpeed(f physic.Frequency) error {
	if f > 30*physic.MegaHertz {
		return errors.New("d2xx: maximum supported clock is 30MHz")
	}
	if f < 100*physic.Hertz {
		return errors.New("d2xx: minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?")
	}
	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	if s.maxFreq != 0 && s.maxFreq <= f {
		return nil
	}
	s.maxFreq = f
	_, err := s.c.f.h.mpsseClock(s.maxFreq)
	return err
}

// CLK returns the SCK (clock) pin.
func (s *spiMPSEEPort) CLK() gpio.PinOut {
	return s.c.CLK()
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiMPSEEPort) MOSI() gpio.PinOut {
	return s.c.MOSI()
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiMPSEEPort) MISO() gpio.PinIn {
	return s.c.MISO()
}

// CS returns the CSN (chip select) pin.
func (s *spiMPSEEPort) CS() gpio.PinOut {
	return s.c.CS()
}

type spiMPSEEConn struct {
	// Immutable.
	f *FT232H

	// Initialized at Connect().
	ew gpio.Edge
	er gpio.Edge
}

func (s *spiMPSEEConn) String() string {
	return s.f.String()
}

func (s *spiMPSEEConn) Tx(w, r []byte) error {
	var p = [1]spi.Packet{{W: w, R: r}}
	return s.TxPackets(p[:])
}

func (s *spiMPSEEConn) Duplex() conn.Duplex {
	// TODO(maruel): Support half if there's a need.
	return conn.Full
}

func (s *spiMPSEEConn) TxPackets(pkts []spi.Packet) error {
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
func (s *spiMPSEEConn) CLK() gpio.PinOut {
	return s.f.D0
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiMPSEEConn) MOSI() gpio.PinOut {
	return s.f.D1
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiMPSEEConn) MISO() gpio.PinIn {
	return s.f.D2
}

// CS returns the CSN (chip select) pin.
func (s *spiMPSEEConn) CS() gpio.PinOut {
	return s.f.D3
}

//

// spiSyncPort is an SPI port over a FTDI device in synchronous bit-bang mode.
type spiSyncPort struct {
	c spiSyncConn

	// Mutable.
	maxFreq physic.Frequency
}

func (s *spiSyncPort) Close() error {
	s.c.f.mu.Lock()
	s.c.f.usingSPI = false
	s.c.f.mu.Unlock()
	return nil
}

func (s *spiSyncPort) String() string {
	return s.c.f.String()
}

// Connect implements spi.Port.
func (s *spiSyncPort) Connect(f physic.Frequency, m spi.Mode, bits int) (spi.Conn, error) {
	if f > 4*physic.MegaHertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 4MHz", f)
	}
	if f < 100*physic.Hertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	s.c.ew = gpio.FallingEdge
	s.c.er = gpio.RisingEdge
	s.c.clk = gpio.Low
	switch m {
	case spi.Mode0:
	case spi.Mode1:
		s.c.ew = gpio.RisingEdge
		s.c.er = gpio.FallingEdge
	case spi.Mode2:
		s.c.clk = gpio.High
	case spi.Mode3:
		s.c.ew = gpio.RisingEdge
		s.c.er = gpio.FallingEdge
		s.c.clk = gpio.High
	default:
		return nil, errors.New("d2xx: unknown mode")
	}

	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	if s.maxFreq == 0 || f < s.maxFreq {
		if err := s.c.f.SetSpeed(s.maxFreq); err != nil {
			return nil, err
		}
		s.maxFreq = f
	}
	// D1 and D3 are output. D4~D7 are kept as-is.
	mask := byte(1)<<1 | byte(1)<<3 | (s.c.f.dmask & 0xF0)
	if err := s.c.f.SetDBusMask(mask); err != nil {
		return nil, err
	}
	s.c.f.usingSPI = true
	return &s.c, nil
}

// LimitSpeed implements spi.Port.
func (s *spiSyncPort) LimitSpeed(f physic.Frequency) error {
	if f > 4*physic.MegaHertz {
		return fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 4MHz", f)
	}
	if f < 100*physic.Hertz {
		return fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	if s.maxFreq != 0 && s.maxFreq <= f {
		return nil
	}
	if err := s.c.f.SetSpeed(s.maxFreq); err == nil {
		s.maxFreq = f
	}
	return nil
}

// CLK returns the SCK (clock) pin.
func (s *spiSyncPort) CLK() gpio.PinOut {
	return s.c.CLK()
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiSyncPort) MOSI() gpio.PinOut {
	return s.c.MOSI()
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiSyncPort) MISO() gpio.PinIn {
	return s.c.MISO()
}

// CS returns the CSN (chip select) pin.
func (s *spiSyncPort) CS() gpio.PinOut {
	return s.c.CS()
}

type spiSyncConn struct {
	// Immutable.
	f *FT232R

	// Initialized at Connect().
	ew  gpio.Edge
	er  gpio.Edge
	clk gpio.Level
}

func (s *spiSyncConn) String() string {
	return s.f.String()
}

func (s *spiSyncConn) Tx(w, r []byte) error {
	var p = [1]spi.Packet{{W: w, R: r}}
	return s.TxPackets(p[:])
}

func (s *spiSyncConn) Duplex() conn.Duplex {
	// TODO(maruel): Support half if there's a need.
	return conn.Full
}

func (s *spiSyncConn) TxPackets(pkts []spi.Packet) error {
	// We need to 'expand' each bit 4 times * 8 bits, which leads
	// to a 32x memory usage increase.
	// TODO(maruel): It could be possible to lower to 2*8 but starting 'safe'.
	total := 0
	for _, p := range pkts {
		if p.BitsPerWord&7 != 0 {
			return errors.New("d2xx: bits must be a multiple of 8")
		}
		if err := verifyBuffers(p.W, p.R); err != nil {
			return err
		}
		if len(p.W) != 0 {
			total += 4 * 8 * len(p.W)
		} else {
			total += 4 * 8 * len(p.R)
		}
	}
	// Create a large, single chunk.
	we := make([]byte, 0, total)
	re := make([]byte, 0, total)
	m := s.f.dvalue & s.f.dmask & 0xF0
	for _, p := range pkts {
		if len(p.W) == 0 && len(p.R) == 0 {
			continue
		}
		// TODO(maruel): Assert CS.
		// TODO(maruel): Bits handling.
		we = append(we, m)
		// TODO(maruel): Deassert CS.
	}
	if err := s.f.Tx(we, re); err != nil {
		return err
	}
	// Extract data from re into r.
	return nil
}

// CLK returns the SCK (clock) pin.
func (s *spiSyncConn) CLK() gpio.PinOut {
	return s.f.D2 // RTS
}

// MOSI returns the SDO (master out, slave in) pin.
func (s *spiSyncConn) MOSI() gpio.PinOut {
	return s.f.D0 // TX
}

// MISO returns the SDI (master in, slave out) pin.
func (s *spiSyncConn) MISO() gpio.PinIn {
	return s.f.D1 // RX
}

// CS returns the CSN (chip select) pin.
func (s *spiSyncConn) CS() gpio.PinOut {
	return s.f.D3 // CTS
}

//

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

var _ spi.PortCloser = &spiMPSEEPort{}
var _ spi.Conn = &spiMPSEEConn{}
var _ spi.PortCloser = &spiSyncPort{}
var _ spi.Conn = &spiSyncConn{}
