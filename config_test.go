package pcp

import (
	"fmt"
	"testing"
)

func TestFunc(t *testing.T) {
	fmt.Println("detected PCP root:", RootPath)
	fmt.Println("detected PCP config file:", ConfPath)
	fmt.Println("Config options obtained:")
	for k, v := range Config {
		fmt.Println(k, ":", v)
	}
}
