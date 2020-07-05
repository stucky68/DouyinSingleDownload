package main

type Data struct {
	AwemeList  []Item `json:"item_list"`
}

type Item struct {
	Desc         string      `json:"desc"`
	Video        video       `json:"video"`
}

type video struct {
	PlayAddr     uriStr      `json:"play_addr"`
	Vid          string      `json:"vid"`
}

type uriStr struct {
	Uri     string   `json:"uri"`
	UrlList []string `json:"url_list"`
	width   int      `json:"width"`
	height  int      `json:"height"`
}
