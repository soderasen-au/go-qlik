package engine

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/loggers"
	"github.com/soderasen-au/go-common/util"
)

type (
	ObjWalkEntry struct {
		AppId  *string
		Doc    *enigma.Doc
		Item   *NxContainerEntry
		Info   *enigma.NxInfo
		Parent *enigma.NxInfo
		Logger *zerolog.Logger
	}

	ObjWalkResult[T any] struct {
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

func FlattenObject[T any](obj *ObjWalkResult[T], m ListWalkResult[T]) {
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

func (w *AppWalker[T]) Walk(doc *enigma.Doc) (AppWalkResult[T], *util.Result) {
	return WalkApp(doc, w.Walkers, w.Logger)
}

func (w *AppWalker[T]) WalkSheets(doc *enigma.Doc, walker ObjWalkFunc[T]) (AppWalkResult[T], *util.Result) {
	walkers := make(ListWalkFuncMap[T])
	walkers[SHEET_LIST] = NewRecurObjWalkFunc(walker)
	return WalkApp(doc, walkers, w.Logger)
}

func (w *ObjWalker[T]) Walk(doc *enigma.Doc, info *enigma.NxInfo) (*ObjWalkResult[T], *util.Result) {
	return RecurWalkObject(w.Entry, w.Walker)
}

func WalkApp[T any](doc *enigma.Doc, walkers ListWalkFuncMap[T], _logger *zerolog.Logger) (AppWalkResult[T], *util.Result) {
	if _logger == nil {
		_logger = loggers.CoreDebugLogger
	}
	logger := _logger.With().
		Str("mod", "engine").
		Str("func", "WalkApp").
		Logger()

	app, err := doc.GetAppLayout(ConnCtx)
	if err != nil {
		return nil, util.Error("GetAppLayout", err)
	}
	appid := app.FileName
	_logger.Trace().Msgf("layout.FileName: %s", appid)

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
		//logger.Debug().Msgf("layout field: %s", listName)
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

			var mutex = &sync.RWMutex{}
			errArray := make([]*util.Result, len(items))
			listResult := make(ListWalkResult[T])
			var wg sync.WaitGroup
			for i, item := range items {
				i := i
				item := item
				wg.Add(1)

				go func() {
					defer wg.Done()
					entry := ObjWalkEntry{
						AppId:  util.Ptr(appid),
						Doc:    doc,
						Item:   item,
						Info:   item.Info,
						Parent: nil,
						Logger: &logger,
					}

					objRes, res := walker(entry)

					mutex.Lock()
					defer mutex.Unlock()
					errArray[i] = res
					listResult[item.Info.Id] = objRes
				}()
			}
			wg.Wait()

			for i, res := range errArray {
				if res != nil {
					return nil, res.LogWith(&logger, fmt.Sprintf("%s[%d]: %s/%s", listName, i, items[i].Info.Type, items[i].Info.Id))
				}
			}
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
		Str("mod", "engine").
		Str("func", "RecursiveGetSnapshots").
		Str("entry", fmt.Sprintf("%s/%s", qtype, qid)).
		Logger()

	objResult, res := walker(e)
	if res != nil {
		return nil, res.With(fmt.Sprintf("walker[%s/%s]", e.Info.Type, e.Info.Id))
	}

	logger.Trace().Msg("GetChildInfos")
	obj, err := GetObject(e.Doc, qtype, qid)
	if err != nil {
		return nil, util.Error("can't get obj "+qid, err)
	}

	if !HasMethodOn(obj, "GetChildInfos") {
		return objResult, nil
	}
	logger.Trace().Msg("get child infos")
	ret, err := Invoke1Res1ErrOn(obj, "GetChildInfos", ConnCtx)
	if err != nil {
		return nil, util.Error("can't get obj children info "+qid, err)
	}

	childrenInfos := ret.Interface().([]*enigma.NxInfo)

	mutex := sync.RWMutex{}
	childArray := make([]*ObjWalkResult[T], len(childrenInfos))
	resArray := make([]*util.Result, len(childrenInfos))
	var wg sync.WaitGroup
	for i, child := range childrenInfos {
		i := i
		child := child
		wg.Add(1)
		go func(i int, child *enigma.NxInfo) {
			defer wg.Done()
			entry := ObjWalkEntry{
				AppId: e.AppId,
				Doc:   e.Doc,
				Item: &NxContainerEntry{
					Info: child,
					Meta: &NxMeta{},
				},
				Info:   child,
				Parent: e.Info,
				Logger: e.Logger,
			}
			logger.Trace().Msgf("child[%d]: %s/%s start", i, child.Type, child.Id)
			childObjShot, res := RecurWalkObject(entry, walker)
			logger.Trace().Msgf("child[%d]: %s/%s finished", i, child.Type, child.Id)

			mutex.Lock()
			defer mutex.Unlock()
			resArray[i] = res
			childArray[i] = childObjShot
		}(i, child)
	}
	wg.Wait()

	for i, res := range resArray {
		if res != nil {
			return nil, res.LogWith(&logger, fmt.Sprintf("child[%d]: (%s: %s)", i, childrenInfos[i].Id, childrenInfos[i].Type))
		}
	}

	objResult.ChildResults = childArray
	return objResult, nil
}
