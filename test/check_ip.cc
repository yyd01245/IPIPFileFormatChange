#include <iostream>
#include <string>
#include <string.h>
#include <arpa/inet.h>
#include <fstream>
#include <sstream>
#include <cctype>
#include <algorithm>
#include <vector>
using namespace std; 

#define _IP_MARK "."
#define SRC_FILE "./ipip.csv"
#define DST_FILE "./iproute.csv"
typedef struct SrcIPInfo {
    in_addr_t begin_int_ipaddr;
    in_addr_t end_int_ipaddr;
    string isp;
    string country;
}SrcIPInfo_t;

typedef struct DstIPInfo {
    in_addr_t int_ipaddr;
    int prefix;
    string isp;
    string country;
}DstIPInfo_t;

in_addr_t netmask(int prefix) {
	/* Shifting by 32 is undefined behavior, so 0 prefix is a special case. */
	return prefix == 0 ? 0 : ~(in_addr_t)0 << (32 - prefix);
}
string Int2Ip(uint32_t num)
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
in_addr_t Ip2Int(const string & strIP)
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

inline void GetSrcIPInfo(string& line, SrcIPInfo_t&  ip_msg);
void GetSrcIPInfo(string& line, SrcIPInfo_t&  ip_msg)  
{  
    int nSPos = 0; 
    int nEPos = 0;
    /* 字段序号以0开始 */
    int field_index = 0; 
    string str;
    while ((nEPos = line.find('\t', nSPos)) != string::npos) {
        str = line.substr(nSPos, nEPos - nSPos);
        nSPos = nEPos + 1;  // 为下一次检索做准备
        /* 提取自己关注的字段 */
        switch (field_index){
            case 0:
                ip_msg.begin_int_ipaddr = Ip2Int(str);
                break;
            case 1:
                ip_msg.end_int_ipaddr = Ip2Int(str);
                break;
            case 6:
                ip_msg.isp = str;
                break;
            case 13:
                ip_msg.country = str;
                break;
            default:
                break;
        }

        field_index++;
    }

    return;
} 

void GetDstIPInfo(string& line, DstIPInfo_t&  ip_msg)  
{  
    int nSPos = 0; 
    int nEPos = 0;
    /* 字段序号以0开始 */
    int field_index = 0; 
    string str;
    while ((nEPos = line.find(',', nSPos)) != string::npos) {
        str = line.substr(nSPos, nEPos - nSPos);
        nSPos = nEPos + 1;  // 为下一次检索做准备
        stringstream ss;
        /* 提取自己关注的字段 */
        switch (field_index){
            case 0:
                ip_msg.int_ipaddr = Ip2Int(str.substr(0, str.find('/')));
                ss << str.substr(str.find('/') + 1 , str.length() - str.find('/'));
                ss >> ip_msg.prefix;
                break;
            case 1:
                ip_msg.isp = str;
                break;
            default:
                break;
        }

        field_index++;
    }
    /* 最后一个字段, -1 是为了去掉末尾的 \r\n */
    str = line.substr(nSPos, line.size() - nSPos -1);
    ip_msg.country = str;

    //  cout <<"ip@@@@@@@@@@@: "<<ip_msg.int_ipaddr<<endl
    // <<"mask@@@@@@@@@: "<<ip_msg.prefix<<endl;
    return;
} 

inline void ReplaceIsp(SrcIPInfo_t& src_info);
void ReplaceIsp(SrcIPInfo_t& src_info) {
    /* 将纯大写的中国运营商改为首先大写的样式 */
        if(src_info.isp == "CHINATELECOM") {
            transform(src_info.isp.begin(), src_info.isp.end(), src_info.isp.begin(), towlower);
            src_info.isp.replace(0,1,"C");
            src_info.isp.replace(5,1,"T");
        } else if(src_info.isp == "CHINAUNICOM"){
            transform(src_info.isp.begin(), src_info.isp.end(), src_info.isp.begin(), towlower);
            src_info.isp.replace(0,1,"C");
            src_info.isp.replace(5,1,"U");
        } else if (src_info.isp == "CHINAMOBILE"){
            transform(src_info.isp.begin(), src_info.isp.end(), src_info.isp.begin(), towlower);
            src_info.isp.replace(0,1,"C");
            src_info.isp.replace(5,1,"M");            
        }

        /* 运营商修改，将符合标准的运营商修改为 BGP */
        // cout<<"运营商："<<src_info.isp<<endl<<"###"<<src_info.isp.rfind(".cn")<<endl;
        if(src_info.isp.rfind("ALIYUN") != string::npos ||
            src_info.isp.rfind(".cn") != string::npos ||
            src_info.isp.rfind(".org") != string::npos ||
            src_info.isp.rfind(".net") != string::npos ||
            src_info.isp.rfind(".com") != string::npos
            ) {
            src_info.isp = "BGP";
        }
        
        /* 国家为 中国和香港，且运营商为*的 都将运行商改为BGP */
        if((src_info.country == "CN" ||
        src_info.country == "HK") &&
        src_info.isp == "*") {
            src_info.isp = "BGP";
        }

        return;
}
inline bool CheckIPInfo(SrcIPInfo_t src_info, DstIPInfo_t dst_info,bool check_begin);
bool CheckIPInfo(SrcIPInfo_t src_info, DstIPInfo_t dst_info,bool check_begin) {
    bool founded = false;
    unsigned int check_ipaddr = check_begin ? src_info.begin_int_ipaddr : src_info.end_int_ipaddr;
    // cout <<"begin check *******"<<endl;
    // cout << Int2Ip(check_ipaddr)<<endl;
    // cout << Int2Ip(netmask(dst_info.prefix))<<endl;
    // cout << Int2Ip((src_info.begin_int_ipaddr & netmask(dst_info.prefix)))<<endl;
    // cout << Int2Ip(dst_info.int_ipaddr)<<"/"<<dst_info.prefix<<endl;
    // cout <<"end check #######"<<endl;
    if((check_ipaddr & netmask(dst_info.prefix)) == dst_info.int_ipaddr){
        founded = true;
        if(!(src_info.isp == dst_info.isp && src_info.country == dst_info.country)){
            cout<<"*******"<<endl
            <<"is isp equal？"<<(src_info.isp == dst_info.isp)<<endl
            <<"is country equal？"<<(src_info.country == dst_info.country)<<endl
            << "error: find but isp or country not match, begin ip: "
            <<Int2Ip(src_info.begin_int_ipaddr)<<" end ip:"<<Int2Ip(src_info.end_int_ipaddr)<<endl
            << Int2Ip(dst_info.int_ipaddr)<<"/"<<dst_info.prefix<<endl
            <<"src isp: <"<<src_info.isp<<"> dst isp: <"<<dst_info.isp<<">"<<endl
            <<"src country: <"<<src_info.country<<"> dst country: <"<<dst_info.country<<">"<<endl
            <<"########"<<endl;
        }
    }

    return founded;
}
int main(int argc, char** argv) {
    fstream src_file;
    fstream dst_file;
    src_file.open(SRC_FILE, ios::in);
    dst_file.open(DST_FILE, ios::in);
    string src_line, dst_line;
    SrcIPInfo_t src_info;
    DstIPInfo_t dst_info;
    bool find_begin = false;
    bool find_end = false;
    unsigned int find_begin_num = 0;
    unsigned int find_end_num = 0;
    time_t start = 0, end = 0;

    // ifstream  ifs("Title.pic", std::ios::binary);
    vector<DstIPInfo_t> buffer;
    // buffer.resize(ifs.seekg(0, std::ios::end).tellg());
    // ifs.seekg(0, std::ios::beg).read( &buffer[0], static_cast<std::streamsize>(buffer.size()) );
    while(getline(dst_file, dst_line)) {
        GetDstIPInfo(dst_line, dst_info);
        buffer.push_back(dst_info);
    }
    size_t count = buffer.size();
    time(&start);
    while(getline(src_file, src_line)){
        find_begin = false;
        find_end= false;
        find_begin_num = 0;
        find_end_num = 0;
        GetSrcIPInfo(src_line, src_info);  
        ReplaceIsp(src_info);

        for (size_t i = 0; i < count; ++i){

            if((src_info.begin_int_ipaddr & netmask(buffer[i].prefix)) == buffer[i].int_ipaddr) {
                if(find_begin == true){
                    find_begin_num++;
                    cout << find_begin_num;
                    cout<< "find more then one , begin ip: "<<Int2Ip(src_info.begin_int_ipaddr)<<endl
                    <<" dst cidr :"<< Int2Ip(buffer[i].int_ipaddr)<<"/"<<buffer[i].prefix<<endl;
                }
                find_begin = true;
                if(!(src_info.isp == buffer[i].isp && src_info.country == buffer[i].country)){
                    cout<<"*******"<<endl
                    <<"is isp equal？"<<(src_info.isp == buffer[i].isp)<<endl
                    <<"is country equal？"<<(src_info.country == buffer[i].country)<<endl
                    << "error: find but isp or country not match, begin ip: "
                    <<Int2Ip(src_info.begin_int_ipaddr)<<" end ip:"<<Int2Ip(src_info.end_int_ipaddr)<<endl
                    << Int2Ip(buffer[i].int_ipaddr)<<"/"<<buffer[i].prefix<<endl
                    <<"src isp: <"<<src_info.isp<<"> dst isp: <"<<buffer[i].isp<<">"<<endl
                    <<"src country: <"<<src_info.country<<"> dst country: <"<<buffer[i].country<<">"<<endl
                    <<"########"<<endl;
                }
            }

            if((src_info.end_int_ipaddr & netmask(buffer[i].prefix)) == buffer[i].int_ipaddr){
                if(find_end == true){
                    find_end_num++;
                    cout << find_begin_num;
                    cout<< "find more then one, end ip: "<<Int2Ip(src_info.end_int_ipaddr)<<endl
                    <<" dst cidr :"<< Int2Ip(buffer[i].int_ipaddr)<<"/"<<buffer[i].prefix<<endl;
                }
                
                find_end = true;
                if(!(src_info.isp == buffer[i].isp && src_info.country == buffer[i].country)){
                    cout<<"*******"<<endl
                    <<"is isp equal？"<<(src_info.isp == buffer[i].isp)<<endl
                    <<"is country equal？"<<(src_info.country == buffer[i].country)<<endl
                    << "error: find but isp or country not match, begin ip: "
                    <<Int2Ip(src_info.begin_int_ipaddr)<<" end ip:"<<Int2Ip(src_info.end_int_ipaddr)<<endl
                    << Int2Ip(buffer[i].int_ipaddr)<<"/"<<buffer[i].prefix<<endl
                    <<"src isp: <"<<src_info.isp<<"> dst isp: <"<<buffer[i].isp<<">"<<endl
                    <<"src country: <"<<src_info.country<<"> dst country: <"<<buffer[i].country<<">"<<endl
                    <<"########"<<endl;
                }
            }

            /* 封装进函数慢，内联函数貌似不生效 */
            // if(CheckIPInfo(src_info, buffer[i], true)){
            //     if(find_begin == true){
            //         find_begin_num++;
            //         cout << find_begin_num;
            //         cout<< "find more then one , begin ip: "<<Int2Ip(src_info.begin_int_ipaddr)<<endl
            //         <<" dst cidr :"<< Int2Ip(buffer[i].int_ipaddr)<<"/"<<buffer[i].prefix<<endl;
            //     }

            //     find_begin = true;
            // }
      
            // if(CheckIPInfo(src_info, buffer[i], false)){
            //     if(find_end == true){
            //         find_end_num++;
            //         cout << find_begin_num;
            //         cout<< "find more then one, end ip: "<<Int2Ip(src_info.end_int_ipaddr)<<endl
            //         <<" dst cidr :"<< Int2Ip(buffer[i].int_ipaddr)<<"/"<<dst_info.prefix<<endl;
            //     }
            //     find_end = true;
            // }
        }

        if(find_begin == false){
            cout<< "error: cat not find match cidr, begin ip: "<<Int2Ip(src_info.begin_int_ipaddr)<<endl;
        }

        if(find_end == false){
            cout<< "error: cat not find match cidr, end ip: "<<Int2Ip(src_info.end_int_ipaddr)<<endl;
        }

        if(find_begin_num != 0){
            cout<< "find more then one , begin ip: "<<Int2Ip(src_info.begin_int_ipaddr)<<endl;
        }

        if(find_end_num != 0){
            cout<< "find more then one, end ip: "<<Int2Ip(src_info.end_int_ipaddr)<<endl;
        }
    }
    src_file.close();
    dst_file.close();
    time(&end);
    cout << "take total time ：" << (end-start) << "s" << endl;
}