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
	ObjWalkResult[T any] struct {
		Info         *enigma.NxInfo      `json:"qInfo"`
		Meta         *NxMeta             `json:"qMeta"`
		Result       *T                  `json:"qProperties"`
		ChildResults []*ObjWalkResult[T] `json:"qChildInfos"`
	}
	ListWalkResult[T any] map[string]*ObjWalkResult[T] // objectId => Object Info
	AppWalkResult[T any]  map[string]ListWalkResult[T] // listName => Objects Infos of list

	ObjWalkFunc[T any]     func(doc *enigma.Doc, item NxContainerEntry, _logger *zerolog.Logger) (*ObjWalkResult[T], *util.Result)
	ObjWalkFuncEx[T any]   func(doc *enigma.Doc, info, parent *enigma.NxInfo, _logger *zerolog.Logger) (*ObjWalkResult[T], *util.Result)
	ListWalkFuncMap[T any] map[string]ObjWalkFunc[T] // listName => ObjWalkFunc

	ObjWalker[T any] struct {
		Walker ObjWalkFuncEx[T]
		Logger *zerolog.Logger
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

func NewRecurObjWalkFunc[T any](walker ObjWalkFuncEx[T]) ObjWalkFunc[T] {
	return func(doc *enigma.Doc, item NxContainerEntry, _logger *zerolog.Logger) (*ObjWalkResult[T], *util.Result) {
		return RecurWalkObject(doc, item, nil, walker, _logger)
	}
}

func (w *AppWalker[T]) Walk(doc *enigma.Doc) (AppWalkResult[T], *util.Result) {
	return WalkApp(doc, w.Walkers, w.Logger)
}

func (w *AppWalker[T]) WalkSheets(doc *enigma.Doc, walker ObjWalkFuncEx[T]) (AppWalkResult[T], *util.Result) {
	walkers := make(ListWalkFuncMap[T])
	walkers[SHEET_LIST] = NewRecurObjWalkFunc(walker)
	return WalkApp(doc, walkers, w.Logger)
}

func (w *ObjWalker[T]) Walk(doc *enigma.Doc, info *enigma.NxInfo) (*ObjWalkResult[T], *util.Result) {
	rootEntry := NxContainerEntry{
		Info: info,
		Meta: &NxMeta{},
	}
	return RecurWalkObject(doc, rootEntry, nil, w.Walker, w.Logger)
}

func WalkApp[T any](doc *enigma.Doc, walkers ListWalkFuncMap[T], _logger *zerolog.Logger) (AppWalkResult[T], *util.Result) {
	if _logger == nil {
		_logger = loggers.CoreDebugLogger
	}
	logger := _logger.With().
		Str("mod", "engine").
		Str("func", "WalkApp").
		Logger()

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
			logger.Info().Msgf("Walk object list: %s", listName)
			walker, ok := walkers[listName]
			if !ok {
				logger.Warn().Msgf(" - no walker for %s", listName)
				walker, ok = walkers[ANY_LIST]
				if !ok {
					logger.Warn().Msg("   - no default walker to use, skip this list")
					continue
				}
				logger.Warn().Msg("   - use default walker for any list")
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
					if item.Meta == nil {
						item.Meta = &NxMeta{}
					}
					objRes, res := walker(doc, *item, &logger)

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

func RecurWalkObject[T any](doc *enigma.Doc, entry NxContainerEntry, parent *enigma.NxInfo, walker ObjWalkFuncEx[T], _logger *zerolog.Logger) (*ObjWalkResult[T], *util.Result) {
	if _logger == nil {
		_logger = loggers.CoreDebugLogger
	}
	qid := entry.Info.Id
	qtype := entry.Info.Type
	logger := _logger.With().
		Str("mod", "engine").
		Str("func", "RecursiveGetSnapshots").
		Str("entry", fmt.Sprintf("%s/%s", qtype, qid)).
		Logger()

	objResult, res := walker(doc, entry.Info, parent, &logger)
	if res != nil {
		return nil, res.With(fmt.Sprintf("walker[%s/%s]", entry.Info.Type, entry.Info.Id))
	}

	logger.Trace().Msg("GetChildInfos")
	obj, err := GetObject(doc, qtype, qid)
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
			chiledEntry := NxContainerEntry{
				Info: child,
				Meta: &NxMeta{},
			}
			logger.Trace().Msgf("child[%d]: %s/%s start", i, child.Type, child.Id)
			childObjShot, res := RecurWalkObject(doc, chiledEntry, entry.Info, walker, _logger)
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
