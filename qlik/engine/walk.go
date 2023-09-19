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

type WalkListResult[T any] map[string]T                // objectId => Object Info
type WalkAppResult[T any] map[string]WalkListResult[T] // listName => Objects Infos of list

type ObjectWalker[T any] func(doc *enigma.Doc, listName string, item NxContainerEntry, _logger *zerolog.Logger) (T, *util.Result)

func WalkApp[T any](doc *enigma.Doc, walker ObjectWalker[T], _logger *zerolog.Logger) (*WalkAppResult[T], *util.Result) {
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

	appResult := make(WalkAppResult[T])

	layoutValue := reflect.ValueOf(*layout)
	layoutType := layoutValue.Type()
	for i := 0; i < layoutValue.NumField(); i++ {
		listField := layoutValue.Field(i)
		listName := layoutType.Field(i).Name
		logger.Debug().Msgf("layout field: %s", listName)
		if _, ok := ObjectListNameMap[listName]; ok {
			logger.Info().Msgf("Walk object list: %s", listName)
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

			errArray := make([]*util.Result, len(items))
			listResult := make(WalkListResult[T])
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
					objRes, res := walker(doc, listName, *item, &logger)
					errArray[i] = res
					listResult[item.Info.Id] = objRes
				}()
			}
			wg.Wait()

			for i, err := range errArray {
				if err != nil {
					return nil, res.LogWith(&logger, fmt.Sprintf("%s[%d]: %s/%s", listName, i, items[i].Info.Type, items[i].Info.Id))
				}
			}

			appResult[listName] = listResult
		}
	}

	return &appResult, nil
}
