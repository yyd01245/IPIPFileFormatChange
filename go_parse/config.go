package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

//定义配置文件解析后的结构
type CountryConfig struct {
	XAsia 			 []string `json:"XA"`
	XEurope 		 []string `json:"XE"`
	XAmerica  	 []string `json:"XS"`
	XIndia  		 []string `json:"XI"`
	XAfrica 		 []string `json:"XF"`
}

var config CountryConfig = CountryConfig{}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
    return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
		//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
		fmt.Printf("---- load file  %v\n",filename)		
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return
    }

    //读取的数据为json格式，需要进行解码
		err = json.Unmarshal(data, v)
		fmt.Printf("---- load %v\n",v)
    if err != nil {
			fmt.Printf("---- load error %v\n",err)
        return
    }
		fmt.Printf("---- load success %v\n",v)
}

func GetJsonConfig (path string ) error{
	JsonParse := NewJsonStruct()

	//下面使用的是相对路径，config.json文件和main.go文件处于同一目录下
	JsonParse.Load(path, &config)

	return nil
}
