#!/usr/bin/bash

url='https://www.irishlottery.com/daily-million-archive'
agent='Mozilla/5.0 (X11; Linux x86_64; rv:93.0) Gecko/20100101 Firefox/93.0'

curl -H "User-Agent: ${agent}" "${url}" >lottery.html

date_pat='[A-Z]{1}[a-z]+ [0-9]{1,2}[a-z]{2} [0-9]+'
time_pat='[0-9]{1,2}:[0-9]{1,2}[a-z]{2}'
pup -f lottery.html 'tr th a text{}' | grep -E -o "${date_pat}" >dates.txt
pup -f lottery.html 'tr th a text{}' | grep -E -o "${time_pat}" >times.txt
pup -f lottery.html 'tr ul.balls text{}' | grep -E -o '[0-9]+' >balls.txt

paste -d ' ' dates.txt times.txt > datetimes.txt
offset=1
while read -r result_date
do
    balls=$(tail -n +$offset balls.txt | head -n 6 | tr '\n' ' ')
    offset=$(expr $offset + 6)
    zball=$(tail -n +$offset balls.txt | head -n 1)
    offset=$(expr $offset + 1)
    echo -e "${result_date}\t${balls}\tZz:${zball}"
done <datetimes.txt
