package setting

import (
	"testing"
)

func TestSetPostion(t *testing.T) {
	ms := NewMSetting()
	ms.SetSetting([]int{1, 1, 1})
	ms.SetPositionValue(1, 2)
	ms.SetPositionValue(2, 3)
	ms.SetPositionValue(5, 108)
	t.Logf("setting:%v", ms.setting)
}
