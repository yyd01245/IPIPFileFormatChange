# IPIPFileFormatChange
change ipip db file to custom file format

## Bulid
g++ -o ipipfileformatchange ipipfileformatchange.cc

## Usage
usage: cidr [-i] [input file] [-o] [output file]
eg. :./ipipfileformatchange -i ipip.csv -o out.csv

## Change rule
Get the 1st, 2nd, 7th and 15th columns of the source file
and  make up the new file.

change the format rule by yourself