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
	s.maxFreq = 0
	s.c.edgeInvert = false
	s.c.clkActiveLow = false
	s.c.noCS = false
	s.c.lsbFirst = false
	s.c.halfDuplex = false
	s.c.f.mu.Unlock()
	return nil
}

func (s *spiMPSEEPort) String() string {
	return s.c.f.String()
}

// Connect implements spi.Port.
func (s *spiMPSEEPort) Connect(f physic.Frequency, m spi.Mode, bits int) (spi.Conn, error) {
	if f > physic.GigaHertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 30MHz", f)
	}
	if f > 30*physic.MegaHertz {
		// TODO(maruel): Figure out a way to communicate that the speed was lowered.
		// https://github.com/google/periph/issues/255
		f = 30 * physic.MegaHertz
	}
	if f < 100*physic.Hertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	if bits != 8 {
		return nil, errors.New("d2xx: implement bits per word above 8")
	}

	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	s.c.noCS = m&spi.NoCS != 0
	s.c.halfDuplex = m&spi.HalfDuplex != 0
	s.c.lsbFirst = m&spi.LSBFirst != 0
	m &^= spi.NoCS | spi.HalfDuplex | spi.LSBFirst
	if s.c.halfDuplex {
		return nil, errors.New("d2xx: spi.HalfDuplex is not yet supported (implementing wouldn't be too hard, please submit a PR")
	}
	if m < 0 || m > 3 {
		return nil, errors.New("d2xx: unknown spi mode")
	}
	s.c.edgeInvert = m&1 != 0
	s.c.clkActiveLow = m&2 != 0
	if s.maxFreq == 0 || f < s.maxFreq {
		if _, err := s.c.f.h.mpsseClock(s.maxFreq); err != nil {
			return nil, err
		}
		s.maxFreq = f
	}
	// Note: D4~D7 are unusable. TODO(maruel): Keep them as-is when transmitting.
	const clk = byte(1) << 0
	const mosi = byte(1) << 1
	const miso = byte(1) << 2
	const cs = byte(1) << 3
	b := byte(0)
	if !s.c.noCS {
		b |= cs
	}
	if s.c.clkActiveLow {
		b |= clk
	}
	// D0, D1 and D3 are output.
	if err := s.c.f.h.mpsseDBus(mosi|clk|cs, b); err != nil {
		return nil, err
	}
	s.c.f.usingSPI = true
	return &s.c, nil
}

// LimitSpeed implements spi.Port.
func (s *spiMPSEEPort) LimitSpeed(f physic.Frequency) error {
	if f > physic.GigaHertz {
		return fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 30MHz", f)
	}
	if f > 30*physic.MegaHertz {
		f = 30 * physic.MegaHertz
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
	edgeInvert   bool // CPHA=1
	clkActiveLow bool // CPOL=1
	noCS         bool // CS line is not changed
	lsbFirst     bool // Default is MSB first
	halfDuplex   bool // 3 wire mode
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
		if p.KeepCS {
			return errors.New("d2xx: implement spi.Packet.KeepCS")
		}
		if p.BitsPerWord&7 != 0 {
			return errors.New("d2xx: bits must be a multiple of 8")
		}
		if p.BitsPerWord != 0 && p.BitsPerWord != 8 {
			return errors.New("d2xx: implement spi.Packet.BitsPerWord")
		}
		if err := verifyBuffers(p.W, p.R); err != nil {
			return err
		}
	}
	const clk = byte(1) << 0
	const mosi = byte(1) << 1
	const miso = byte(1) << 2
	const cs = byte(1) << 3
	// D0, D1 and D3 are output.
	const mask = mosi | clk | cs
	for _, p := range pkts {
		if len(p.W) == 0 && len(p.R) == 0 {
			continue
		}
		// TODO(maruel): s.halfDuplex.
		// TODO(maruel): Package as one big transaction?

		// Assert CS.
		b := byte(0)
		if s.clkActiveLow {
			b |= clk
		}
		if err := s.f.h.mpsseDBus(mask, b); err != nil {
			return err
		}

		ew := gpio.FallingEdge
		er := gpio.RisingEdge
		if s.edgeInvert {
			ew, er = er, ew
		}
		if s.clkActiveLow {
			// TODO(maruel): Not sure.
			ew, er = er, ew
		}
		if err := s.f.h.mpsseTx(p.W, p.R, ew, er, s.lsbFirst); err != nil {
			return err
		}

		// Deassert CS.
		b = byte(0)
		if !s.noCS {
			b |= cs
		}
		if s.clkActiveLow {
			b |= clk
		}
		if err := s.f.h.mpsseDBus(mask, b); err != nil {
			return err
		}
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
	s.maxFreq = 0
	s.c.edgeInvert = false
	s.c.clkActiveLow = false
	s.c.noCS = false
	s.c.lsbFirst = false
	s.c.halfDuplex = false
	s.c.f.mu.Unlock()
	return nil
}

func (s *spiSyncPort) String() string {
	return s.c.f.String()
}

const ft232rMaxSpeed = 3*physic.MegaHertz + 50*physic.KiloHertz

// Connect implements spi.Port.
func (s *spiSyncPort) Connect(f physic.Frequency, m spi.Mode, bits int) (spi.Conn, error) {
	if f > physic.GigaHertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 1.5MHz", f)
	}
	if f > ft232rMaxSpeed/2 {
		// TODO(maruel): Figure out a way to communicate that the speed was lowered.
		// https://github.com/google/periph/issues/255
		f = ft232rMaxSpeed / 2
	}
	if f < 100*physic.Hertz {
		return nil, fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	if bits&7 != 0 {
		return nil, errors.New("d2xx: bits must be multiple of 8")
	}
	if bits != 8 {
		return nil, errors.New("d2xx: implement bits per word above 8")
	}

	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	s.c.noCS = m&spi.NoCS != 0
	s.c.halfDuplex = m&spi.HalfDuplex != 0
	s.c.lsbFirst = m&spi.LSBFirst != 0
	m &^= spi.NoCS | spi.HalfDuplex | spi.LSBFirst
	if s.c.halfDuplex {
		return nil, errors.New("d2xx: spi.HalfDuplex is not yet supported (implementing wouldn't be too hard, please submit a PR")
	}
	if m < 0 || m > 3 {
		return nil, errors.New("d2xx: unknown spi mode")
	}
	s.c.edgeInvert = m&1 != 0
	s.c.clkActiveLow = m&2 != 0
	if s.maxFreq == 0 || f < s.maxFreq {
		if err := s.c.f.SetSpeed(f * 2); err != nil {
			return nil, err
		}
		s.maxFreq = f
	}
	// D0, D2 and D3 are output. D4~D7 are kept as-is.
	const mosi = byte(1) << 0 // TX
	const miso = byte(1) << 1 // RX
	const clk = byte(1) << 2  // RTS
	const cs = byte(1) << 3   // CTS
	mask := mosi | clk | cs | (s.c.f.dmask & 0xF0)
	if err := s.c.f.setDBusMaskLocked(mask); err != nil {
		return nil, err
	}
	// TODO(maruel): Combine both following calls if possible. We'd shave off a
	// few ms.
	if !s.c.noCS {
		// CTS/SPI_CS is active low.
		if err := s.c.f.dbusSyncGPIOOutLocked(3, gpio.High); err != nil {
			return nil, err
		}
	}
	if s.c.clkActiveLow {
		// RTS/SPI_CLK is active low.
		if err := s.c.f.dbusSyncGPIOOutLocked(2, gpio.High); err != nil {
			return nil, err
		}
	}
	s.c.f.usingSPI = true
	return &s.c, nil
}

// LimitSpeed implements spi.Port.
func (s *spiSyncPort) LimitSpeed(f physic.Frequency) error {
	if f > physic.GigaHertz {
		return fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 1.5MHz", f)
	}
	if f < 100*physic.Hertz {
		return fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	s.c.f.mu.Lock()
	defer s.c.f.mu.Unlock()
	if s.maxFreq != 0 && s.maxFreq <= f {
		return nil
	}
	if err := s.c.f.SetSpeed(f * 2); err == nil {
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
	edgeInvert   bool // CPHA=1
	clkActiveLow bool // CPOL=1
	noCS         bool // CS line is not changed
	lsbFirst     bool // Default is MSB first
	halfDuplex   bool // 3 wire mode
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
	// We need to 'expand' each bit 2 times * 8 bits, which leads
	// to a 16x memory usage increase. Adds 5 samples before and after.
	totalW := 0
	totalR := 0
	for _, p := range pkts {
		if p.KeepCS {
			return errors.New("d2xx: implement spi.Packet.KeepCS")
		}
		if p.BitsPerWord&7 != 0 {
			return errors.New("d2xx: bits must be a multiple of 8")
		}
		if p.BitsPerWord != 0 && p.BitsPerWord != 8 {
			return errors.New("d2xx: implement spi.Packet.BitsPerWord")
		}
		if err := verifyBuffers(p.W, p.R); err != nil {
			return err
		}
		// TODO(maruel): Correctly calculate offsets.
		if len(p.W) != 0 {
			totalW += 2 * 8 * len(p.W)
		}
		if len(p.R) != 0 {
			totalR += 2 * 8 * len(p.R)
		}
	}

	// Create a large, single chunk.
	var we, re []byte
	if totalW != 0 {
		totalW += 10
		we = make([]byte, 0, totalW)
	}
	if totalR != 0 {
		totalR += 10
		re = make([]byte, totalR)
	}
	const mosi = byte(1) << 0 // TX
	const miso = byte(1) << 1 // RX
	const clk = byte(1) << 2  // RTS
	const cs = byte(1) << 3   // CTS

	s.f.mu.Lock()
	defer s.f.mu.Unlock()

	// https://en.wikipedia.org/wiki/Serial_Peripheral_Interface#Data_transmission

	csActive := s.f.dvalue & s.f.dmask & 0xF0
	csIdle := csActive
	if !s.noCS {
		csIdle = csActive | cs
	}
	clkIdle := csActive
	clkActive := clkIdle | clk
	if s.clkActiveLow {
		clkActive, clkIdle = clkIdle, clkActive
		csIdle |= clk
	}
	// Start of tx; assert CS if needed.
	we = append(we, csIdle, clkIdle, clkIdle, clkIdle, clkIdle)
	for _, p := range pkts {
		if len(p.W) == 0 && len(p.R) == 0 {
			continue
		}
		// TODO(maruel): s.halfDuplex.
		for _, b := range p.W {
			for j := uint(0); j < 8; j++ {
				// For each bit, handle clock phase and data phase.
				bit := byte(0)
				if !s.lsbFirst {
					// MSBF
					if b&(0x80>>j) != 0 {
						bit = mosi
					}
				} else {
					// LSBF
					if b&(1<<j) != 0 {
						bit = mosi
					}
				}
				if !s.edgeInvert {
					// Mode0/2; CPHA=0
					we = append(we, clkIdle|bit, clkActive|bit)
				} else {
					// Mode1/3; CPHA=1
					we = append(we, clkActive|bit, clkIdle|bit)
				}
			}
		}
	}
	// End of tx; deassert CS.
	we = append(we, clkIdle, clkIdle, clkIdle, clkIdle, csIdle)

	if err := s.f.txLocked(we, re); err != nil {
		return err
	}

	// Extract data from re into r.
	for _, p := range pkts {
		// TODO(maruel): Correctly calculate offsets.
		if len(p.W) == 0 && len(p.R) == 0 {
			continue
		}
		// TODO(maruel): halfDuplex.
		for i := range p.R {
			// For each bit, read at the right data phase.
			b := byte(0)
			for j := 0; j < 8; j++ {
				if re[5+i*8*2+j*2+1]&byte(1)<<1 != 0 {
					if !s.lsbFirst {
						// MSBF
						b |= 0x80 >> uint(j)
					} else {
						// LSBF
						b |= 1 << uint(j)
					}
				}
			}
			p.R[i] = b
		}
	}
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
		// TODO(maruel): Remove, this is not a problem.
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
