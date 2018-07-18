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
	for checkType, checkSum := range data["info_sum"].(map[string]interface{}) {
		sum.Info[checkType] += int64(checkSum.(float64))
	}

	for checkType, checkSum := range data["block_sum"].(map[string]interface{}) {
		sum.Block[checkType] += int64(checkSum.(float64))
	}

	for checkType, checkSum := range data["hook_sum"].(map[string]interface{}) {
		sum.Hook[checkType] += int64(checkSum.(float64))
	}

	sum.Request += int64(data["request_sum"].(float64))
}

func NewSumData() *SumData {
	return &SumData{
		Block: make(map[string]int64),
		Info:  make(map[string]int64),
		Hook:  make(map[string]int64),
	}
}
