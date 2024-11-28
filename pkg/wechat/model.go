package wechat

import "encoding/xml"

const (
	MsgTypeText = "text"
)

type CommonMessage struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string   `xml:"Encrypt" json:"Encrypt"`
	ToUserName   string   `xml:"ToUserName" json:"ToUserName"`
	FromUserName string   `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime" json:"CreateTime"`
	MsgType      string   `xml:"MsgType" json:"MsgType"`
	MsgID        string   `xml:"MsgId" json:"MsgId"`
}

type TextMessage struct {
	CommonMessage
	Content string `xml:"Content" json:"Content"`
}

type ImageMessage struct {
	CommonMessage
	PicURL  string `xml:"PicUrl" json:"PicUrl"`
	MediaId string `xml:"MediaId" json:"MediaId"`
}

type EncryptMessage struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string   `xml:"Encrypt" json:"Encrypt"`
	MsgSignature string   `xml:"MsgSignature" json:"MsgSignature"`
	Timestamp    int64    `xml:"TimeStamp" json:"TimeStamp"`
	Nonce        string   `xml:"Nonce" json:"Nonce"`
}
