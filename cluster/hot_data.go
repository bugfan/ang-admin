package cluster

import (
	"encoding/json"
	"fmt"
)

func Put(key string, value interface{}) {
	data, _ := json.Marshal(value)
	fmt.Println("hot data:", key, string(data))
}
