package main

import (
	"fmt"
	"strings"
	"strconv"
	// "./csv"
	"io"
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

func (this *IPIPExchange) ExchangeIPIP() {
	// 读取文件
	inFile, err := os.OpenFile(this.InFile,os.O_RDONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer inFile.Close()
	// 创建输出文件
	file, err := os.OpenFile(this.OutFile,os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)

	fileCN, err := os.OpenFile("china_geoip.csv",os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer fileCN.Close()
	csvCNWriter := csv.NewWriter(fileCN)

	csvReader := csv.NewReader(inFile)

	// stringBody := string(body)
	// // 获取一行行数据
	// outLine := strings.Split(stringBody,"\n")
	total := 0
	for  {
		record,err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("read csv line error: %v",err)
			continue
		}
		log.Debugf("--- get data: %v,len=%d",record,len(record))
		log.Debugf("record 0=%s",record[0])
		data := strings.Split(record[0],"\t")
		// log.Infof("data-=%v,len=%d",data,len(data))
		// 行数据
		if len(data) != 15 {
			log.Errorf("data:%v, len=%d, is wrong!",data,len(data))
			continue
		}
		startIPstr := data[0]
		endIPstr := data[1]
		dotIPIntList := strings.Split(startIPstr,".")
		dotEndIPIntList := strings.Split(endIPstr,".")
		log.Debugf("--- dotstart:%v,endStart:%v",dotIPIntList,dotEndIPIntList)
		intStartIP := []string{}
		intEndIP := []string{}
		for _,v := range dotIPIntList {
			a,err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intStartIP = append(intStartIP,strconv.Itoa(a))
			log.Debugf("get start int : %d",a)
		}
		for _,v := range dotEndIPIntList {
			a,_ := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intEndIP = append(intEndIP,strconv.Itoa(a))
			log.Debugf("get end int : %d",a)
		}		
		log.Debugf("intStartIP=%v,len=%d",intStartIP,len(intStartIP))
		log.Debugf("intEndIP=%v,len=%d",intEndIP,len(intEndIP))

		if len(intStartIP) != 4 || len(intEndIP) != 4 {
			log.Errorf("error ip:start=%v,end=%v",dotIPIntList,dotEndIPIntList)
		}
		// break;
		startIP := strings.Join(intStartIP,".")
		endIP := strings.Join(intEndIP,".")
		total++
		// if total > 20 {
		// 	break;
		// }
		// 得到国家
		country := data[2]
		isp := data[6]
		countryCode := data[13]
		start_local_id := cm.InetAtoN(startIP)
		end_local_id := cm.InetAtoN(endIP)

		if country == "114DNS.COM" || 
			country == "ALIDNS.COM" || country == "TENCENT.COM" || 
			country == "DNSPOD.COM" || country == "CHINANETCENTER.COM" ||
			country == "SDNS.CN" {
				country = "China"
				countryCode = "CN"
			}
		if country == "China" && countryCode == "CN"{
				// country = "China"
				// countryCode = "CN"
				chinaCountry := ""
				chinaCountryCode := countryCode
				// 区分运营商
				localISP := strings.ToUpper(isp)
				if localISP == "CHINATELECOM" {
					// isp = ""
					chinaCountry = "ChinaCTN"
					chinaCountryCode = "XT"
				}else if localISP == "CHINAUNICOM" || localISP == "WASU" {
					// isp = ""
					chinaCountry = "ChinaCUN"
					chinaCountryCode = "XU"
				}else if localISP == "CHINAMOBILE" || localISP == "CHINARAILCOM" {
					// isp = ""
					chinaCountry = "ChinaCMN"
					chinaCountryCode = "XM"
				}else {
					// if strings.Index(isp,"ALIYUN") >= 0 ||
					// strings.Index(isp,"TENCENT") >= 0 ||
					// strings.Index(isp,".cn") >= 0 ||
					// strings.Index(isp,".org") >= 0 ||
					// strings.Index(isp,".net") >= 0 ||
					// strings.Index(isp,".com") >= 0 
						chinaCountry = "ChinaBGP"
						chinaCountryCode = "XB"
				}
				outCNCsv := []string{"\"%s\"","\"%s\"","\"%d\"","\"%d\"","\"%s\"","\"%s\""}
				outCNCsv[0] = fmt.Sprintf("\"%s\"",startIP)
				outCNCsv[1] = fmt.Sprintf("\"%s\"",endIP)
				outCNCsv[2] = fmt.Sprintf("\"%d\"",start_local_id)
				outCNCsv[3] = fmt.Sprintf("\"%d\"",end_local_id)
				outCNCsv[4] = fmt.Sprintf("\"%s\"",chinaCountryCode)
				outCNCsv[5] = fmt.Sprintf("\"%s\"",chinaCountry)
				err = csvCNWriter.Write(outCNCsv)
				if err != nil {
					log.Errorf("write line to %v, line:%v","china_geoip.csv",outCNCsv) 
				}
				log.Infof("get china country:%v",outCNCsv)

				csvCNWriter.Flush()
		}
		if country == "Asia Pacific Regions" && countryCode == "*" {
			countryCode = "HK"
		}
		if countryCode == "*" {  
			// country == "*" && 
			// 未知国家归入 XX
			countryCode = "XX"

		}
		outCsv := []string{"\"%s\"","\"%s\"","\"%d\"","\"%d\"","\"%s\"","\"%s\""}
		outCsv[0] = fmt.Sprintf("\"%s\"",startIP)
		outCsv[1] = fmt.Sprintf("\"%s\"",endIP)
		outCsv[2] = fmt.Sprintf("\"%d\"",start_local_id)
		outCsv[3] = fmt.Sprintf("\"%d\"",end_local_id)
		outCsv[4] = fmt.Sprintf("\"%s\"",countryCode)
		outCsv[5] = fmt.Sprintf("\"%s\"",country)
		err = csvWriter.Write(outCsv)
		if err != nil {
			log.Errorf("write line to %v, line:%v",this.OutFile,outCsv) 
		}
		csvWriter.Flush()
		log.Infof("get global country:%v",outCsv)



	}
	// log.Infof("---- over: total len= %d",total)
	csvWriter.Flush()
	csvCNWriter.Flush()
	log.Errorf("---- total: len = %v",total)
}

func (this *IPIPExchange) CheckIPIPGlobal() {
	// 读取原始文件
	inFile, err := os.OpenFile(this.InFile,os.O_RDONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer inFile.Close()
	// 读取对比的文件
	file, err := os.OpenFile(this.OutFile,os.O_RDONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer file.Close()
	csvComperReader := csv.NewReader(file)
	csvReader := csv.NewReader(inFile)

	allComper,err := csvComperReader.ReadAll()
	if err != nil {
		log.Errorf("read compare csv error: %v",err)
		return
	}
	// stringBody := string(body)
	// // 获取一行行数据
	// outLine := strings.Split(stringBody,"\n")
	total := 0
	lastUnMatchList := []string{}
	for  {
		record,err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("read csv line error: %v",err)
			continue
		}
		log.Debugf("--- get data: %v,len=%d",record,len(record))
		log.Debugf("record 0=%s",record[0])
		data := strings.Split(record[0],"\t")
		// log.Infof("data-=%v,len=%d",data,len(data))
		// 行数据
		// if len(data) != 15 {
		// 	log.Errorf("data:%v, len=%d, is wrong!",data,len(data))
		// 	continue
		// }
		startIPstr := data[0]
		endIPstr := data[1]
		dotIPIntList := strings.Split(startIPstr,".")
		dotEndIPIntList := strings.Split(endIPstr,".")
		log.Debugf("--- dotstart:%v,endStart:%v",dotIPIntList,dotEndIPIntList)
		intStartIP := []string{}
		intEndIP := []string{}
		for _,v := range dotIPIntList {
			a,err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intStartIP = append(intStartIP,strconv.Itoa(a))
			log.Debugf("get start int : %d",a)
		}
		for _,v := range dotEndIPIntList {
			a,_ := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intEndIP = append(intEndIP,strconv.Itoa(a))
			log.Debugf("get end int : %d",a)
		}		
		log.Debugf("intStartIP=%v,len=%d",intStartIP,len(intStartIP))
		log.Debugf("intEndIP=%v,len=%d",intEndIP,len(intEndIP))

		if len(intStartIP) != 4 || len(intEndIP) != 4 {
			log.Errorf("error ip:start=%v,end=%v",dotIPIntList,dotEndIPIntList)
		}
		// break;
		startIP := strings.Join(intStartIP,".")
		endIP := strings.Join(intEndIP,".")
		total++
		// if total > 20 {
		// 	break;
		// }
		// 得到国家
		country := data[2]
		// isp := data[6]
		countryCode := data[13]
		start_local_id := cm.InetAtoN(startIP)
		end_local_id := cm.InetAtoN(endIP)
		if country == "114DNS.COM" || 
			country == "ALIDNS.COM" || country == "TENCENT.COM" || 
			country == "DNSPOD.COM" || country == "CHINANETCENTER.COM" ||
			country == "SDNS.CN" {
				country = "China"
				countryCode = "CN"
		}
		if country == "Asia Pacific Regions" && countryCode == "*" {
			countryCode = "HK"
		}
		if countryCode == "*" {  
			// country == "*" && 
			// 未知国家归入 XX
			countryCode = "XX"

		}
		findFlag := false
		for index, value := range allComper {
			log.Infof("---index=%d,value=%v",index,value)
			if len(value) != 6 {
				log.Errorf("compare file line len!=6, value=%v,len=%d!!!",value,len(value))
			}
			// data := strings.Split(value,"\t")
			startID,_ :=strconv.Atoi(value[2])
			endID,_ :=strconv.Atoi(value[3])
			if value[0]==startIP && value[1]==endIP && int64(startID)==start_local_id &&
				int64(endID) == end_local_id && value[4] == countryCode && value[5] == country {
					log.Infof("find line in compare file:%v",value)
					allComper = append(allComper[:index],allComper[index+1:]...)
					findFlag = true
					break;
			}
		}
		if findFlag == false {
			lastUnMatchList = append(lastUnMatchList,record[0])
		}
		// break
	}

	log.Errorf("commpare file last : %v,len=%d!",allComper,len(allComper))
	log.Errorf("---org file  diff: %v,len=%d!",lastUnMatchList,len(lastUnMatchList))

	log.Errorf("---- total: len = %v",total)
}
func (this *IPIPExchange) CheckIPIPChina() {
	// 读取原始文件
	inFile, err := os.OpenFile(this.InFile,os.O_RDONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer inFile.Close()
	// 读取对比的文件
	file, err := os.OpenFile(this.OutFile,os.O_RDONLY,0644)
	if err != nil {
		log.Warnf("open file failed !")
		return
	}
	defer file.Close()
	csvComperReader := csv.NewReader(file)
	csvReader := csv.NewReader(inFile)

	allComper,err := csvComperReader.ReadAll()
	if err != nil {
		log.Errorf("read compare csv error: %v",err)
		return
	}
	// stringBody := string(body)
	// // 获取一行行数据
	// outLine := strings.Split(stringBody,"\n")
	total := 0
	lastUnMatchList := []string{}
	for  {
		record,err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("read csv line error: %v",err)
			continue
		}
		log.Debugf("--- get data: %v,len=%d",record,len(record))
		log.Debugf("record 0=%s",record[0])
		data := strings.Split(record[0],"\t")
		// log.Infof("data-=%v,len=%d",data,len(data))
		// 行数据
		// if len(data) != 15 {
		// 	log.Errorf("data:%v, len=%d, is wrong!",data,len(data))
		// 	continue
		// }
		startIPstr := data[0]
		endIPstr := data[1]
		dotIPIntList := strings.Split(startIPstr,".")
		dotEndIPIntList := strings.Split(endIPstr,".")
		log.Debugf("--- dotstart:%v,endStart:%v",dotIPIntList,dotEndIPIntList)
		intStartIP := []string{}
		intEndIP := []string{}
		for _,v := range dotIPIntList {
			a,err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intStartIP = append(intStartIP,strconv.Itoa(a))
			log.Debugf("get start int : %d",a)
		}
		for _,v := range dotEndIPIntList {
			a,_ := strconv.Atoi(v)
			if err != nil {
				continue
			}
			intEndIP = append(intEndIP,strconv.Itoa(a))
			log.Debugf("get end int : %d",a)
		}		
		log.Debugf("intStartIP=%v,len=%d",intStartIP,len(intStartIP))
		log.Debugf("intEndIP=%v,len=%d",intEndIP,len(intEndIP))

		if len(intStartIP) != 4 || len(intEndIP) != 4 {
			log.Errorf("error ip:start=%v,end=%v",dotIPIntList,dotEndIPIntList)
		}
		// break;
		startIP := strings.Join(intStartIP,".")
		endIP := strings.Join(intEndIP,".")
		total++
		// if total > 20 {
		// 	break;
		// }
		// 得到国家
		country := data[2]
		isp := data[6]
		countryCode := data[13]
		start_local_id := cm.InetAtoN(startIP)
		end_local_id := cm.InetAtoN(endIP)
		if country == "114DNS.COM" || 
			country == "ALIDNS.COM" || country == "TENCENT.COM" || 
			country == "DNSPOD.COM" || country == "CHINANETCENTER.COM" ||
			country == "SDNS.CN" {
				country = "China"
				countryCode = "CN"
		}
		if countryCode != "CN" {
			continue
		}
		localISP := strings.ToUpper(isp)
		if localISP == "CHINATELECOM" {
			// isp = ""
			country = "ChinaCTN"
			countryCode = "XT"
		}else if localISP == "CHINAUNICOM" || localISP == "WASU" {
			// isp = ""
			country = "ChinaCUN"
			countryCode = "XU"
		}else if localISP == "CHINAMOBILE" || localISP == "CHINARAILCOM" {
			// isp = ""
			country = "ChinaCMN"
			countryCode = "XM"
		}else {
			// if strings.Index(isp,"ALIYUN") >= 0 ||
			// strings.Index(isp,"TENCENT") >= 0 ||
			// strings.Index(isp,".cn") >= 0 ||
			// strings.Index(isp,".org") >= 0 ||
			// strings.Index(isp,".net") >= 0 ||
			// strings.Index(isp,".com") >= 0 
				country = "ChinaBGP"
				countryCode = "XB"
		}
		findFlag := false
		for index, value := range allComper {
			log.Infof("---index=%d,value=%v",index,value)
			if len(value) != 6 {
				log.Errorf("compare file line len!=6, value=%v,len=%d!!!",value,len(value))
			}
			// data := strings.Split(value,"\t")
			startID,_ :=strconv.Atoi(value[2])
			endID,_ :=strconv.Atoi(value[3])
			if value[0]==startIP && value[1]==endIP && int64(startID)==start_local_id &&
				int64(endID) == end_local_id && value[4] == countryCode && value[5] == country {
					log.Infof("find line in compare file:%v",value)
					allComper = append(allComper[:index],allComper[index+1:]...)
					findFlag = true
					break;
			}
		}
		if findFlag == false {
			lastUnMatchList = append(lastUnMatchList,record[0])
		}
		// break
	}

	log.Errorf("commpare file last : %v,len=%d!",allComper,len(allComper))
	log.Errorf("---org file  diff: %v,len=%d!",lastUnMatchList,len(lastUnMatchList))
	log.Errorf("---- total: len = %v",total)
}

func main(){
	log.Info("----- begin -----")
	log.Infof("opts: %v",opts)

	app := NewExchange(opts.OutPut,opts.InPut,opts.Filter)
	// app.Exchange()
	if opts.Type == "ipmask" {
		app.Exchange2X()
	}else if opts.Type == "ipip" {
		app.ExchangeIPIP()
	}else if opts.Type == "checkglobal" {
		app.CheckIPIPGlobal()
	}else if opts.Type == "checkchina" {
		app.CheckIPIPChina()
	}
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