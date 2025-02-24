package main

import (
	"bufio"
	"eeecdn/util"
	"flag"
	"fmt"
	"github.com/projectdiscovery/cdncheck"
	"github.com/xuri/excelize/v2"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// 全局变量
var client *cdncheck.Client
var cdnResult []cdn      // 保存CDN的域名
var nonCdnResult []nocdn // 保存无CDN的域名
var cipList []cip        // 保存c段

type cip struct {
	domain string
	cip    string
	isp    string
}

type cdn struct {
	cdnname string
	domain  string
	ip      string
	city    string
	isp     string
}

type nocdn struct {
	cdnname string
	domain  string
	ip      string
	city    string
	isp     string
}

// 传入的参数
var target string
var tarfile string
var savePath string
var threads int
var cancelIsp string
var escanText string
var canResult []string

// 检查CDN函数
func checkCDN(domain string) {
	domain, err := util.CheckHttp(domain)
	if err != nil {
		return
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		util.Timelog(domain+" 解析ip错误！请检查域名是否存活！", "red")
		return
	}

	uniqueIps := make(map[string]struct{})
	for _, ip := range ips {
		uniqueIps[ip.String()] = struct{}{}
	}

	for ip := range uniqueIps {
		if strings.Contains(ip, ":") {
			continue
		}
		city, isp := util.CheckQqwr(ip)
		checkWithMethods(domain, ip, city, isp)
	}
}

func checkWithMethods(domain, ip, city, isp string) {
	if found, _ := util.CheckCNAME(domain); found {
		logAndSaveCDN(domain, ip, city, isp, "CNAME检测")
	} else if found, provider, err := client.CheckCDN(net.ParseIP(ip).To4()); found && err == nil {
		logAndSaveCDN(domain, ip, city, isp, provider)
	} else if found, _ := util.CheckCIDR(ip); found {
		logAndSaveCDN(domain, ip, city, isp, "CIDR检测")
	} else if found, _ := util.CheckASN(ip); found {
		logAndSaveCDN(domain, ip, city, isp, "ASN检测")
	} else {
		logAndSaveNonCDN(domain, ip, city, isp)
	}
}

func logAndSaveCDN(domain, ip, city, isp, cdnname string) {
	util.Timelog(fmt.Sprintf("%s ==> %s 通过%s检测到cdn  地区:%s 运营商:%s", domain, ip, cdnname, city, isp), "yellow")
	cdnResult = append(cdnResult, cdn{cdnname, domain, ip, city, isp})
}

func logAndSaveNonCDN(domain, ip, city, isp string) {
	util.Timelog(fmt.Sprintf("%s ==> %s 经过检测不存在cdn！ 地区:%s 运营商:%s", domain, ip, city, isp), "green")
	nonCdnResult = append(nonCdnResult, nocdn{"无cdn", domain, ip, city, isp})
	ipcip, _ := util.IpToCIDR(ip)
	cipList = append(cipList, cip{domain, ipcip, isp})
}

func checkCip() {
	util.Timelog(strings.Repeat("-", 10)+" 开始输出整理c段 "+strings.Repeat("-", 10), "cyan")

	cipList = removeDuplicateCIPs(cipList)
	for _, ipc := range cipList {
		util.Timelog(fmt.Sprintf("domain:%s ==> c段%s  运营商:%s", ipc.domain, ipc.cip, ipc.isp), "blue")
	}
}

func removeDuplicateCIPs(cipList []cip) []cip {
	if cancelIsp != "" {
		canResult = strings.Split(cancelIsp, ",")
	}

	uniqueCips := make(map[string]cip)

	for _, c := range cipList {
		shouldSkip := false
		for _, cancelIspItem := range canResult {
			if strings.Contains(c.isp, cancelIspItem) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip {
			uniqueCips[c.cip] = c
		}
	}

	var result []cip
	for _, v := range uniqueCips {
		result = append(result, v)
	}

	return result
}

func saveToXLSX(path string) {
	if path == "" {
		return
	}

	// 创建一个新的 Excel 文件
	file := excelize.NewFile()

	// 创建两个工作表
	cdnSheet := "CDN"
	nonCdnSheet := "无CDN"
	cipSheet := "c段ip"

	// 写入工作表表头
	createSheetHeader := func(sheetName string) {
		file.SetCellValue(sheetName, "A1", "CDN Status")
		file.SetCellValue(sheetName, "B1", "Domain")
		file.SetCellValue(sheetName, "C1", "ip")
		file.SetCellValue(sheetName, "D1", "地址")
		file.SetCellValue(sheetName, "E1", "运营商")
	}

	cipindex, _ := file.NewSheet(cipSheet)
	cipSheetHeader := func(sheetName string) {
		file.SetCellValue(sheetName, "A1", "Domain")
		file.SetCellValue(sheetName, "B1", "c段ip")
		file.SetCellValue(sheetName, "C1", "运营商")
	}
	cipSheetHeader(cipSheet)

	row := 2
	for _, ipc := range cipList {
		file.SetCellValue(cipSheet, fmt.Sprintf("A%d", row), ipc.domain)
		file.SetCellValue(cipSheet, fmt.Sprintf("B%d", row), ipc.cip)
		file.SetCellValue(cipSheet, fmt.Sprintf("C%d", row), ipc.isp)
		row++
	}

	// 写入 CDN 域名与 IP 数据
	file.NewSheet(cdnSheet)
	createSheetHeader(cdnSheet)

	row = 2
	for _, cdn := range cdnResult {
		file.SetCellValue(cdnSheet, fmt.Sprintf("A%d", row), cdn.cdnname)
		file.SetCellValue(cdnSheet, fmt.Sprintf("B%d", row), cdn.domain)
		file.SetCellValue(cdnSheet, fmt.Sprintf("C%d", row), cdn.ip)
		file.SetCellValue(cdnSheet, fmt.Sprintf("D%d", row), cdn.city)
		file.SetCellValue(cdnSheet, fmt.Sprintf("E%d", row), cdn.isp)
		row++
	}

	// 写入 Non-CDN 域名与 IP 数据
	file.NewSheet(nonCdnSheet)
	createSheetHeader(nonCdnSheet)

	row = 2
	for _, cdn := range nonCdnResult {
		file.SetCellValue(nonCdnSheet, fmt.Sprintf("A%d", row), cdn.cdnname)
		file.SetCellValue(nonCdnSheet, fmt.Sprintf("B%d", row), cdn.domain)
		file.SetCellValue(nonCdnSheet, fmt.Sprintf("C%d", row), cdn.ip)
		file.SetCellValue(nonCdnSheet, fmt.Sprintf("D%d", row), cdn.city)
		file.SetCellValue(nonCdnSheet, fmt.Sprintf("E%d", row), cdn.isp)
		row++
	}

	// 设置活动工作表，选择 CDN 表
	file.SetActiveSheet(cipindex) // 你可以根据需要设置默认活动工作表为 CDN
	file.DeleteSheet("Sheet1")

	// 保存文件
	if err := file.SaveAs(path); err != nil {
		util.Timelog("无法保存XLSX文件", "red")
	}
}

func theadstart() {
	client = cdncheck.New()
	// 文件读取
	file, err := os.Open(tarfile)
	defer file.Close()
	if err != nil {
		log.Panicln(err)
		return
	}

	// 使用 sync.WaitGroup 来控制并发
	sem := make(chan struct{}, threads)

	var wg sync.WaitGroup

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		sem <- struct{}{} // 获取信号量，阻塞直到有空间
		wg.Add(1)
		go func(line string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量
			checkCDN(line)
		}(line)
	}

	wg.Wait()
}

func flagInit() {
	flag.StringVar(&tarfile, "f", "", "需要扫描的文件")
	flag.StringVar(&target, "u", "", "需要扫描单个域名")
	flag.StringVar(&savePath, "o", "cdn.xlsx", "保存结果文件路径，必须为 .xlsx 格式, 默认保存至cdn.xlsx")
	flag.StringVar(&cancelIsp, "c", "", "指定不输出c段的运营商,如-c 阿里云,华为云")
	flag.IntVar(&threads, "thread", 5, "并发数，默认5")
	flag.Parse()
}

func checkinit() {

	if !strings.HasSuffix(savePath, ".xlsx") {
		util.Timelog("-o请指定xlsx后缀", "red")
		flag.Usage()
		os.Exit(1)
	}

	if tarfile != "" {
		util.Timelog(strings.Repeat("-", 10)+" 开始进行扫描 "+strings.Repeat("-", 10), "cyan")
		theadstart()
		checkCip()
	} else if target != "" {
		util.Timelog(strings.Repeat("-", 10)+" 开始进行扫描 "+strings.Repeat("-", 10), "cyan")
		checkCDN(target)
		checkCip()
	} else {
		util.Timelog("-f或-u未指定，请指定后再进行执行", "red")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	// 参数初始化
	flagInit()
	escanText = " ----------------------------------\n|                         _        |\n|   ___  ___  ___  ___ __| |_ __   |\n|  / _ \\/ _ \\/ _ \\/ __/ _` | '_ \\  |\n| |  __/  __/  __/ (_| (_| | | | | |\n|  \\___|\\___|\\___|\\___\\__,_|_| |_| |\n|    -- by:额呃e  版本：1.0 --     |\n|  --  深潜sec安全团队公众号 --\t   |\n|github.com/eeeeeeeeee-code/eeecdn |\n ----------------------------------\n"
	util.PrintColor(escanText, "magenta")
	checkinit()

	// 保存结果到XLSX文件
	if savePath != "" {
		util.Timelog(strings.Repeat("-", 10)+" 开始进行保存 "+strings.Repeat("-", 10), "cyan")
		saveToXLSX(savePath)
		util.Timelog("结果成功保存至"+savePath+"~~~~~", "magenta")
	} else {
		util.Timelog("请提供保存路径", "red")
	}
}
