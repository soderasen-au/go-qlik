package engine

import (
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/soderasen-au/go-common/util"
)

var SessionBookmarkListDef = []byte(`
{
	"qInfo": {
		"qId": "BookmarkList",
		"qType": "BookmarkList"
	},
	"qBookmarkListDef": {
		"qType": "bookmark",
		"qData": {
			"sheetId": "/sheetId",
			"selectionFields": "/selectionFields",
			"creationDate": "/creationDate"
		}
	}
}
`)

type SessionBookmark struct {
	Info            *enigma.NxInfo     `json:"qInfo,omitempty"`
	Meta            *NxMeta            `json:"qMeta,omitempty"`
	SheetId         string             `json:"sheetId,omitempty"`
	SelectionFields string             `json:"selectionFields,omitempty"`
	Bookmark        *enigma.NxBookmark `json:"qBookmark,omitempty"`
}

type BookmarkMeta struct {
	Title string `json:"title,omitempty"`
}
type BookmarkEntry struct {
	Info *enigma.NxInfo  `json:"qInfo,omitempty"`
	Meta *BookmarkMeta   `json:"qMeta,omitempty"`
	Data json.RawMessage `json:"qData,omitempty"`
}

func GetBookmarks(doc *enigma.Doc) ([]*BookmarkEntry, *util.Result) {
	opt := &enigma.NxGetBookmarkOptions{
		Types: []string{"bookmark"},
	}
	bmData, err := doc.GetBookmarksRaw(ConnCtx, opt)
	if err != nil {
		return nil, util.Error("GetBookmarks", err)
	}

	bms := make([]*BookmarkEntry, 0)
	err = json.Unmarshal(bmData, &bms)
	if err != nil {
		return nil, util.Error("ParseBookmarks", err)
	}

	return bms, nil
}

func GetSessionBookmarks(doc *enigma.Doc) ([]SessionBookmark, *util.Result) {
	var prop enigma.GenericObjectProperties
	if err := json.Unmarshal(SessionBookmarkListDef, &prop); err != nil {
		return nil, util.Error("cretae session bookmark list definition", err)
	}

	obj, err := doc.CreateSessionObject(ConnCtx, &prop)
	if err != nil {
		return nil, util.Error("cretae session bookmark list", err)
	}

	layoutBuf, err := obj.GetLayoutRaw(ConnCtx)
	if err != nil {
		return nil, util.Error("get session object", err)
	}

	var layout SessionObjectLayout
	err = json.Unmarshal(layoutBuf, &layout)
	if err != nil {
		return nil, util.Error("parse session object", err)
	}

	ret := make([]SessionBookmark, 0)
	if layout.BookmarkList != nil {
		for bi, bm := range layout.BookmarkList.Items {
			var sessionBM SessionBookmark
			if err := json.Unmarshal(bm.Data, &sessionBM); err != nil {
				return nil, util.Error(fmt.Sprintf("parse session bookmark[%d]: %s", bi, bm.Info.Id), err)
			}
			sessionBM.Info = bm.Info
			sessionBM.Meta = bm.Meta
			ret = append(ret, sessionBM)
		}
	}
	return ret, nil
}
