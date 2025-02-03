package engine

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
	"golang.org/x/sync/semaphore"
)

type (
	FilterOptions struct {
		SheetBlackList  []string `json:"sheet_black_list,omitempty" yaml:"sheet_black_list"`
		ObjectBlackList []string `json:"object_black_list,omitempty" yaml:"object_black_list"`
		ObjectWhiteList []string `json:"object_white_list,omitempty" yaml:"object_white_list"`
		SheetWhiteList  []string `json:"sheet_white_list,omitempty" yaml:"sheet_white_list"`

		sheetBlackMap  map[string]bool
		objectBlackMap map[string]bool
		objectWhiteMap map[string]bool
		sheetWhiteMap  map[string]bool
	}

	WalkOptions struct {
		Sync           bool           `json:"sync" yaml:"sync"`
		MaxWorkers     int            `json:"max_workers" yaml:"max_workers"`
		IgnoreError    bool           `json:"ignore_error" yaml:"ignore_error"`
		OpendocRetries int            `json:"opendoc_retries" yaml:"opendoc_retries"`
		RetryDelay     int            `json:"retry_delay" yaml:"retry_delay"`
		Filter         *FilterOptions `json:"filter" yaml:"filter"`
	}

	ObjWalkEntry struct {
		*WalkOptions

		Config    *MixedConfig
		AppId     *string
		Doc       *enigma.Doc
		Item      *NxContainerEntry
		SheetId   string
		SheetName string
		Info      *enigma.NxInfo
		Parent    *enigma.NxInfo
		Logger    *zerolog.Logger
	}

	ObjWalkResult[T any] struct {
		SheetId      string              `json:"sheet_id" yaml:"sheet_id"`
		SheetName    string              `json:"sheet_name" yaml:"sheet_name"`
		ObjectTitle  *string             `json:"object_title" yaml:"object_title"`
		Info         *enigma.NxInfo      `json:"info,omitempty" yaml:"info,omitempty"`
		Parent       *enigma.NxInfo      `json:"parent,omitempty" yaml:"parent,omitempty"`
		Meta         *NxMeta             `json:"meta,omitempty" yaml:"meta,omitempty"`
		Result       *T                  `json:"result,omitempty" yaml:"result,omitempty"`
		ChildResults []*ObjWalkResult[T] `json:"child_results,omitempty" yaml:"child_results,omitempty"`
	}

	ListWalkResult[T any] map[string]*ObjWalkResult[T] // objectId => Object Info
	AppWalkResult[T any]  map[string]ListWalkResult[T] // listName => Objects Infos of list

	ObjWalkFunc[T any]     func(e ObjWalkEntry) (*ObjWalkResult[T], *util.Result)
	ListWalkFuncMap[T any] map[string]ObjWalkFunc[T] // listName => ObjWalkFunc

	ObjWalker[T any] struct {
		Walker ObjWalkFunc[T]
		Entry  ObjWalkEntry
	}

	AppWalker[T any] struct {
		Walkers ListWalkFuncMap[T]
		Logger  *zerolog.Logger
	}
)

func (f *FilterOptions) BuildMap() {
	if f.SheetBlackList != nil {
		f.sheetBlackMap = make(map[string]bool)
		for _, sheet := range f.SheetBlackList {
			f.sheetBlackMap[sheet] = true
		}
	}

	if f.ObjectBlackList != nil {
		f.objectBlackMap = make(map[string]bool)
		for _, object := range f.ObjectBlackList {
			f.objectBlackMap[object] = true
		}
	}

	if f.ObjectWhiteList != nil {
		f.objectWhiteMap = make(map[string]bool)
		for _, object := range f.ObjectWhiteList {
			f.objectWhiteMap[object] = true
		}
	}

	if f.SheetWhiteList != nil {
		f.sheetWhiteMap = make(map[string]bool)
		for _, sheet := range f.SheetWhiteList {
			f.sheetWhiteMap[sheet] = true
		}
	}
}

func (f *FilterOptions) IsSheetBlackListed(sheet string) bool {
	if f.SheetBlackList == nil || f.sheetBlackMap == nil {
		return false
	}
	_, ok := f.sheetBlackMap[sheet]
	return ok
}

func (f *FilterOptions) IsObjectBlackListed(object string) bool {
	if f.ObjectBlackList == nil || f.objectBlackMap == nil {
		return false
	}
	_, ok := f.objectBlackMap[object]
	return ok
}

func (f *FilterOptions) IsObjectWhiteListed(object string) bool {
	if f.ObjectWhiteList == nil || f.objectWhiteMap == nil {
		return false
	}
	_, ok := f.objectWhiteMap[object]
	return ok
}

func (f *FilterOptions) IsSheetWhiteListed(sheet string) bool {
	if f.SheetWhiteList == nil || f.sheetWhiteMap == nil {
		return false
	}
	_, ok := f.sheetWhiteMap[sheet]
	return ok
}

func DefaultWalkOptions() *WalkOptions {
	return &WalkOptions{
		Sync:           false,
		MaxWorkers:     runtime.NumCPU(),
		IgnoreError:    false,
		OpendocRetries: 3,
		RetryDelay:     30,
		Filter:         &FilterOptions{},
	}
}

func WalkOptionsForHuge() *WalkOptions {
	return &WalkOptions{
		Sync:           true,
		MaxWorkers:     2,
		IgnoreError:    true,
		OpendocRetries: 5,
		RetryDelay:     60,
		Filter:         &FilterOptions{},
	}
}

func FlattenObject[T any](obj *ObjWalkResult[T], m ListWalkResult[T]) {
	if obj == nil || obj.Info == nil {
		return
	}

	m[obj.Info.Id] = obj
	if obj.ChildResults != nil {
		for _, c := range obj.ChildResults {
			FlattenObject(c, m)
		}
	}
}

func FlattenList[T any](list ListWalkResult[T]) ListWalkResult[T] {
	newMap := make(ListWalkResult[T])
	for _, obj := range list {
		FlattenObject(obj, newMap)
	}
	return newMap
}

func NewRecurObjWalkFunc[T any](walker ObjWalkFunc[T]) ObjWalkFunc[T] {
	return func(e ObjWalkEntry) (*ObjWalkResult[T], *util.Result) {
		return RecurWalkObject(e, walker)
	}
}

func NewRecurObjWalkFuncSync[T any](walker ObjWalkFunc[T]) ObjWalkFunc[T] {
	return func(e ObjWalkEntry) (*ObjWalkResult[T], *util.Result) {
		return RecurWalkObjectSync(e, walker)
	}
}

func (w *AppWalker[T]) Walk(doc *enigma.Doc, cfg MixedConfig, opts *WalkOptions) (AppWalkResult[T], *util.Result) {
	return WalkApp(doc, cfg, opts, w.Walkers, w.Logger)
}

func (w *AppWalker[T]) WalkSheets(doc *enigma.Doc, cfg MixedConfig, opts *WalkOptions, walker ObjWalkFunc[T]) (AppWalkResult[T], *util.Result) {
	walkers := make(ListWalkFuncMap[T])
	walkers[SHEET_LIST] = NewRecurObjWalkFunc(walker)
	return WalkApp(doc, cfg, opts, walkers, w.Logger)
}

func (w *ObjWalker[T]) Walk(doc *enigma.Doc, info *enigma.NxInfo) (*ObjWalkResult[T], *util.Result) {
	return RecurWalkObject(w.Entry, w.Walker)
}

func WalkApp[T any](doc *enigma.Doc, cfg MixedConfig, opts *WalkOptions, walkers ListWalkFuncMap[T], _logger *zerolog.Logger) (AppWalkResult[T], *util.Result) {
	if _logger == nil {
		_logger = loggers.CoreDebugLogger
	}
	if opts == nil {
		opts = DefaultWalkOptions()
	}

	app, err := doc.GetAppLayout(ConnCtx)
	if err != nil {
		return nil, util.Error("GetAppLayout", err)
	}
	appid := app.FileName

	logger := _logger.With().
		Str("mod", "engine").
		Str("func", "WalkApp").
		Str("appid", appid).
		Logger()
	logger.Info().Msg("start")

	layout, res := GetSessionObjectLayout(doc)
	if res != nil {
		logger.Err(res).Msg("GetSessionObjectLayout")
		return nil, res.With("GetSessionObjectLayout")
	}

	appResult := make(AppWalkResult[T])

	layoutValue := reflect.ValueOf(*layout)
	layoutType := layoutValue.Type()
	for i := 0; i < layoutValue.NumField(); i++ {
		listField := layoutValue.Field(i)
		listName := layoutType.Field(i).Name
		if _, ok := ObjectListNameMap[listName]; ok {
			logger.Debug().Msgf("Walk object list: %s", listName)
			walker, ok := walkers[listName]
			if !ok {
				logger.Debug().Msgf(" - no walker for %s", listName)
				walker, ok = walkers[ANY_LIST]
				if !ok {
					logger.Debug().Msg("   - no default walker to use, skip this list")
					continue
				}
				logger.Debug().Msg("   - use default walker for any list")
			}
			if listField.Kind() != reflect.Ptr {
				return nil, util.LogMsgError(&logger,
					"TranslateAppObjectList["+listName+"]", " - object list field is not a pointer")
			}
			if listField.IsNil() {
				return nil, util.LogMsgError(&logger,
					"TranslateAppObjectList["+listName+"]", " - object list field is Nil")
			}
			itemsOfList := reflect.Indirect(listField).FieldByName("Items")
			if itemsOfList.IsZero() {
				return nil, util.LogMsgError(&logger,
					"TranslateAppObjectList["+listName+"]", " - object list has no `Items` field")
			}
			items, ok := itemsOfList.Interface().([]*NxContainerEntry)
			if !ok {
				return nil, util.LogMsgError(&logger,
					"TranslateAppObjectList["+listName+"]", " - `Items` field is not `[]*NxContainerEntry`")
			}
			logger.Info().Msgf("%s has %d items", listName, len(items))

			errArray := make([]*util.Result, len(items))
			listResult := make(ListWalkResult[T])

			var sheetId, sheetName string
			for i, item := range items {
				ilog := logger.With().
					Str("list", listName).
					Str("itemNo", fmt.Sprintf("%d/%d", i, len(items))).
					Str("objType", item.Info.Type).
					Str("objId", item.Info.Id).
					Logger()
				ilog.Info().Msg("walking item ...")

				if item.Info.Type == "sheet" {
					sheetObj, err := doc.GetObject(ConnCtx, item.Info.Id)
					if err != nil {
						ilog.Error().Msgf("GetSheetObject error: %s", err.Error())
						return nil, util.LogError(&logger, "GetSheetObject", err)
					}
					properties, err := sheetObj.GetPropertiesRaw(ConnCtx)
					if err != nil {
						ilog.Error().Msgf("GetPropertiesRaw error: %s", err.Error())
						return nil, util.LogError(&logger, "GetPropertiesRaw", err)
					}
					prop := ObjectPropeties{
						Info:       item.Info,
						Properties: properties,
					}
					title, _ := GetTitle(nil, &prop, &ilog)
					sheetId = item.Info.Id
					sheetName = util.MaybeNil(title)
				}

				entry := ObjWalkEntry{
					Config:      &cfg,
					WalkOptions: opts,
					AppId:       util.Ptr(appid),
					Doc:         doc,
					Item:        item,
					Info:        item.Info,
					SheetId:     sheetId,
					SheetName:   sheetName,
					Parent:      nil,
					Logger:      &ilog,
				}

				objRes, res := walker(entry)
				for reties := 0; res != nil && reties < opts.OpendocRetries; reties++ {
					relog := ilog.With().Int("reties", reties).Logger()
					relog.Warn().Msgf("result: %s", res.Error())

					if doc != nil {
						doc.DisconnectFromServer()
					}

					relog.Warn().Msgf("Try to reconnect to the doc after %ds", opts.RetryDelay)
					time.Sleep(time.Duration(opts.RetryDelay) * time.Second)
					conn, res := cfg.Connect()
					if res != nil {
						return nil, res.LogWith(&relog, "MixedConfig.Connect")
					}

					var ver *enigma.NxEngineVersion
					err = nil
					ver, err = conn.Global.EngineVersion(ConnCtx)
					if err != nil {
						relog.Info().Msgf("engine version error: %s", err.Error())
					}
					if ver != nil {
						relog.Info().Msgf("engine version: %s", ver)
					}

					relog.Info().Msgf("Opening app: %s", cfg.AppId)
					err = nil
					doc = nil
					doc, err = conn.Global.OpenDoc(ConnCtx, cfg.AppId, "", "", "", false)
					if err != nil {
						relog.Warn().Msgf("1st open doc error: %s, will reopen after 30 seconds", err.Error())

						conn.Global.DisconnectFromServer()
						time.Sleep(time.Duration(opts.RetryDelay) * time.Second)
						doc, err = conn.Global.OpenDoc(ConnCtx, cfg.AppId, "", "", "", false)
						if err != nil {
							relog.Info().Msgf("2nd open doc error: %s", err.Error())
							continue
						}
						//return nil, util.LogError(&logger, "OpenDoc", err)
					}

					entry := ObjWalkEntry{
						Config:      &cfg,
						WalkOptions: opts,
						AppId:       util.Ptr(appid),
						Doc:         doc,
						Item:        item,
						Info:        item.Info,
						SheetId:     sheetId,
						SheetName:   sheetName,
						Parent:      nil,
						Logger:      &relog,
					}

					objRes, res = walker(entry)
				}

				//if res != nil {
				//	return nil, res.LogWith(&ilog, "failed after 3 retries")
				//}

				errArray[i] = res
				listResult[item.Info.Id] = objRes
			}

			//for i, res := range errArray {
			//	if res != nil {
			//		return nil, res.LogWith(&logger, fmt.Sprintf("%s[%d]: %s/%s", listName, i, items[i].Info.Type, items[i].Info.Id))
			//	}
			//}
			appResult[listName] = listResult
		}
	}

	return appResult, nil
}

func RecurWalkObject[T any](e ObjWalkEntry, walker ObjWalkFunc[T]) (*ObjWalkResult[T], *util.Result) {
	if e.Logger == nil {
		e.Logger = loggers.CoreDebugLogger
	}
	qid := e.Item.Info.Id
	qtype := e.Item.Info.Type
	logger := e.Logger.With().
		Str(qid+"-walk", "RecurWalkObject").
		Logger()

	if e.Filter != nil && strings.ToLower(qtype) != "sheet" {
		if e.Filter.ObjectBlackList != nil {
			logger.Info().Msgf("filtering with black list, object id: %s", qid)
			if e.Filter.IsObjectBlackListed(qid) {
				logger.Info().Msgf(" - object is black listed, ignore it")
				return nil, nil
			}
		}

		if e.Filter.ObjectWhiteList != nil {
			logger.Info().Msgf("filtering with white-list, object id: %s", qid)
			if !e.Filter.IsObjectWhiteListed(qid) {
				logger.Info().Msgf(" - object is NOT white-listed")
				if !e.Filter.IsSheetWhiteListed(e.SheetId) {
					logger.Info().Msgf("   - object's sheet is NOT white-listed either, ignore it.")
					return nil, nil
				} else {
					logger.Info().Msgf("   - object's sheet is white-listed")
				}
			}
		}

		if e.Filter.IsSheetBlackListed(e.SheetId) {
			logger.Info().Msgf(" - sheet: %s is black listed, ignore it", e.SheetId)
			return nil, nil
		}

		if e.Filter.SheetWhiteList != nil {
			logger.Info().Msgf("filtering with white-list, sheet: %s", e.SheetId)
			if !e.Filter.IsSheetWhiteListed(e.SheetId) {
				logger.Info().Msgf("   - object's sheet is NOT white-listed")
				return nil, nil
			}
		}
	}

	emptyObjShot := ObjWalkResult[T]{
		Info:         e.Info,
		Result:       new(T),
		ChildResults: make([]*ObjWalkResult[T], 0),
	}

	objResult, res := walker(e)
	if res != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "walker", res.Error())
			return &emptyObjShot, nil
		}
		return nil, res.With(fmt.Sprintf("walker[%s/%s]", e.Info.Type, e.Info.Id))
	}

	logger.Trace().Msg("GetChildInfos")
	obj, err := GetObject(e.Doc, qtype, qid)
	if err != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "GetObject", err)
			return &emptyObjShot, nil
		}
		return nil, util.Error("can't get obj "+qid, err)
	}

	if !HasMethodOn(obj, "GetChildInfos") {
		return objResult, nil
	}
	logger.Trace().Msg("get child infos")
	ret, err := Invoke1Res1ErrOn(obj, "GetChildInfos", ConnCtx)
	if err != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "GetChildInfos", err)
			return &emptyObjShot, nil
		}
		return nil, util.Error("can't get obj children info "+qid, err)
	}
	childrenInfos := ret.Interface().([]*enigma.NxInfo)
	logger.Info().Msgf("get %d children", len(childrenInfos))

	var (
		semaphores = semaphore.NewWeighted(int64(e.MaxWorkers))
	)
	mutex := sync.RWMutex{}
	childArray := make([]*ObjWalkResult[T], len(childrenInfos))
	resArray := make([]*util.Result, len(childrenInfos))

	for i, child := range childrenInfos {
		clog := logger.With().Str(qid+"-child", fmt.Sprintf("%d/%d", i, len(childrenInfos))).Logger()
		i := i
		child := child

		if err := semaphores.Acquire(ConnCtx, 1); err != nil {
			clog.Error().Msgf("semaphore acquire error: %v", err)
			break
		}
		go func(i int, child *enigma.NxInfo) {
			defer semaphores.Release(1)
			entry := ObjWalkEntry{
				Config:      e.Config,
				WalkOptions: e.WalkOptions,
				AppId:       e.AppId,
				Doc:         e.Doc,
				Item: &NxContainerEntry{
					Info: child,
					Meta: &NxMeta{},
				},
				Info:      child,
				Parent:    e.Info,
				SheetId:   e.SheetId,
				SheetName: e.SheetName,
				Logger:    &clog,
			}
			clog.Trace().Msg("start")
			childObjShot, res := RecurWalkObject[T](entry, walker)
			clog.Trace().Msg("end")

			mutex.Lock()
			defer mutex.Unlock()
			resArray[i] = res
			childArray[i] = childObjShot
		}(i, child)
	}

	if err := semaphores.Acquire(ConnCtx, int64(e.MaxWorkers)); err != nil {
		logger.Error().Msgf("last semaphore acquire error: %v", err)
		if !e.IgnoreError {
			return nil, util.Error("can't acquire last semaphore", err)
		}
	}

	for i, res := range resArray {
		if res != nil && !e.IgnoreError {
			return nil, res.LogWith(&logger, fmt.Sprintf("child[%d]: (%s: %s)", i, childrenInfos[i].Id, childrenInfos[i].Type))
		}
	}

	objResult.ChildResults = childArray
	return objResult, nil
}

func RecurWalkObjectSync[T any](e ObjWalkEntry, walker ObjWalkFunc[T]) (*ObjWalkResult[T], *util.Result) {
	if e.Logger == nil {
		e.Logger = loggers.CoreDebugLogger
	}
	qid := e.Item.Info.Id
	qtype := e.Item.Info.Type
	logger := e.Logger.With().
		Str("mod", "engine").
		Str("func", "RecursiveGetSnapshots").
		Str("entry", fmt.Sprintf("%s/%s", qtype, qid)).
		Logger()

	if e.Filter != nil && strings.ToLower(qtype) != "sheet" {
		if e.Filter.ObjectBlackList != nil {
			logger.Info().Msgf("filtering with black list, object id: %s", qid)
			if e.Filter.IsObjectBlackListed(qid) {
				logger.Info().Msgf(" - object is black listed, ignore it")
				return nil, nil
			}
		}

		if e.Filter.ObjectWhiteList != nil {
			logger.Info().Msgf("filtering with white-list, object id: %s", qid)
			if !e.Filter.IsObjectWhiteListed(qid) {
				logger.Info().Msgf(" - object is NOT white-listed")
				if !e.Filter.IsSheetWhiteListed(e.SheetId) {
					logger.Info().Msgf("   - object's sheet is NOT white-listed either, ignore it.")
					return nil, nil
				} else {
					logger.Info().Msgf("   - object's sheet is white-listed")
				}
			}
		}

		if e.Filter.IsSheetBlackListed(e.SheetId) {
			logger.Info().Msgf(" - sheet: %s is black listed, ignore it", e.SheetId)
			return nil, nil
		}
	}

	emptyObjShot := ObjWalkResult[T]{
		Info:         e.Info,
		Result:       new(T),
		ChildResults: make([]*ObjWalkResult[T], 0),
	}

	objResult, res := walker(e)
	if res != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "walker", res.Error())
			return &emptyObjShot, nil
		}
		return nil, res.With(fmt.Sprintf("walker[%s/%s]", e.Info.Type, e.Info.Id))
	}

	logger.Trace().Msg("GetChildInfos")
	obj, err := GetObject(e.Doc, qtype, qid)
	if err != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "engine.GetObject", err)
			return &emptyObjShot, nil
		}
		return nil, util.Error("can't get obj "+qid, err)
	}

	if !HasMethodOn(obj, "GetChildInfos") {
		return objResult, nil
	}
	logger.Trace().Msg("get child infos")
	ret, err := Invoke1Res1ErrOn(obj, "GetChildInfos", ConnCtx)
	if err != nil {
		if e.IgnoreError {
			logger.Warn().Msgf("%s: ignored error: %v ", "engine.GetChildInfos", err)
			return &emptyObjShot, nil
		}
		return nil, util.Error("can't get obj children info "+qid, err)
	}

	childrenInfos := ret.Interface().([]*enigma.NxInfo)

	childArray := make([]*ObjWalkResult[T], len(childrenInfos))

	for i, child := range childrenInfos {
		entry := ObjWalkEntry{
			Config:      e.Config,
			WalkOptions: e.WalkOptions,
			AppId:       e.AppId,
			Doc:         e.Doc,
			Item: &NxContainerEntry{
				Info: child,
				Meta: &NxMeta{},
			},
			Info:      child,
			Parent:    e.Info,
			SheetId:   e.SheetId,
			SheetName: e.SheetName,
			Logger:    e.Logger,
		}
		logger.Trace().Msgf("child[%d]: %s/%s start", i, child.Type, child.Id)
		childObjShot, res := RecurWalkObjectSync[T](entry, walker)
		logger.Trace().Msgf("child[%d]: %s/%s finished", i, child.Type, child.Id)
		if res != nil {
			return nil, res.With(fmt.Sprintf("walker[%s/%s]", e.Info.Type, e.Info.Id))
		}
		childArray[i] = childObjShot

	}

	objResult.ChildResults = childArray
	return objResult, nil
}
