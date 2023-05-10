#!/bin/bash

# 循环1000次
for i in {1..2}
do
  # 执行go代码并测量执行时间
  time go run main.go

  # 打印循环结果
  echo "Test $i: Done"
done