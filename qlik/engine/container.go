package engine

import "github.com/qlik-oss/enigma-go/v4"

type ContainerChildInfo struct {
	ContainerChildId string `json:"containerChildId,omitempty"`
	QExtendsId       string `json:"qExtendsId,omitempty"`
	ShowCondition    string `json:"showCondition,omitempty"`
	Title            string `json:"title,omitempty"`
	Visualization    string `json:"visualization,omitempty"`
}

type ContainerChildItem struct {
	Entry *enigma.NxContainerEntry
	Info  *ContainerChildInfo
}
