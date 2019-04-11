package main

import  "cy/game/util/tools/FilterLib"

func main()  {
	FilterLib := FilterLib.NewFilterLib([]string{"大爷","ox"})
	s := FilterLib.Replace("你大爷的oxox")
	println("过滤后的字符 ggo ",s)
	is := FilterLib.CheckFilter("你小爷的ox")
	println("检测字符串是否包含敏感字 ",is)
}
