package firstpass

import (
	"google.golang.org/protobuf/compiler/protogen"
)

type MessageInfo struct {
	File *protogen.File
	Msg  *protogen.Message
}

var AllTopMessages map[string]*MessageInfo

func init() {
	AllTopMessages = make(map[string]*MessageInfo)
}

func Init(gen *protogen.Plugin) {
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, message := range f.Messages {
			AllTopMessages[string(message.Desc.FullName())] = &MessageInfo{
				File: f,
				Msg:  message,
			}
		}
	}
}
