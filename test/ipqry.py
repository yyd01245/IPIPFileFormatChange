# coding: utf-8
import sys
import csv
from IPy import IP
import time
import re

CMD_OK = 0
CMD_ERR = 1
CMD_EXIT = 2
CMD_CONT = 3

def usage():
    print("Usage:")
    print("\t<qry_ip>")
    print("")
    print("\tquit/exit to exit")

def checkip(ip):  
    p = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')  
    if p.match(ip):  
        return True  
    else:  
        return False 

def find_ip_msg(open_file, qry_ip):
    b_find = False
    time_start = time.clock()
    reader = csv.reader(open(open_file, 'rb'))
    for dst_record in reader:
        if (qry_ip in IP(dst_record[0])) == True:
            b_find = True
            print qry_ip + " founded:"
            print dst_record
            break
    if b_find == False :
        print "error: cat not find src match " + qry_ip

    time_end = time.clock()
    time_cost = (time_end - time_start)
    print "time_cost: ", time_cost, "s"


def cmd_loop(open_file):
    line = raw_input('> ')
    line = line.lower().strip()
    if not line:
        return CMD_CONT

    if line == 'quit' or line == 'exit':
        return CMD_EXIT

    args = line.split()
    qry_ip = args[0]
    if checkip(qry_ip):
        find_ip_msg(open_file, qry_ip)
    else:
        usage()

if __name__ == "__main__":
    if len(sys.argv) == 2:
        open_file = sys.argv[1]
    else:
        print("Usage:")
        print("\t<file>")
        sys.exit(0)
    try:
        while True:
            r = cmd_loop(open_file)
            if r == CMD_EXIT:
                break
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(e)
    # finally: