#!/usr/bin/env bash

#usage ./godeps
#options
#   -s 仅处理核心

# 循环加载并实时解析和处理go语言依赖
# shell变量默认值初始化工具
isShort=${1:-""}

if [ "$isShort" = "-s" ]
then
   #仅解析不要部分，不包括外部类库，但是会获取必要的
   depList="$(govendor list -no-status +local)"
else
   #使用到的类库完全获取
   depList="$(govendor list -no-status)"
fi

for repo in $depList
do
    echo "Process dependency: "$repo
    #实际下载依赖
    go get -v $repo
done

