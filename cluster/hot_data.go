package cluster

import (
	"encoding/json"
	"fmt"
)

func Put(key string, value interface{}) {
	data, _ := json.Marshal(value)
	fmt.Println("hot data:", key, string(data))
}

func PrintFullServerConfig(cfg interface{}) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("=== Overall server.json Error: %v ===\n", err)
		return
	}
	fmt.Println("\n=================== OVERALL ANG SERVER.JSON CONFIG ===================")
	fmt.Println(string(data))
	fmt.Println("======================================================================\n")
}

func PrintFullTunnelConfig(cfg interface{}) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("=== Overall tunnel.json Error: %v ===\n", err)
		return
	}
	fmt.Println("\n=================== OVERALL ANG TUNNEL.JSON CONFIG ===================")
	fmt.Println(string(data))
	fmt.Println("======================================================================\n")
}

func PrintFullCertificateConfig(cfg interface{}) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("=== Overall certificate.json Error: %v ===\n", err)
		return
	}
	fmt.Println("\n=================== OVERALL ANG CERTIFICATE.JSON CONFIG ===================")
	fmt.Println(string(data))
	fmt.Println("===========================================================================\n")
}

func PrintFullUserConfig(cfg interface{}) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("=== Overall user.json Error: %v ===\n", err)
		return
	}
	fmt.Println("\n=================== OVERALL ANG USER.JSON CONFIG ===================")
	fmt.Println(string(data))
	fmt.Println("====================================================================\n")
}

func PrintFullGroupConfig(cfg interface{}) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("=== Overall group.json Error: %v ===\n", err)
		return
	}
	fmt.Println("\n=================== OVERALL ANG GROUP.JSON CONFIG ===================")
	fmt.Println(string(data))
	fmt.Println("=====================================================================\n")
}

