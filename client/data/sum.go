package data

import . "openrasp-cloud/proto"

var Sum = make(chan map[string]*SumData, 1)
var RaspBasicData = make(map[string]*Rasp)

func init() {
	Sum <- make(map[string]*SumData)
}

func AddData(id string, data map[string]interface{}) {
	sumData := <-Sum
	defer func() { Sum <- sumData }()

	if sumData[id] == nil {
		sumData[id] = NewSumData()
	}

	sum := sumData[id]
	for checkType, checkSum := range data["info_sum"].(map[string]int64) {
		sum.Info[checkType] += checkSum
	}

	for checkType, checkSum := range data["block_sum"].(map[string]int64) {
		sum.Info[checkType] += checkSum
	}

	sum.Hook += data["hook_sum"].(int64)
	sum.Request += data["request_sum"].(int64)
}

func NewSumData() *SumData {
	return &SumData{
		Block: make(map[string]int64),
		Info:  make(map[string]int64),
	}
}
