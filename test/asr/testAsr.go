package main

import (
	"fmt"

	"github.com/henryleu/vads/hly/util"
)

func main() {
	//asr_result := util.AsrByAicyber("/mnt/3.wav")
	asr_result := util.AsrClient("/mnt/3.wav")
	fmt.Printf("11111111111111111:%s ", asr_result)
}
