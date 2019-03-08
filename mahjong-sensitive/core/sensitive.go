package core

import (
	"os"

	filter "github.com/fwhappy/go-dirtyfilter"
	"github.com/fwhappy/go-dirtyfilter/store"
)

// FilterManager 过滤器
var FilterManager *filter.DirtyManager

// InitFilter 初始化过滤器
func InitFilter(cfgFile string) {
	f, err := os.Open(cfgFile)
	if err != nil {
		Logger.Warn("加载屏蔽词库失败:%v,err:%v", cfgFile, err.Error())
	}
	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		Reader: f,
	})
	if err != nil {
		Logger.Warn("解析屏蔽词库失败:%v,err:%v", cfgFile, err.Error())
	}
	FilterManager = filter.NewDirtyManager(memStore)
}
