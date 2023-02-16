package main

//go:generate odin name --file $GOFILE

type TestBase struct {
	Index string `json:"index"`
	Type  string `json:"type"`
}

type TestEmbed struct {
	TestBase
	Fields []Field `json:"fields"`
}
