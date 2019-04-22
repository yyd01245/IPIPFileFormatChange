#include <iostream>  
#include <fstream>
#include <vector>
#include <string>
#include <sstream>
#include <string.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <cctype>
#include <algorithm>
#include <unistd.h>
#define BLOCK_SIZE 1000
#define SEPARATOR ","
#define _IP_MARK "."

/* ipip 库 共计 2920757 条数据，364M 大小 
直接合并第一次变为 203577条 ip 段 , ip 段展开为 cidr 完毕后变为 296246 条*/

static uint32_t test_merge_num = 0;
static uint32_t test_zhankai_num = 0;
using namespace std;  
typedef struct ChangeInfo {
    in_addr_t begin_int_ipaddr;
    in_addr_t end_int_ipaddr;
    string country;
    string cidr;
    string isp;
    string cN;
}DataInfo_t;

static bool isValidMask(unsigned int mask)
{
    mask = ~mask + 1;
    if ((mask & (mask - 1)) == 0)//判断是否W为2^n次方
        return true;
    else
        return false;
}


const string JType = "1000";

const string GlobalType = "100";

// const int ChinaType = 100;



class IPIPFileFormatChange {
    public:
        IPIPFileFormatChange(const char* file_path,const char* file_param);
        ~IPIPFileFormatChange();
        void HandleInfo(string& line);

    private:
        in_addr_t netmask(int prefix);
        in_addr_t Ip2Int(const string & strIP);
        string Int2Ip(uint32_t num);
        void AddLine(DataInfo_t & line_info);
        void IpRange2Cidr(DataInfo_t src_info);
        void IpRange2Cidr(in_addr_t begin_addr, in_addr_t end_addr);
        void Write2DstFile(fstream &out_file, vector<DataInfo_t> & file_block, uint32_t write_line_len);
        void Flush2DstFile();

        /* 第一次合并的 缓存，即 cN,isp 一致，且 ip 连续的都合并 */
        DataInfo_t cache_info_;
        /* 第二次待合并的缓存区,country 字段为 China, 
            未超过 /16 子网 内的 country字段为其他, 直到超出/16,
            或者再次碰到country 字段为 China 释放 */
        vector<DataInfo_t> cN_buffer_;
        /* 在本场景中，此缓存意义不大，但无多少影响，暂时保留使用 */
        vector<DataInfo_t> file_block_;
        // global file
        fstream out_file_;
        // china file
        fstream out_china_file_;
};

IPIPFileFormatChange::IPIPFileFormatChange(const char*  file_path,const char* china_path) {
    out_file_.open(file_path,ios::out);
    out_china_file_.open(china_path,ios::out);
    // IpRange2Cidr(Ip2Int("59.32.0.0"), Ip2Int("59.43.180.105"));
}

IPIPFileFormatChange::~IPIPFileFormatChange() {
    Flush2DstFile();
}

in_addr_t IPIPFileFormatChange::netmask(int prefix) {
	/* Shifting by 32 is undefined behavior, so 0 prefix is a special case. */
	return prefix == 0 ? 0 : ~(in_addr_t)0 << (32 - prefix);
}

in_addr_t IPIPFileFormatChange::Ip2Int(const string & strIP)
{
    in_addr_t nRet = 0;

    char chBuf[16] = "";
    memcpy(chBuf, strIP.c_str(), 15);

    char* szBufTemp = NULL;
    char* szBuf = strtok_r(chBuf,_IP_MARK,&szBufTemp);

    int i = 0;//计数
    while(NULL != szBuf)//取一个
    {
        nRet += atoi(szBuf)<<((3-i)*8);
        szBuf = strtok_r(NULL,_IP_MARK,&szBufTemp);
        i++;
    }

    return nRet;
}
string IPIPFileFormatChange::Int2Ip(uint32_t num)
{  

    string strRet = "";  
    for (int i=0;i<4;i++)  
    {  
        uint32_t tmp=(num>>((3-i)*8))&0xFF;  

        char chBuf[8] = "";
        sprintf(chBuf, "%d", tmp);
        strRet += chBuf;

        if (i < 3)
        {
            strRet += _IP_MARK;
        }
    }  

    return strRet;  
} 
void IPIPFileFormatChange::AddLine(DataInfo_t & line_info){
    /* 到达 BLOCK_SIZE 长度则写入文件 */
    if(file_block_.size() == BLOCK_SIZE){
        Write2DstFile(out_file_, file_block_, BLOCK_SIZE); 
    }

    file_block_.push_back(line_info);
}

void IPIPFileFormatChange::IpRange2Cidr(DataInfo_t src_info) {
	int prefix;
	in_addr_t brdcst, mask;
    DataInfo_t new_info;
    stringstream cidr_str;
    in_addr_t begin_addr = src_info.begin_int_ipaddr;
    in_addr_t end_addr = src_info.end_int_ipaddr;

	do {
        cidr_str.clear(); 
		for(prefix = 0;; prefix++) {
			mask = netmask(prefix);
			brdcst = begin_addr | ~mask;
			if((begin_addr & mask) == begin_addr && brdcst <= end_addr){
				break;
            }
        }

        new_info.begin_int_ipaddr = begin_addr & mask;

        new_info.end_int_ipaddr = brdcst;

        /* 填充 cidr 字段 */
	    cidr_str << Int2Ip(begin_addr)<<"/"<< prefix;
        new_info.cidr  = cidr_str.str();

        new_info.isp = src_info.isp;

        new_info.cN = src_info.cN;

        /* 增加条目 */
        AddLine(new_info);
        cidr_str.str(""); 
        test_zhankai_num++;
		if(brdcst == ~(in_addr_t)0)
			break; /* Prevent overflow on the next line. */
		begin_addr = brdcst + 1;
	} while(begin_addr <= end_addr);
    test_zhankai_num--;
    return;
}

void IPIPFileFormatChange::IpRange2Cidr(in_addr_t begin_addr, in_addr_t end_addr) {
	int prefix;
	in_addr_t brdcst, mask;
	do {
		for(prefix = 0;; prefix++) {
			mask = netmask(prefix);
			brdcst = begin_addr | ~mask;
			if((begin_addr & mask) == begin_addr && brdcst <= end_addr){
                // printf("brdcast = %s\n",Int2Ip(brdcst).c_str());
				break;
            }
        }
        
		printf("/%d\n", prefix);
        printf("begin_addr = %s\n",Int2Ip(begin_addr).c_str());
        printf("brdcast = %s\n",Int2Ip(brdcst).c_str());
        
		if(brdcst == ~(in_addr_t)0)
			break; /* Prevent overflow on the next line. */
		begin_addr = brdcst + 1;
	} while(begin_addr <= end_addr);
}

void IPIPFileFormatChange::Write2DstFile(fstream &out_file, vector<DataInfo_t>& file_block, uint32_t write_line_len){
    std::vector<DataInfo_t>::iterator it;
    for( it = file_block.begin(); it != file_block.end();) {
        /* 将纯大写的中国运营商改为首先大写的样式 */

        transform((*it).isp.begin(), (*it).isp.end(), (*it).isp.begin(), towupper);
        // #ifdef OLD_CODE
          // ipip 转换使用
          if((*it).isp == "CHINATELECOM") {
              (*it).isp = "ctn";
          } else if((*it).isp == "CHINAUNICOM" || (*it).isp == "WASU"){
              (*it).isp = "cun";
          } else if ((*it).isp == "CHINAMOBILE" || (*it).isp == "CHINARAILCOM"){
              (*it).isp = "cmn";         
          }
        // #endif

        // if((*it).isp == "CHINATELECOM") {
        //     transform((*it).isp.begin(), (*it).isp.end(), (*it).isp.begin(), towlower);
        //     (*it).isp.replace(0,1,"C");
        //     (*it).isp.replace(5,1,"T");
        // } else if((*it).isp == "CHINAUNICOM"){
        //     transform((*it).isp.begin(), (*it).isp.end(), (*it).isp.begin(), towlower);
        //     (*it).isp.replace(0,1,"C");
        //     (*it).isp.replace(5,1,"U");
        // } else if ((*it).isp == "CHINAMOBILE"){
        //     transform((*it).isp.begin(), (*it).isp.end(), (*it).isp.begin(), towlower);
        //     (*it).isp.replace(0,1,"C");
        //     (*it).isp.replace(5,1,"M");            
        // }
        // #ifdef OLD_CODE
          // ipip 转换使用
          /* 运营商修改，将符合标准的运营商修改为 BGP */
          if((*it).isp.rfind("ALIYUN") != string::npos ||
              (*it).isp.rfind("TENCENT") != string::npos ||
              (*it).isp.rfind(".cn") != string::npos ||
              (*it).isp.rfind(".org") != string::npos ||
              (*it).isp.rfind(".net") != string::npos ||
              (*it).isp.rfind(".com") != string::npos
              ) {
              (*it).isp = "BGP";
          }
          
          /* 国家为 中国和香港，且运营商为*的 都将运行商改为BGP */
          if(((*it).cN == "CN" ||
          (*it).cN == "HK") &&
          (*it).isp == "*") {
              (*it).isp = "BGP";
          }
        // #endif
        // out_file
        // <<Int2Ip((*it).begin_int_ipaddr)<<"\t"
        // <<Int2Ip((*it).end_int_ipaddr)<<"\t"
        // <<(*it).isp<<"\t"
        // <<(*it).cN<<"\r\n";
        // 过滤掉多余的
        if ((*it).cidr == "0.0.0.0/8" || 
            (*it).cidr == "224.0.0.0/3" ||
            (*it).cidr == "169.254.0.0/16") {
            cout<< "ignor local cidr:"<<(*it).cidr<<endl;
        }

        if ((*it).cN == "CN") {
            string out_data = "1";
            if ((*it).isp == "cun") {
                out_data = "2";
            }else if ((*it).isp == "cmn") {
                out_data = "3";
            }
            out_china_file_
                <<(*it).cidr<<SEPARATOR
                // <<(*it).isp<<SEPARATOR
                <<out_data<<"\r\n";
        }else {
          out_file_
            <<(*it).cidr<<SEPARATOR
            // <<(*it).isp<<SEPARATOR
            <<GlobalType<<"\r\n";
            // <<"10"<<"\r\n";
        }
        #ifdef OLD_CODE
          // ipip 转换使用
          out_file
          // <<Int2Ip((*it).begin_int_ipaddr)<<SEPARATOR
          // <<Int2Ip((*it).end_int_ipaddr)<<SEPARATOR
          // <<(*it).begin_int_ipaddr<<SEPARATOR
          // <<(*it).end_int_ipaddr<<SEPARATOR
          
          <<(*it).cidr<<SEPARATOR
          <<(*it).isp<<SEPARATOR
          <<(*it).cN<<SEPARATOR<<(*it).country<<"\r\n";
        #endif

        // out_file
        // <<"\""<<Int2Ip((*it).begin_int_ipaddr)<<"\""<<SEPARATOR
        // <<"\""<<Int2Ip((*it).end_int_ipaddr)<<"\""<<SEPARATOR
        // <<"\""<<(*it).begin_int_ipaddr<<"\""<<SEPARATOR
        // <<"\""<<(*it).end_int_ipaddr<<"\""<<SEPARATOR
        // // <<"\""<<(*it).cidr<<"\""<<SEPARATOR
        // // <<"\""<<(*it).isp<<"\""<<SEPARATOR
        // <<"\""<<(*it).cN<<"\""<<"\r\n";

        file_block.erase(it);
    }

    // /* 计算剩余队列长度 */
    // cout<<"写入长度:" <<write_line_len
    //     <<"剩余长度:" <<file_block.size()<<endl;
    return;
}

void IPIPFileFormatChange::HandleInfo(string& line)  
{  
    int nSPos = 0;  
    int nEPos = 0;
    /* 字段序号以0开始 */
    int field_index = 0; 
    DataInfo_t curr_cache_info;
    string str;
    string pri;
    while ((nEPos = line.find('\t', nSPos)) != string::npos) {
        str = line.substr(nSPos, nEPos - nSPos);
        nSPos = nEPos + 1;  // 为下一次检索做准备
        /* 提取自己关注的字段 */
        switch (field_index){
            case 0:
                curr_cache_info.begin_int_ipaddr = Ip2Int(str);
                break;
            case 1:
                curr_cache_info.end_int_ipaddr = Ip2Int(str);
                break;
            case 2:
                curr_cache_info.country = str;
                break;
            case 3:
                pri = str;
                break;
            case 6:
                curr_cache_info.isp = str;
                break;
            case 13:
                curr_cache_info.cN = str;
                break;
            default:
                break;
        }
        if(curr_cache_info.country == "Hong Kong" || pri == "Hong Kong"){
            curr_cache_info.cN = "HK";
        }

        field_index++;
    }
    if(curr_cache_info.country == "114DNS.COM" || curr_cache_info.country == "ALIDNS.COM"
        || curr_cache_info.country == "TENCENT.COM" || curr_cache_info.country == "DNSPOD.COM"
        || curr_cache_info.country == "CHINANETCENTER.COM"){
        curr_cache_info.cN = "CN";
        curr_cache_info.country = "China";
    }

    /* 读取最后一个字段,暂时不需要最后一个字段 */
    // str = line.substr(nSPos, line.size() - nSPos);
    // cout << str <<endl;  

    /* 如果是第一个数据，仅刷新缓存 */
    if(cache_info_.cN == ""){
        cache_info_ = curr_cache_info;
        return;
    }
     /* 如果 运营商和国家一致，且当前起始 ip 与 上一条缓存 终止 ip 连续（差值为1），则合并 */
    if(curr_cache_info.isp == cache_info_.isp &&
    curr_cache_info.cN == cache_info_.cN &&
    (curr_cache_info.begin_int_ipaddr - cache_info_.end_int_ipaddr == 1)) {
        /* 将当前的信息合并进 file_block */
        cache_info_.end_int_ipaddr = curr_cache_info.end_int_ipaddr;
        test_merge_num++;
    }else{
        /* 缓存写入 file_block_ */
        IpRange2Cidr(cache_info_);
        // AddLine(cache_info_);

        /* 刷新缓存数据*/
        cache_info_ = curr_cache_info;
    }

    return;
} 

void IPIPFileFormatChange::Flush2DstFile(){
    /* 缓存写入 file_block_ */
    IpRange2Cidr(cache_info_);
    // AddLine(cache_info_);
    /* 写入剩余的数据 */
    Write2DstFile(out_file_, file_block_, file_block_.size()); 
    out_file_.close();
}

int main(int argc, char *argv[])  
{  

    cout<<"argv:---"<<argv[argc-1]<<endl;
    fstream input_file;
    unique_ptr<IPIPFileFormatChange> ipRange2Cidr;
    int opt = getopt( argc, argv, "i:c:g:");
    // char* global_file = NULL;
    // char* china_file = NULL;
    string global_file = "";
    string china_file = "";

    while(  opt != -1 ) {
        cout<< "opt="<< opt <<endl;
        switch( opt ) {
            case 'i':
                input_file.open(optarg, ios::in);
                break;
            // case 'o':
            //     ipRange2Cidr.reset(new IPIPFileFormatChange(optarg,"./china_ip.txt"));
            //     break;
            case 'g':
                cout<<"optarg:"<<optarg<<endl;
                global_file = optarg;
                cout<< "get global file ="<< global_file<<endl;
                break;
            case 'c':
                cout<<"optarg:"<<optarg<<endl;
                china_file = optarg;
                cout<< "get china file ="<< china_file<<endl;
                break;
            // case '?':
                // display_usage();
                // break;   
            default:
                /* You won't actually get here. */
                fputs("usage: cidr [-i] [input file] [-c] [output file] [-g] [output file]\n", stderr);
			    exit(1);
                break;
        }
                
        opt = getopt( argc, argv, "i:c:g:");
        // argc -= optind;
	    // argv += optind;
        cout<< "argc="<< argc<<endl;
        cout<< "argv="<< argv[argc-1]<<endl;
    }
    cout<< "opt="<< opt <<endl;
    
    cout<< "global_file ="<<global_file<<endl;
    cout<< "china_file ="<<china_file<<endl;

    // if (global_file != NULL && china_file != NULL) {
    if (global_file != "" && china_file != "") {
      ipRange2Cidr.reset(new IPIPFileFormatChange(global_file.c_str(),china_file.c_str()));
    }

    string line;  
    if(ipRange2Cidr == nullptr){
        cout << "ipRange2Cidr is null!" << endl;
        fputs("usage: cidr [-i] [input file] [-c] [output file] [-g] [output file]\n", stderr);
		exit(1);
    }
    if(!input_file.is_open())  
    {  
        cout << "open file fail!" << endl; 
        fputs("usage: cidr [-i] [input file] [-c] [output file] [-g] [output file]\n", stderr);
		exit(1); 
    }     
    int i;
    uint32_t write_len = 0;
    while(getline(input_file,line)){  
        ipRange2Cidr->HandleInfo(line); 
    }

    input_file.close();  
    cout<< "merge number:"<<test_merge_num<<
        ", expand number:"<<test_zhankai_num<<endl;
    return 0;  
}
