package http_service

import (
	"fmt"
	"story-monitor/types"
	"testing"
)

func TestHttpService(t *testing.T) {
	_, err := GetValPerformance(1029700, []*types.MonitorObj{})
	if err != nil {
		fmt.Println("http get block data err:", err)
	}

}
func TestGetLatestBlock(t *testing.T) {
	fmt.Println(GetLatestBlock())
}
