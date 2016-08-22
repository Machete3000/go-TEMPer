  package main
/*
 * by Machete3000 (c) 2016 github.com/Machete3000
 * based on kylelemons/gousb/lsusb/main.go Copyright 2013 Google Inc.
 * based on Temper.go by Thomas Burke (c) 2015 (tburke@tb99.com)
 * based on pcsensor.c by Juan Carlos Perez (c) 2011 (cray@isp-sl.com)
 * based on Temper.c by Robert Kavaler (c) 2009 (relavak.com)
*/

import (
  "flag"
	"log"
	"fmt"
	
  "github.com/truveris/gousb/usb"
	"github.com/truveris/gousb/usbid"
  
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

func temperature() (float64, error) {
    
	ctx := usb.NewContext()
	defer ctx.Close()
 
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		return desc.Vendor == 0x0c45 && desc.Product == 0x7401
	})

	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
  
	if err != nil {
    log.Printf("ListDevices failed");
		return 0.0, err
	}

	if len(devs) == 0 {
		return 0.0, fmt.Errorf("No thermometers found.")
	}

	dev := devs[0]
 
  if err = dev.DetachKernelDriver(0); err != nil {
     // Keep going...
  }
 
	if err = dev.SetConfig(1); err != nil {
		return 0.0, err
	}

	ep, err := dev.OpenEndpoint(1, 1, 0, 0x82)
	if err != nil {
		return 0.0, err
	}
	
  if _, err = dev.Control(0x21, 0x09, 0x0200, 0x01, []byte{0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00}); err != nil {
		return 0.0, err
	}
	
  buf := make([]byte, 8)
	
  if _, err = ep.Read(buf); err != nil {
  		return 0.0, err
	}
	  
  return float64(buf[2]) + float64(buf[3])/256, nil
}

func listDevices() {
  flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	// Debugging can be turned on; this shows some of the inner workings of the libusb package.
	ctx.Debug(*debug)

	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		// The usbid package can be used to print out human readable information.
		fmt.Printf("%03d.%03d %s:%s %s\n", desc.Bus, desc.Address, desc.Vendor, desc.Product, usbid.Describe(desc))
		fmt.Printf("  Protocol: %s\n", usbid.Classify(desc))

		// The configurations can be examined from the Descriptor, though they can only
		// be set once the device is opened.  All configuration references must be closed,
		// to free up the memory in libusb.
		for _, cfg := range desc.Configs {
			// This loop just uses more of the built-in and usbid pretty printing to list
			// the USB devices.
			fmt.Printf("  %s:\n", cfg)
			for _, alt := range cfg.Interfaces {
				fmt.Printf("    --------------\n")
				for _, iface := range alt.Setups {
					fmt.Printf("    %s\n", iface)
					fmt.Printf("      %s\n", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						fmt.Printf("      %s\n", end)
					}
				}
			}
			fmt.Printf("    --------------\n")
		}

		// After inspecting the descriptor, return true or false depending on whether
		// the device is "interesting" or not.  Any descriptor for which true is returned
		// opens a Device which is retuned in a slice (and must be subsequently closed).
		return false
	})

	// All Devices returned from ListDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// ListDevices can occaionally fail, so be sure to check its return value.
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	for _, dev := range devs {
		// Once the device has been selected from ListDevices, it is opened
		// and can be interacted with.
		_ = dev
	}
}



func main() {
  //listDevices()
	c, err := temperature()
	if err == nil {
		log.Printf("Temperature: %.2fF %.2fC\n", 9.0/5.0*c+32, c)
	} else {
		log.Fatalf("Failed: %s", err)
	}
}
