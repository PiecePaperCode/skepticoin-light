package main

import (
	"fmt"
)

func ip(ip [16]byte, port uint16) string {
	interfaceIp := make([]interface{}, 16)
	for i, b := range ip {
		interfaceIp[i] = b
	}
	IPv6 := false
	for _, b := range ip[0:12] {
		if b != 0 {
			IPv6 = true
			break
		}
	}
	if IPv6 {
		return fmt.Sprintf(
			// example Output [00:00:00:00:00:00:1758:6343]:2412
			"%s:%d",
			fmt.Sprintf(
				"[%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x]",
				interfaceIp...,
			),
			port,
		)
	} else {
		return fmt.Sprintf(
			// example Output 192.168.1.123:2412
			"%s:%d",
			fmt.Sprintf(
				"%d.%d.%d.%d",
				interfaceIp[12:]...,
			),
			port,
		)
	}
}
