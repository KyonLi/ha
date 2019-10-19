package main

import (
	"flag"
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"ha/adb"
	"log"
	"os"
	"path/filepath"
)

var (
	help       = flag.Bool("h", false, "This help.")
	remoteIP   = flag.String("r", "", "Wireless ADB IP address.")
	remotePort = flag.String("p", "5555", "Wireless ADB port.")

	target adb.Target
)

func storagePath() string {
	basePath := "."
	ex, err := os.Executable()
	if err == nil {
		basePath = filepath.Dir(ex)
	}
	return basePath + "/data"
}

func addInputSource(t *accessory.Television, id int, name string, inputSourceType int) {
	in := service.NewInputSource()

	in.Identifier.SetValue(id)
	in.ConfiguredName.SetValue(name)
	in.Name.SetValue(name)
	in.InputSourceType.SetValue(inputSourceType)
	in.IsConfigured.SetValue(characteristic.IsConfiguredConfigured)

	t.AddService(in.Service)
	t.UpdateIDs()
	t.Television.AddLinkedService(in.Service)

	in.ConfiguredName.OnValueRemoteUpdate(func(str string) {
		fmt.Printf(" %s configured name => %s\n", name, str)
	})
	in.InputSourceType.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s source type => %v\n", name, v)
	})
	in.IsConfigured.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s configured => %v\n", name, v)
	})
	in.CurrentVisibilityState.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s current visibility => %v\n", name, v)
	})
	in.Identifier.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s identifier => %v\n", name, v)
	})
	in.InputDeviceType.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s device type => %v\n", name, v)
	})
	in.TargetVisibilityState.OnValueRemoteUpdate(func(v int) {
		fmt.Printf(" %s target visibility => %v\n", name, v)
	})
	in.Name.OnValueRemoteUpdate(func(str string) {
		fmt.Printf(" %s name => %s\n", name, str)
	})
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *remoteIP == "" {
		log.New(os.Stderr, "", 0).Fatal("invalid adb ip address")
	}

	target = adb.NewTarget(adb.Config{
		IP:   *remoteIP,
		Port: *remotePort,
	})

	active := characteristic.ActiveInactive
	if awake, _ := target.IsAwake(); awake {
		active = characteristic.ActiveActive
	}

	info := accessory.Info{
		Name: "TV",
	}
	acc := accessory.NewTelevision(info)

	acc.Television.Active.SetValue(active)
	acc.Television.ActiveIdentifier.SetValue(1)
	acc.Television.SleepDiscoveryMode.SetValue(characteristic.SleepDiscoveryModeAlwaysDiscoverable)
	acc.Television.CurrentMediaState.SetValue(characteristic.CurrentMediaStatePause)
	acc.Television.TargetMediaState.SetValue(characteristic.TargetMediaStatePause)

	acc.Television.Active.OnValueRemoteUpdate(func(v int) {
		switch v {
		case characteristic.ActiveActive:
			if err := target.SendKey(adb.KEYCODE_WAKEUP); err == nil {
				acc.Television.Active.SetValue(characteristic.ActiveActive)
			}
		case characteristic.ActiveInactive:
			if err := target.SendKey(adb.KEYCODE_SLEEP); err == nil {
				acc.Television.Active.SetValue(characteristic.ActiveInactive)
			}
		}
	})

	acc.Television.ActiveIdentifier.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("active identifier => %d\n", v)
	})

	acc.Television.ConfiguredName.OnValueRemoteUpdate(func(v string) {
		acc.Television.ConfiguredName.SetValue(v)
	})
	acc.Television.SleepDiscoveryMode.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("sleep discovery mode => %d\n", v)
	})
	acc.Television.Brightness.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("brightness => %d\n", v)
	})
	acc.Television.ClosedCaptions.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("closed captions => %d\n", v)
	})
	acc.Television.DisplayOrder.OnValueRemoteUpdate(func(v []byte) {
		fmt.Printf("display order => %v\n", v)
	})
	acc.Television.CurrentMediaState.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("current media state => %d\n", v)
	})
	acc.Television.TargetMediaState.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("target media state => %d\n", v)
	})

	acc.Television.PowerModeSelection.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("power mode selection => %d\n", v)
	})

	acc.Television.PictureMode.OnValueRemoteUpdate(func(v int) {
		fmt.Printf("PictureMode => %d\n", v)
	})

	acc.Television.RemoteKey.OnValueRemoteUpdate(func(v int) {
		switch v {
		case characteristic.RemoteKeyRewind:
			fmt.Println("Rewind")
		case characteristic.RemoteKeyFastForward:
			fmt.Println("Fast forward")
		case characteristic.RemoteKeyExit:
			_ = target.SendKey(adb.KEYCODE_HOME)
		case characteristic.RemoteKeyPlayPause:
			fmt.Println("Play/Pause")
		case characteristic.RemoteKeyInfo:
			_ = target.SendKey(adb.KEYCODE_MENU)
		case characteristic.RemoteKeyNextTrack:
			fmt.Println("Next")
		case characteristic.RemoteKeyPrevTrack:
			fmt.Println("Prev")
		case characteristic.RemoteKeyArrowUp:
			_ = target.SendKey(adb.KEYCODE_DPAD_UP)
		case characteristic.RemoteKeyArrowDown:
			_ = target.SendKey(adb.KEYCODE_DPAD_DOWN)
		case characteristic.RemoteKeyArrowLeft:
			_ = target.SendKey(adb.KEYCODE_DPAD_LEFT)
		case characteristic.RemoteKeyArrowRight:
			_ = target.SendKey(adb.KEYCODE_DPAD_RIGHT)
		case characteristic.RemoteKeySelect:
			_ = target.SendKey(adb.KEYCODE_DPAD_CENTER)
		case characteristic.RemoteKeyBack:
			_ = target.SendKey(adb.KEYCODE_BACK)
		}
	})

	config := hc.Config{Pin: "11451419", StoragePath: storagePath()}
	t, err := hc.NewIPTransport(config, acc.Accessory)

	addInputSource(acc, 1, "Home Screen", characteristic.InputSourceTypeHomeScreen)

	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
