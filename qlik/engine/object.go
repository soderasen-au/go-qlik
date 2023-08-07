package engine

import (
	"fmt"
	"github.com/qlik-oss/enigma-go/v4"
	"sync"

	"github.com/soderasen-au/go-common/util"
)

var (
	PAGE_MAX_CELLS int = 4096
)

func Paging(rect enigma.Rect) []*enigma.NxPage {
	pages := make([]*enigma.NxPage, 0)
	if rect.Height*rect.Width < PAGE_MAX_CELLS {
		pages = append(pages, &enigma.NxPage{
			Top:    rect.Top,
			Left:   rect.Left,
			Height: rect.Height,
			Width:  rect.Width,
		})
		return pages
	}

	rectRight := rect.Left + rect.Width
	rectBottom := rect.Top + rect.Height
	batchHeight := util.Min(PAGE_MAX_CELLS, rect.Height)
	batchWidth := int(PAGE_MAX_CELLS / batchHeight)

	for c0 := rect.Left; c0 < rectRight+batchWidth; c0 += batchWidth {
		for r0 := rect.Top; r0 < rectBottom+batchHeight; r0 += batchHeight {
			pages = append(pages, &enigma.NxPage{
				Top:    r0,
				Left:   c0,
				Height: batchHeight,
				Width:  batchWidth,
			})
		}
	}

	return pages
}

func GetHyperCubeData(obj *enigma.GenericObject, sz enigma.Size) ([]*enigma.NxDataPage, *util.Result) {
	rect := enigma.Rect{
		Top:    0,
		Left:   0,
		Height: sz.Cy,
		Width:  sz.Cx,
	}
	pages := Paging(rect)

	type _Result struct {
		Err      error
		DataPage *enigma.NxDataPage
	}

	errArray := make([]*_Result, len(pages))
	var wg sync.WaitGroup
	for i, page := range pages {
		i := i
		page := page
		wg.Add(1)

		go func() {
			defer wg.Done()
			_pages := make([]*enigma.NxPage, 0)
			_pages = append(_pages, page)
			_dataPages, err := obj.GetHyperCubeData(ConnCtx, "/qHyperCubeDef", _pages)
			result := &_Result{
				Err: err,
			}
			if err == nil {
				if len(_dataPages) > 0 {
					result.DataPage = _dataPages[0]
				} else {
					result.DataPage = &enigma.NxDataPage{
						Area: &rect,
					}
				}
			}
			errArray[i] = result

		}()
	}
	wg.Wait()

	dataPages := make([]*enigma.NxDataPage, 0)
	for i, result := range errArray {
		if result.Err != nil {
			return nil, util.Error(fmt.Sprintf("page[%d]", i), result.Err)
		}
		dataPages = append(dataPages, result.DataPage)
	}

	return dataPages, nil
}

func PivotPaging(rect enigma.Rect) []*enigma.NxPage {
	pages := make([]*enigma.NxPage, 0)
	if rect.Height*rect.Width < PAGE_MAX_CELLS {
		pages = append(pages, &enigma.NxPage{
			Top:    rect.Top,
			Left:   rect.Left,
			Height: rect.Height,
			Width:  rect.Width,
		})
		return pages
	}

	rectBottom := rect.Top + rect.Height
	batchHeight := PAGE_MAX_CELLS / rect.Width

	for r0 := rect.Top; r0 < rectBottom+batchHeight; r0 += batchHeight {
		pages = append(pages, &enigma.NxPage{
			Top:    r0,
			Left:   rect.Left,
			Height: batchHeight,
			Width:  rect.Width,
		})
	}

	return pages
}

func GetHyperCubePivotData(obj *enigma.GenericObject, sz enigma.Size) ([]*enigma.NxPivotPage, *util.Result) {
	rect := enigma.Rect{
		Top:    0,
		Left:   0,
		Height: sz.Cy,
		Width:  sz.Cx,
	}
	pages := PivotPaging(rect)

	type _Result struct {
		Err      error
		DataPage *enigma.NxPivotPage
	}

	errArray := make([]*_Result, len(pages))
	var wg sync.WaitGroup
	for i, page := range pages {
		i := i
		page := page
		wg.Add(1)

		go func() {
			defer wg.Done()
			_pages := make([]*enigma.NxPage, 0)
			_pages = append(_pages, page)
			_dataPages, err := obj.GetHyperCubePivotData(ConnCtx, "/qHyperCubeDef", _pages)
			result := &_Result{
				Err: err,
			}
			if err == nil {
				if len(_dataPages) > 0 {
					result.DataPage = _dataPages[0]
				} else {
					result.DataPage = &enigma.NxPivotPage{
						Area: &rect,
					}
				}
			}
			errArray[i] = result

		}()
	}
	wg.Wait()

	dataPages := make([]*enigma.NxPivotPage, 0)
	for i, result := range errArray {
		if result.Err != nil {
			return nil, util.Error(fmt.Sprintf("page[%d]", i), result.Err)
		}
		dataPages = append(dataPages, result.DataPage)
	}

	return dataPages, nil
}
