package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
)

func encodeDigest(buf json.RawMessage) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write(buf)
	if err != nil {
		return "", fmt.Errorf("sha256 failed: %s", err.Error())
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type ObjectSnapshot struct {
	Info        enigma.NxInfo   `json:"info,omitempty"`
	Parent      *enigma.NxInfo  `json:"parent,omitempty"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Properties  json.RawMessage `json:"properties,omitempty"`
	Digest      string          `json:"digest,omitempty"`
}

type ObjectDiff struct {
	Left  *ObjectSnapshot `json:"left,omitempty"`
	Right *ObjectSnapshot `json:"right,omitempty"`
	Merge *ObjectSnapshot `json:"merge,omitempty"`
}

func (odiff *ObjectDiff) GetInfo() (*enigma.NxInfo, *enigma.NxInfo, *util.Result) {
	if odiff.Merge != nil {
		return &odiff.Merge.Info, odiff.Merge.Parent, nil
	}
	if odiff.Left != nil {
		return &odiff.Left.Info, odiff.Left.Parent, nil
	}
	if odiff.Right != nil {
		return &odiff.Right.Info, odiff.Right.Parent, nil
	}

	return nil, nil, util.MsgError("Get diff info", "no valid object")
}

func (odiff *ObjectDiff) NeedMergeToLeft() bool {
	if odiff.Merge == nil {
		return false
	}
	if odiff.Left == nil {
		return true
	}

	mergeHash, _ := encodeDigest(odiff.Merge.Properties)
	qlikHash, _ := encodeDigest(odiff.Left.Properties)
	return mergeHash != qlikHash
}

type ListSnapshot map[string]ObjectSnapshot

type ListDiff map[string]ObjectDiff

type AppSnapshot map[string]ListSnapshot

type AppDiff map[string]ListDiff

func (app *AppDiff) MergedSnapshot() *AppSnapshot {
	merge := make(AppSnapshot)
	for list, listDiff := range *app {
		merge[list] = make(ListSnapshot)
		for objID, objDiff := range listDiff {
			merge[list][objID] = *objDiff.Merge
		}
	}

	return &merge
}

func (lhs AppSnapshot) Diff(rhs AppSnapshot) *AppDiff {
	diff := make(AppDiff)
	common := map[string]bool{}

	for list, leftList := range lhs {
		diff[list] = make(ListDiff)
		for objID, leftObj := range leftList {
			if rightList, rightListExists := rhs[list]; rightListExists {
				if rightObj, rightObjExists := rightList[objID]; rightObjExists {
					if leftObj.Digest != rightObj.Digest {
						l := leftObj
						r := rightObj
						diff[list][objID] = ObjectDiff{
							Left:  &l,
							Right: &r,
						}
					}
					common[objID] = true
					continue
				}
			}
			l := leftObj
			diff[list][objID] = ObjectDiff{
				Left: &l,
			}
		}
	}

	for f, rf := range rhs {
		for o, ro := range rf {
			if _, ok := common[o]; !ok {
				r := ro
				if _, dok := diff[f]; !dok {
					diff[f] = make(ListDiff)
				}
				diff[f][o] = ObjectDiff{
					Right: &r,
				}
			}
		}
	}

	return &diff
}

func ObjectSnapshoter(doc *enigma.Doc, listName string, item NxContainerEntry, _logger *zerolog.Logger) (*ObjectSnapshot, *util.Result) {
	_logger.Info().Msgf(" - snapshot[%s] object[%s/%s]:", listName, item.Info.Type, item.Info.Id)
	shot := ObjectSnapshot{
		Info:        *item.Info,
		Parent:      nil,
		Title:       nil,
		Description: nil,
		Properties:  nil,
		Digest:      "",
	}
	return &shot, nil
}

//func AppSnapshoter(doc *enigma.Doc)
