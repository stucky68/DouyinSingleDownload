package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func GetData(url string) (itemID, dytk string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/" + s  + ".36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "_ga=GA1.2.685263550.1587277283; _gid=GA1.2.143250871.1587911549; tt_webid=6820028204934923790; _ba=BA0.2-20200301-5199e-c7q9NP0laGm7KfaPfGcH")
	res, err := client.Do(req)
	if err == nil {
		b, _ := ioutil.ReadAll(res.Body)
		result := string(b)
		var itemIDRegexp = regexp.MustCompile(`itemId: "(.*?)"`)
		ids := itemIDRegexp.FindStringSubmatch(result)
		if len(ids) > 1 {
			itemID = ids[1]
		}
		var dytkRegexp = regexp.MustCompile(`dytk: "(.*?)"`)
		dytks := dytkRegexp.FindStringSubmatch(result)
		if len(dytks) > 1 {
			dytk = dytks[1]
		}
	}
	return
}

func read3(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func Download(getUrl, saveFile string) error {
	client := &http.Client{}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/"+s+".1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/"+s+".1")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("pragma", "no-cache")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.Header.Get("content-length") == "0" {
		return errors.New("ip")
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFile, b, 0666)
	if err != nil {
		return err
	}
	return nil
}

func downloadHttpFile(videoUrl string, localVideo string) error {
	timeout := 0
	for {
		err := Download(videoUrl, localVideo)
		if err == nil {
			break
		}
		log.Println(err)
		timeout++
		if timeout > 3 {
			return errors.New("超时三次失败")
		}
	}
	return nil
}

func FilterEmoji(content string) string {
	newContent := ""
	for _, value := range content {
		if unicode.Is(unicode.Han, value) || unicode.IsLetter(value) || unicode.IsDigit(value) || unicode.IsSpace(value) {
			newContent += string(value)
		}
	}
	return newContent
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func HandleJson(data Data) {
	for _, item := range data.AwemeList {
		//err, _ := getXiGuaVideoUrl(item.Video.Vid)
		//if err != nil {
		//	fmt.Println(err, "1234")
		//	continue
		//}

		item.Desc = strings.ReplaceAll(item.Desc, ":", "")
		item.Desc = strings.ReplaceAll(item.Desc, "?", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\\", "")
		item.Desc = strings.ReplaceAll(item.Desc, "/", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\"", "")
		item.Desc = strings.ReplaceAll(item.Desc, "*", "")
		item.Desc = strings.ReplaceAll(item.Desc, "<", "")
		item.Desc = strings.ReplaceAll(item.Desc, ">", "")
		item.Desc = strings.ReplaceAll(item.Desc, "|", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\r", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\n", "")
		item.Desc = FilterEmoji(item.Desc)

		localVideo := "download/" + item.Desc + ".mp4"

		if IsExist(localVideo) == false {
			log.Println("开始处理数据:", item.Desc)
			//fmt.Println(item.Video.PlayAddr.UrlList[0])
			err := downloadHttpFile("https://aweme.snssdk.com/aweme/v1/play/?video_id="+item.Video.Vid+"&media_type=4&vr_type=0&improve_bitrate=0&is_play_url=1&is_support_h265=0&source=PackSourceEnum_PUBLISH", localVideo)
			//err := downloadHttpFile(item.Video.PlayAddr.UrlList[0], localVideo)
			if err != nil {
				log.Println("下载视频失败:", err)
				continue
			} else {
				log.Println("下载视频成功:", localVideo)
			}
		} else {
			log.Println(item.Desc + " " + localVideo + "文件已存在，跳过")
		}
	}
}

func GetVideo(itemID string) (error, Data) {
	client := &http.Client{}
	url := "https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=" + itemID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, Data{}
	}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))

	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	//req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("cookie", "_ga=GA1.2.938284732.1578806304; _gid=GA1.2.1428838914.1578806305")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/" + s + ".36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	req.Header.Add("Host", "www.iesdouyin.com")
	res, err := client.Do(req)
	if err != nil {
		return err, Data{}
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, Data{}
	}
	var data Data
	err = json.Unmarshal(b, &data)
	if err != nil {
		//fmt.Println(err, string(b))
		return err, Data{}
	}
	return nil, data
}


func main()  {
	user := read3("./video.txt")
	os.MkdirAll("download/", os.ModePerm)
	videos := strings.Split(user, "\r\n")
	for _, v := range videos {
		reg := regexp.MustCompile(`/[1-9]\d*/?`)
		itemID := reg.FindString(v)
		itemID = strings.ReplaceAll(itemID, "/", "")

		if itemID != "" {
			err, d := GetVideo(itemID)
			if err == nil {
				HandleJson(d)
			}
		} else {
			fmt.Println(v + " 获取数据失败")
		}
	}
}
