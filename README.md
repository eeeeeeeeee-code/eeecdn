# eeecdn

## 简介
> Eeyes不是更新了，正好有时间，加上想练练goland，便写了这个脚本，快速检测cdn，输出真实ip，并保存至xlsx中，方便查看

## 效果展示
> 默认线程5的情况下，几十个ip或者domain基本都是几秒钟就处理好了，而且所有的结果在输出以后，都会整理到一个xlsx文件当中，方便查看利用。
>

![image](https://github.com/user-attachments/assets/65102361-e654-4d90-8f58-51d975aa9eff)
![image](https://github.com/user-attachments/assets/9dd56a3f-2a0f-438e-8447-35fcc9be03c8)
![image](https://github.com/user-attachments/assets/87fe97de-bbdb-4c0e-9c66-7015081421c3)
![image](https://github.com/user-attachments/assets/2791f449-8059-4273-b8ee-7e3ce124ad83)

## 使用方法
> *编译*
>
> go build -ldflags "-w -s" ./

```
参数

正常执行
eeecdn.exe -f url.txt

不输出c段和保存的运营商
eeecdn.exe -f url.txt -c 阿里云,腾讯云
```

