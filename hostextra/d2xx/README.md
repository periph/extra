# d2xx

Go driver wrapper for the [Future Technology "D2XX" driver](
http://www.ftdichip.com/Drivers/D2XX.htm).

## Installation

### Debian

This includes Raspbian and Ubuntu.

- Configure cgo as explained at
  [extra/README.md](https://github.com/periph/extra/tree/master/README.md#debian).


#### Temporary

Run this command **after** connecting your FTDI device:

```
sudo rmmod ftdi_sio usbserial
```


#### Permanent

Reconnect your device **after** running the following command:

```
cd $GOPATH/src/periph.io/x/extra/hostextra/d2xx
sudo cp debian/98-ft232h.rules /etc/udev/rules.d/
sudo udevadm control --reload-rules
sudo udevadm trigger --verbose
```


### macOS

- Configure cgo as explained at
  [extra/README.md](https://github.com/periph/extra/tree/master/README.md#macos).


#### Temporary

```
sudo kextunload -b com.apple.driver.AppleUSBFTDI
```


#### Permanently

TODO.


### Windows

- Once you connect the device, Windows Update should install the FTDI D2XX
  driver automatically. Wait for it to occur.
- If this fails, install the driver from
  http://www.ftdichip.com/Drivers/D2XX.htm

## Included driver license

- ftd2xx.h is v2.12.28
- WinTypes.h is v1.4.6
- darwin_amd64/libftd2xx.a v1.4.4
- linux_arm/libftd2xx.a v1.4.6 with ARMv6 hard float (RPi compatible)
- linux_amd64/libftd2xx.a v1.4.6

> This software is provided by Future Technology Devices International Limited
> ``as is'' and any express or implied warranties, including, but not limited
> to, the implied warranties of merchantability and fitness for a particular
> purpose are disclaimed. In no event shall future technology devices
> international limited be liable for any direct, indirect, incidental, special,
> exemplary, or consequential damages (including, but not limited to,
> procurement of substitute goods or services; loss of use, data, or profits; or
> business interruption) however caused and on any theory of liability, whether
> in contract, strict liability, or tort (including negligence or otherwise)
> arising in any way out of the use of this software, even if advised of the
> possibility of such damage.  FTDI drivers may be used only in conjunction with
> products based on FTDI parts.
>
> FTDI drivers may be distributed in any form as long as license information is
> not modified.
>
> If a custom vendor ID and/or product ID or description string are used, it is
> the responsibility of the product manufacturer to maintain any changes and
> subsequent WHCK re-certification as a result of making these changes.
>
> For more detail on FTDI Chip Driver licence terms, please [click
> here](http://www.ftdichip.com/Drivers/FTDriverLicenceTermsSummary.htm).


### Modifications

- Fixed ftd2xx.h to UTF-8
- Converted header files from CRLF to LF
- Removed trailing spaces and tabs
