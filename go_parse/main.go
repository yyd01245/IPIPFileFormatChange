package main

import (
	"fmt"
	"strings"
	// "strconv"
	// "./csv"
	"os"
	log "github.com/Sirupsen/logrus"
	cm "github.com/yyd01245/go_common/common"
	"github.com/yyd01245/go_common/csv"
)

const PIDFILE = "/tmp/upwan.pid"

type IPIPExchange struct {
	OutFile  string
	InFile string
	Filter []string
}
func NewExchange(out string ,in string, filter string) *IPIPExchange{
	app := new(IPIPExchange)
	app.OutFile = out
	app.InFile = in
	// app.Filter = append(app.Filter,"BGP,CN")
	filterLine := strings.Split(filter,"&")
	for _,v := range filterLine {
		if v == "" {
			continue
		}
		txt := v + ",CN"
		// if v == "ctn" {
		// 	// 提取 ctn 和 BGP 两个的数据转换
		// 	txt = "ctn,CN"
		// }
		// if v == "cun" {
		// 	// 提取 cun 和 BGP 两个的数据转换		
		// 	txt = "cun,CN"
		// }
		// if v == "cmn" {
		// 	// 提取 cmn 和 BGP 两个的数据转换
		// 	txt = "cmn,CN"
		// }
		if txt != "" {
			app.Filter = append(app.Filter,txt)
		}
	}
	log.Infof("---app: %v",app)
	return app
}

func (this *IPIPExchange) Exchange() {
	// 读取文件
	// if cm.CheckFileIsExist(this.InFile) == false {
	// 	log.Errorf("file: %s is not exist!!!",this.InFile)
	// }
	body := cm.GetFileValue(this.InFile) 
	if body == nil {
		log.Errorf("file: %s open failed !!!",this.InFile)
		return 
	}
	// 创建输出文件
	firstLine := []string{"network","geoname_id","registered_country_geoname_id",
			"represented_country_geoname_id","is_anonymous_proxy","is_satellite_provider"}
	file, err := os.OpenFile(this.OutFile,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write(firstLine)
	if err != nil {
		log.Errorf("write first line to %v, firstline:%v",this.OutFile,firstLine)
		return 
	}
	bak := this.OutFile+"cidr"
	// 复制一份cidr 拷贝项，用于核实结果
	fileCidr, err := os.OpenFile(bak,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer fileCidr.Close()
	
	stringBody := string(body)
	// 获取一行行数据
	outLine := strings.Split(stringBody,"\n")
	total := 0
	for _,v := range outLine {
		// 行数据
		flag := false
		v = strings.Replace(v,"\n","",-1)
		for _,d := range this.Filter {
			// 匹配过滤项
			index := strings.Index(v,d) 
			if index > 0 {
				// 命中
				flag = true
				break
			}
		}
		if flag {
			// 修改内容
			total++
			data := strings.Split(v,",")
			ip := data[0]
			// 1.0.1.0/24,1814991,1814991,,0,0
			// 判断是否是cidr格式
			if cm.CheckPrivateIPValid(ip) == false {
				log.Errorf("get ip is not cidr format: %v",ip)
				continue
			}
			startIP,endIP,_ := cm.GetCidrIpRange(ip)
			if startIP == "" || endIP == "" {
				log.Errorf("get ip cidr format: %v, startIP=%v,endIP=%v!",ip,startIP,endIP )
				continue
			}
			// start_local_id := cm.InetAtoN(startIP)
			// end_local_id := cm.InetAtoN(endIP)

			// out := ip + ","+opts.GeonameID + "," + opts.GeonameID + ",,0,0\n"
			// 输出
			outCsv := []string{"\"%s\"","\"%s\"","\"%d\"","\"%d\"","\"%s\"","\"%s\""}
			outCsv[0] = ip
			outCsv[1] = opts.GeonameID
			outCsv[2] = opts.GeonameID
			outCsv[3] = ""
			outCsv[4] = "0"
			outCsv[5] = "0"
			log.Infof("--- %v --- ",outCsv)
			err := csvWriter.Write(outCsv)
			csvWriter.Flush()
			// out := ip + ","+opts.GeonameID + "," + opts.GeonameID + ",,0,0\n"
			// 输出
			// _,err = file.WriteString(out)
			if err != nil {
				log.Errorf("write line to %v, line:%v",this.OutFile,outCsv) 
			}
			_,err = fileCidr.WriteString(ip+"\n")
			if err != nil {
				log.Errorf("write line to %v, line:%v",bak,ip) 
			}
		}
		
	}
	csvWriter.Flush()
	log.Infof("---- total: len = %v",total)
}

func (this *IPIPExchange) Exchange2X() {
	// 读取文件
	// if cm.CheckFileIsExist(this.InFile) == false {
	// 	log.Errorf("file: %s is not exist!!!",this.InFile)
	// }
	body := cm.GetFileValue(this.InFile) 
	if body == nil {
		log.Errorf("file: %s open failed !!!",this.InFile)
		return 
	}
	// 创建输出文件
	file, err := os.OpenFile(this.OutFile,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)

	bak := this.OutFile+"cidr"
	// 复制一份cidr 拷贝项，用于核实结果
	fileCidr, err := os.OpenFile(bak,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer fileCidr.Close()

	// start_local_id,_ := strconv.Atoi(opts.GeonameID)
	stringBody := string(body)
	// 获取一行行数据
	outLine := strings.Split(stringBody,"\n")
	total := 0
	for _,v := range outLine {
		// 行数据
		flag := false
		v = strings.Replace(v,"\n","",-1)
		for _,d := range this.Filter {
			// 匹配过滤项
			index := strings.Index(v,d) 
			if index > 0 {
				// 命中
				flag = true
				break
			}
		}
		if flag {
			// 修改内容
			total++
			data := strings.Split(v,",")
			ip := data[0]
			// 1.0.1.0/24,1814991,1814991,,0,0
			// 判断是否是cidr格式
			if cm.CheckPrivateIPValid(ip) == false {
				log.Errorf("get ip is not cidr format: %v",ip)
				continue
			}
			startIP,endIP,_ := cm.GetCidrIpRange(ip)
			if startIP == "" || endIP == "" {
				log.Errorf("get ip cidr format: %v, startIP=%v,endIP=%v!",ip,startIP,endIP )
				continue
			}
			start_local_id := cm.InetAtoN(startIP)
			end_local_id := cm.InetAtoN(endIP)

			// if (end_local_id-start_local_id) != int64(number) {
			// 	log.Errorf("number=%d != end-start (%d-%d) ",number,end_local_id,start_local_id)
			// }
			// out := fmt.Sprintf("\"%s\",\"%s\",\"%d\",\"%d\",\"%s\",\"%s\"\n",
			// 	startIP,endIP,start_local_id,end_local_id,opts.CountryCode,
			// 	opts.CountryName)
			outCsv := []string{"\"%s\"","\"%s\"","\"%d\"","\"%d\"","\"%s\"","\"%s\""}
			outCsv[0] = fmt.Sprintf("\"%s\"",startIP)
			outCsv[1] = fmt.Sprintf("\"%s\"",endIP)
			outCsv[2] = fmt.Sprintf("\"%d\"",start_local_id)
			outCsv[3] = fmt.Sprintf("\"%d\"",end_local_id)
			outCsv[4] = fmt.Sprintf("\"%s\"",opts.CountryCode)
			outCsv[5] = fmt.Sprintf("\"%s\"",opts.CountryName)
			// outCsv[0] = startIP +`"`
			// // 
			// outCsv[1] = endIP
			// outCsv[2] = fmt.Sprintf("%d",start_local_id)
			// outCsv[3] = fmt.Sprintf("%d",end_local_id)
			// outCsv[4] = opts.CountryCode
			// outCsv[5] = opts.CountryName
			err := csvWriter.Write(outCsv)
			csvWriter.Flush()
			// out := ip + ","+opts.GeonameID + "," + opts.GeonameID + ",,0,0\n"
			// 输出
			// _,err = file.WriteString(out)
			if err != nil {
				log.Errorf("write line to %v, line:%v",this.OutFile,outCsv) 
			}
			_,err = fileCidr.WriteString(ip+"\n")
			if err != nil {
				log.Errorf("write line to %v, line:%v",bak,ip) 
			}
		}
		
	}
	csvWriter.Flush()
	log.Infof("---- total: len = %v",total)
}

func main(){
	log.Info("----- begin -----")
	log.Infof("opts: %v",opts)

	app := NewExchange(opts.OutPut,opts.InPut,opts.Filter)
	// app.Exchange()
	app.Exchange2X()
	// switch opts.Command {
	// 	case "ctn":
	// 		// 提取 ctn 和 BGP 两个的数据转换
	// 	case "cun":
	// 		// 提取 ctn 和 BGP 两个的数据转换		
	// 	case "cmn":
	// 		// 提取 ctn 和 BGP 两个的数据转换
	// 	case "ctn&cun":
	// 		// 提取 ctn cun 和 BGP 的数据转换
	// 	case "ctn&cmn":
	// 		// 提取 ctn cmn 和 BGP 的数据转换
	// 	case "cun&cmn":
	// 		// 提取 cmn cun和 BGP 的数据转换			
	// 	case "cun&cmn&ctn":
	// 		// 提取 ctn cun cmn 和 BGP 的数据转换		
	// 	default:
	// 		log.Warnf("do nothing")
	// }

	log.Info("----- over -----")
	
}