package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"time"
)

/*
I'm not 100% sure about this schema but this seems to work fine for my Wiz bulb.
Sourced from: https://github.com/sbidy/pywizlight/blob/master/README.md#bulb-methods-udp-native
{ "method": "getPilot" | "getUserConfig" | "getSystemConfig" | "getWifiConfig" | "reboot" | "getDevInfo" } |
{ "method": "pulse", "params": { "delta": int, "duration": int } } |
{
	"method": "setPilot",
	"params": { "state": true | false, dimming: 10-100 } &
	(
		{ "r": 0-255, "g": 0-255, "b": 0-255, "w": 0-255, "c": 0-255 } |
		{ "temp": 2200-6500 } |
		{ "sceneId": int, "speed": 10-200 } |
	)
}

TODO: integrate notifications
TODO: dsl for writing scenes
TODO: figure out how to get current lighting state
TODO: map the color space
*/

func main() {
	if len(os.Args) <= 1 {
		panic("missing wiz IP address")
	}

	wizIP := net.ParseIP(os.Args[1]).String()
	if wizIP == "<nil>" {
		panic("invalid IP address")
	}

	addr := net.JoinHostPort(wizIP, "38899")

	test(addr)
	// demo(addr)
}

func test(addr string) {
	command := `
{"method":"getWifiConfig", "params":{"delta":-50,"duration":30}}
	`

	resp, err := send(addr, command)
	if err != nil {
		panic(err)
	}
	prettyPrint(resp)
}

func demo(addr string) {
	for i := 0; i < 10000; i++ {
		i := i % 360
		r, g, b := hsl2rgb(i, 100, 50)

		fmt.Printf("HSL(%d, 100, 50) -> RGB(%d, %d, %d)\n", i, r, g, b)

		command := fmt.Sprintf(`
		{
			"method": "setPilot",
			"params": {
				"r": %d,
				"g": %d,
				"b": %d
			}
		}
		`, r, g, b)

		_, err := send(addr, command)
		if err != nil {
			panic(err)
		}
		// prettyPrint(resp)

		time.Sleep(125 * time.Millisecond)
	}
}

// h ∈ [0°, 360°), s ∈ [0, 100], l ∈ [0, 100]
// https://en.wikipedia.org/wiki/HSL_and_HSV
func hsl2rgb(hi, si, li int) (int, int, int) {
	h := float64(hi) / 60
	s := float64(si) / 100
	l := float64(li) / 100

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h, 2.0)-1))

	var r, g, b float64
	switch {
	case 0 <= h && h < 1:
		r, g, b = c, x, 0
	case 1 <= h && h < 2:
		r, g, b = x, c, 0
	case 2 <= h && h < 3:
		r, g, b = 0, c, x
	case 3 <= h && h < 4:
		r, g, b = 0, x, c
	case 4 <= h && h < 5:
		r, g, b = x, 0, c
	case 5 <= h && h < 6:
		r, g, b = c, 0, x
	}

	m := l - c/2

	rr := int((r + m) * 255)
	gg := int((g + m) * 255)
	bb := int((b + m) * 255)

	return rr, gg, bb
}

func send(addr, msg string) ([]byte, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func prettyPrint(jsonData []byte) {
	var pretty bytes.Buffer
	json.Indent(&pretty, jsonData, "", "\t")
	fmt.Println(pretty.String())
}

const (
	turnOn     = `{"method": "setPilot", "params":{"state": true}}`
	turnOff    = `{"method": "setPilot", "params":{"state": false}}`
	getDetails = `{"method": "getPilot", "params":{}}`
)
