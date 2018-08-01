package filters

import (
	"encoding/binary"
	"encoding/hex"
	"sort"
)

var filterList sort.IntSlice

func init() {
	// Porn
	Add("036a792df9df5c5ab22004c2157747d7cd9b7b0137946a29a8a8f11840214edc") // title: Lucy Pinder - Back at the Pool
	Add("076641ebb019d951b21f9cab5c1f3fd52777e7c11fd3a5b5767e479c3d8e3c42") // title: Hidden Camera in my Sister's Room
	Add("3f754d3ec1ee906fd32e888b63668b1281d9329ec3302a6dacc5b16434d881bd") // title: Byronnnna Masturbates
	Add("6ed1eaaef833f8d33e2363c2a98aa01bb97b77c868d85d3cbc28182b9663343c") // title: ebony oatmeal creampie
	Add("fd8010a13fc7a04713b4693cd7886813cdd55bb5c82a0291ee14735d92cd1fcf") // title: Whore Slave

	// DMCA
	Add("465d4d2af3744cecd6aaedc414768bc00be6c6314e6a3f84f52a5a72321c04bd") // title: The Good the Bad and the Ugly
	Add("cf129a5565caa396a5b932a8054f28da382f9dae61f0e308af2db039ff398d02") // title: Wonder.Woman
	Add("b9cb9197233180d06ab91b482eed6512c3fe23cffa2972b9a56372711810fe4b") // title: Serenity
	Add("38bbded1efcfa73a72c1b37dba463c05ff171b6a31e5a6a875e13846e6225e4e") // title: Spider

}

func Add(txid string) {
	if len(txid) < 8 {
		panic("txid too short")
	}
	b, err := hex.DecodeString(txid[0:8])
	if err != nil {
		panic("invalid hex")
	}
	i := binary.BigEndian.Uint32(b)
	insert(&filterList, int(i))
}

func Contains(txid string) bool {
	if len(txid) < 8 {
		panic("txid too short")
	}
	b, err := hex.DecodeString(txid[0:8])
	if err != nil {
		panic("invalid hex")
	}
	ui := binary.BigEndian.Uint32(b)
	s := filterList.Search(int(ui))
	return s < len(filterList) && ([]int(filterList))[s] == int(ui)
}

func Clear() {
	filterList = sort.IntSlice{}
}

func insert(list *sort.IntSlice, i int) {
	if len(*list) == 0 {
		*list = sort.IntSlice{i}
		return
	}
	location := list.Search(i)
	if location >= len(*list) {
		*list = append(*list, i)
	} else if []int(*list)[location] != i {
		*list = append(*list, 0)
		copy([]int(*list)[location+1:], []int(*list)[location:])
		([]int(*list))[location] = i
	}
}
